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

type RateLimiter struct {
	Limit          int
	Available      int
	ResetTimestamp int64
	RetryAfter     int
	UsesReset      bool
	UsesRetryAfter bool
	stopChan       chan struct{}
	mu             sync.Mutex
	Logger         *log.Logger
}

func NewRateLimiter(args ...int) *RateLimiter {

	if len(args) > 0 {
		return &RateLimiter{
			Limit:     args[0],
			Available: args[0],
			stopChan:  make(chan struct{}),
			Logger:    log.NewLogger("{ratelimit}", log.INFO),
		}
	} else {
		return &RateLimiter{
			Limit:     0,
			Available: 0,
			stopChan:  make(chan struct{}),
			Logger:    log.NewLogger("{ratelimit}", log.INFO),
		}
	}
}

func (rl *RateLimiter) Start() {
	rl.Logger.Debug("Starting Rate Limiter")
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
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

        timeUntilReset := time.Until(time.Unix(rl.ResetTimestamp, 0))

        // If the reset time has passed, reset the available limit and return.
        if timeUntilReset <= 0 {
            rl.Available = rl.Limit
            rl.mu.Unlock()
            return
        }

        // Allow requests without delay until 90% of the limit is used.
        if rl.Available > int(float64(rl.Limit) * 0.10) {
            rl.Available--
            rl.mu.Unlock()
            return
        }

        // Scale the wait time based on the ratio of remaining requests.
        // The closer we are to the rate limit, the longer the wait.
        remainingRequestsRatio := float64(rl.Available) / float64(rl.Limit)
        scaledWait := time.Duration(remainingRequestsRatio * 0.5 * float64(timeUntilReset))

        // Apply finer control of wait time when under 7.5% of the quota.
        // This helps in utilizing the available quota more effectively.
        if rl.Available <= int(float64(rl.Limit) * 0.075) {
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

        rl.Logger.Debugf("Waiting for %v (scaled wait) + %v (random increment)\n", scaledWait, randomIncrement)
		rl.Logger.Println("Time left until reset: ", timeUntilReset, " Available: ", rl.Available, " Limit: ", rl.Limit, " ResetTimestamp: ", rl.ResetTimestamp, " RetryAfter: ", rl.RetryAfter, " UsesReset: ", rl.UsesReset, " UsesRetryAfter: ", rl.UsesRetryAfter)

        // Sleep for the calculated duration.
        time.Sleep(scaledWait + randomIncrement)
    }
}

func (rl *RateLimiter) Stop() {
	rl.Logger.Debug("Stopping Rate Limiter")
	close(rl.stopChan)
}

func (rl *RateLimiter) UpdateFromHeaders(headers http.Header) {
    // Log the start of the update process.
	rl.Logger.Debug("Updating Rate Limiter from headers")

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

// UpdateRate can be used to dynamically change the rate limit
func (rl *RateLimiter) UpdateRate(newLimit int) {
	rl.Logger.Trace("Updating Rate Limiter rate")
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.Limit = newLimit
	rl.Available = newLimit
}
