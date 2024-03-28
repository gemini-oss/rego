// pkg/common/retry/retry.go
package retry

import (
	"time"

	"github.com/gemini-oss/rego/pkg/common/crypt"
)

const (
	MaxRetries = 5
	MinBackoff = 500
	MaxBackoff = 3000
)

type Time interface {
	Sleep(duration time.Duration)
}

type RealTime struct{}

func (RealTime) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// BackoffWithJitter returns a duration for exponential backoff with jitter using a secure random source
func BackoffWithJitter(retryCount int) time.Duration {
	backoff := MinBackoff * (1 << retryCount)
	if backoff > MaxBackoff {
		backoff = MaxBackoff
	}

	jitter, err := crypt.SecureRandomInt(backoff)
	if err != nil {
		// Handle the error or default to a non-jittered backoff
		return time.Duration(backoff) * time.Millisecond
	}

	// Ensuring jitter is within the MinBackoff and backoff range
	if jitter < MinBackoff {
		jitter = MinBackoff
	}

	return time.Duration(jitter) * time.Millisecond
}

// Retry retries the given operation up to MaxRetries times, with exponential backoff and jitter
func Retry(operation func() error, time Time) error {
	var err error
	for i := 0; i < MaxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		time.Sleep(BackoffWithJitter(i))
	}
	return err
}
