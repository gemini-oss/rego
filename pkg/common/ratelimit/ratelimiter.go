// pkg/common/ratelimit/ratelimit.go
package ratelimit

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
)

// RateLimiter struct defines the fields for the rate limiter
type RateLimiter struct {
	stopChan       chan struct{} // Channel to stop the rate limiter
	mu             sync.Mutex    // Mutex to lock the rate limiter
	Available      int           // Available requests remaining
	Limit          int           // Total requests allowed in the interval
	Interval       time.Duration // Interval to reset the rate limiter
	Requests       int           // Total requests made
	ResetTimestamp int64         // Timestamp to reset the rate limiter
	RetryAfter     int           // Retry after time
	TimeUntilReset time.Duration // Time until the rate limiter resets
	UsesReset      bool          // Flag to check if the rate limiter retrieves info from specific headers
	UsesRetryAfter bool          // Flag to check if the rate limiter uses a retry after value
	Logger         *log.Logger   // Logger for the rate limiter
}

// NewRateLimiter creates a new RateLimiter instance with the given parameters
func NewRateLimiter(args ...interface{}) *RateLimiter {
	rl := &RateLimiter{
		stopChan: make(chan struct{}),
		Logger:   log.NewLogger("{ratelimit}", log.INFO),
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			rl.Limit = v
			rl.Available = v
		case time.Duration:
			rl.Interval = v
		default:
			rl.Logger.Warning("Unsupported argument type in NewRateLimiter")
		}
	}

	rl.Start()

	return rl
}

// Start begins the rate limiter's internal timer
func (rl *RateLimiter) Start() {
	rl.Logger.Debug("Starting Rate Limiter")
	go func() {
		tickerInterval := rl.Interval
		if tickerInterval == 0 {
			tickerInterval = 1 * time.Minute
		}
		ticker := time.NewTicker(tickerInterval)
		rl.ResetTimestamp = time.Now().Add(tickerInterval).Unix()
		defer ticker.Stop()

		for {
			select {
			case <-rl.stopChan:
				rl.Logger.Debug("Stopping Rate Limiter")
				return
			case <-ticker.C:
				rl.mu.Lock()
				if rl.Limit > 0 && time.Now().Unix() >= rl.ResetTimestamp {
					rl.Available = rl.Limit
					rl.Logger.Debug("Rate limiter reset: Available limit set to ", rl.Limit)
				}
				rl.mu.Unlock()
			}
		}
	}()
}

// Throttle requests based on the remaining available rate limit.
func (rl *RateLimiter) Wait() {
	for {
		rl.mu.Lock()

		rl.TimeUntilReset = time.Until(time.Unix(rl.ResetTimestamp, 0))

		// If the reset time has passed, reset the available limit and return.
		if rl.TimeUntilReset <= 0 {
			if rl.Available < rl.Limit {
				rl.Available = rl.Limit
				rl.Requests = 0
				rl.mu.Unlock()
				return
			}
			rl.Requests = 0
			rl.Update()
			rl.ResetTimestamp = time.Now().Add(rl.Interval).Unix()
			rl.mu.Unlock()
			return
		}

		// Allow requests without delay until 90% of the limit is used
		if rl.Available > int(float64(rl.Limit)*0.10) && rl.Requests < int(float64(rl.Limit)*0.90) {
			rl.Update()
			rl.mu.Unlock()
			return
		}

		// Scale the wait time based on the ratio of remaining requests.
		// The closer we are to the rate limit, the longer the wait.
		remainingRequestsRatio := float64(rl.Available) / float64(rl.Limit)
		scaledWait := time.Duration(remainingRequestsRatio * 0.5 * float64(rl.TimeUntilReset))

		// Apply finer control of wait time when under 7.5% of the quota.
		// This helps in utilizing the available quota more effectively.
		if rl.Available <= int(float64(rl.Limit)*0.075) {
			scaledWait /= 2
		}

		// Cap the maximum wait time to 10 seconds to prevent overly long waits.
		maxWait := 10 * time.Second
		if scaledWait > maxWait {
			scaledWait = maxWait
		}

		// Add a small, random increment (up to 50ms) to the wait time.
		// This helps in avoiding synchronization issues in concurrent environments.
		randomIncrement := time.Duration(rand.Intn(50)) * time.Millisecond

		rl.mu.Unlock()

		rl.Logger.Tracef("Waiting for %v (scaled wait) + %v (random increment)\n", scaledWait, randomIncrement)
		rl.Logger.Debug("Time left until reset: ", rl.TimeUntilReset, " Available: ", rl.Available)

		// Sleep for the calculated duration.
		time.Sleep(scaledWait + randomIncrement)
	}
}

// Stop terminates the rate limiter's internal timer
func (rl *RateLimiter) Stop() {
	rl.Logger.Debug("Stopping Rate Limiter")
	close(rl.stopChan)
}

func (rl *RateLimiter) UpdateFromHeaders(headers http.Header) {
	// Log the start of the update process.
	rl.Logger.Trace("Updating Rate Limiter from headers")

	// Lock the RateLimiter to ensure thread-safe access to its fields.
	rl.mu.Lock()
	// Defer the unlocking so it's done automatically at the end of the function.
	defer rl.mu.Unlock()

	// Check if the RateLimiter uses a reset timestamp.
	if rl.UsesReset {
		// Try to get the "X-Rate-Limit-Reset" header.
		if resetHeader := headers.Get("X-Rate-Limit-Reset"); resetHeader != "" {
			// If the header is present and can be parsed to an int64, update the ResetTimestamp.
			if reset, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
				rl.ResetTimestamp = reset
			}
		}
	}

	// Check if the RateLimiter uses a retry-after value.
	if rl.UsesRetryAfter {
		// Try to get the "Retry-After" header.
		if retryAfterHeader := headers.Get("Retry-After"); retryAfterHeader != "" {
			// If the header is present and can be parsed to an int, update the RetryAfter.
			if retryAfter, err := strconv.Atoi(retryAfterHeader); err == nil {
				rl.RetryAfter = retryAfter
			}
		}
	}

	// Try to get the "X-Rate-Limit-Limit" header.
	if limitHeader := headers.Get("X-Rate-Limit-Limit"); limitHeader != "" {
		// If the header is present and can be parsed to an int, update the Limit and Available.
		if limit, err := strconv.Atoi(limitHeader); err == nil {
			rl.Limit = limit
			rl.Available = limit // Reset the available limit when the main limit changes.
		}
	}

	// Try to get the "X-Rate-Limit-Remaining" header.
	if remainingHeader := headers.Get("X-Rate-Limit-Remaining"); remainingHeader != "" {
		// If the header is present and can be parsed to an int, update the Available.
		if remaining, err := strconv.Atoi(remainingHeader); err == nil {
			rl.Available = remaining
		}
	}

	// Log the updated state of the Rate Limiter.
	rl.Logger.Debug("Rate limiter updated: Limit=", rl.Limit, ", Available=", rl.Available)
}

// Update reduces the available requests by 1
func (rl *RateLimiter) Update() {
	rl.Available--
	rl.Requests++
	rl.Logger.Debug("Rate limiter updated: Limit=", rl.Limit, ", Available=", rl.Available)
}
