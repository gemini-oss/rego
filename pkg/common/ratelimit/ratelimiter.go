// pkg/common/ratelimit/ratelimit.go
package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gemini-oss/rego/pkg/common/crypt"
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
	ResetHeaders   bool          // Flag to check if the rate limiter retrieves info from specific headers
	ResetTimestamp int64         // Timestamp to reset the rate limiter
	RetryAfter     int           // Retry after time
	TimeUntilReset time.Duration // Time until the rate limiter resets
	UsesRetryAfter bool          // Flag to check if the rate limiter uses a retry after value
	Log            *log.Logger   // Logger for the rate limiter
}

// NewRateLimiter creates a new RateLimiter instance with the given parameters
func NewRateLimiter(args ...interface{}) *RateLimiter {
	rl := &RateLimiter{
		stopChan: make(chan struct{}),
		Log:      log.NewLogger("{ratelimit}", log.INFO),
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			rl.Limit = v
			rl.Available = v
		case time.Duration:
			rl.Interval = v
		default:
			rl.Log.Warning("Unsupported argument type in NewRateLimiter")
		}
	}

	rl.Start()

	return rl
}

// Start begins the rate limiter's internal timer
func (rl *RateLimiter) Start() {
	rl.Log.Debug("Starting Rate Limiter")
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
				rl.Log.Debug("Stopping Rate Limiter")
				return
			case <-ticker.C:
				rl.mu.Lock()
				if rl.Limit > 0 && time.Now().Unix() >= rl.ResetTimestamp {
					rl.Available = rl.Limit
					rl.Log.Debug("Rate limiter reset: Available limit set to ", rl.Limit)
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

		// Calculate the time until the next reset.
		timeUntilReset := time.Until(time.Unix(rl.ResetTimestamp, 0))

		// Check if it's time to reset the available limit.
		if timeUntilReset <= 0 {
			rl.resetAvailableLimit()
			rl.mu.Unlock()
			return
		}

		// Determine if a wait is needed based on the available requests.
		if rl.shouldWait() {
			waitDuration := rl.calculateWaitDuration(timeUntilReset)
			rl.mu.Unlock()
			rl.performWait(waitDuration)
			continue
		}

		// Proceed without waiting.
		if !rl.ResetHeaders {
			rl.decrementAvailable()
		}
		rl.mu.Unlock()
		return
	}
}

// resetAvailableLimit resets the available requests and requests count.
func (rl *RateLimiter) resetAvailableLimit() {
	if rl.Available < rl.Limit {
		rl.Available = rl.Limit
	}
	rl.Requests = 0
	rl.ResetTimestamp = time.Now().Add(rl.Interval).Unix()
}

// shouldWait determines if waiting is necessary based on the available requests.
func (rl *RateLimiter) shouldWait() bool {
	return rl.Available <= int(float64(rl.Limit)*0.10) || rl.Requests >= int(float64(rl.Limit)*0.90)
}

// calculateWaitDuration calculates the duration for which to wait.
func (rl *RateLimiter) calculateWaitDuration(timeUntilReset time.Duration) time.Duration {
	remainingRatio := float64(rl.Available) / float64(rl.Limit)
	scaledWait := time.Duration(remainingRatio * 0.5 * float64(timeUntilReset))

	if rl.Available <= int(float64(rl.Limit)*0.075) {
		scaledWait /= 2
	}

	maxWait := 10 * time.Second
	if scaledWait > maxWait {
		scaledWait = maxWait
	}

	randomIncrement, err := crypt.SecureRandomInt(50)
	if err != nil {
		randomIncrement = 0
	}
	return scaledWait + time.Duration(randomIncrement)*time.Millisecond
}

// performWait sleeps for the specified duration.
func (rl *RateLimiter) performWait(duration time.Duration) {
	rl.Log.Tracef("Waiting for %v\n", duration)
	time.Sleep(duration)
}

// decrementAvailable decrements the available requests and increments the request count.
func (rl *RateLimiter) decrementAvailable() {
	rl.Available--
	rl.Requests++
	rl.Log.Debug("Rate limiter updated: Limit=", rl.Limit, ", Available=", rl.Available)
}

// Stop terminates the rate limiter's internal timer
func (rl *RateLimiter) Stop() {
	rl.Log.Debug("Stopping Rate Limiter")
	close(rl.stopChan)
}

func (rl *RateLimiter) UpdateFromHeaders(headers http.Header) {
	// Add a check for nil headers
	if headers == nil {
		return // or handle this case as needed
	}

	// Log the start of the update process.
	rl.Log.Trace("Updating Rate Limiter from headers")

	// Lock the RateLimiter to ensure thread-safe access to its fields.
	rl.mu.Lock()
	// Defer the unlocking so it's done automatically at the end of the function.
	defer rl.mu.Unlock()

	// Try to get the "X-Rate-Limit-Reset" header.
	if resetHeader := headers.Get("X-Rate-Limit-Reset"); resetHeader != "" {
		// If the header is present and can be parsed to an int64, update the ResetTimestamp.
		if reset, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
			rl.ResetTimestamp = reset
		}
	}

	// Try to get the "Retry-After" header.
	if retryAfterHeader := headers.Get("Retry-After"); retryAfterHeader != "" {
		// If the header is present and can be parsed to an int, update the RetryAfter.
		if retryAfter, err := strconv.Atoi(retryAfterHeader); err == nil {
			rl.RetryAfter = retryAfter
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
	rl.Log.Debug("Rate limiter updated: Limit=", rl.Limit, ", Available=", rl.Available)
}
