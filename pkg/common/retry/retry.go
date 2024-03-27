// pkg/common/retry/retry.go
package retry

import (
	"math/rand"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

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

func SetRandomSeed(seed int64) {
	random = rand.New(rand.NewSource(seed))
}

// BackoffWithJitter returns a duration for exponential backoff with jitter
func BackoffWithJitter(retryCount int) time.Duration {
	backoff := MinBackoff * (1 << retryCount)
	if backoff > MaxBackoff {
		backoff = MaxBackoff
	}
	jitter := random.Intn(backoff)
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
