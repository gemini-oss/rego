// pkg/internal/tests/common/retry/retry_test.go
package retry_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/retry"
)

type MockTime struct {
	sleepDurations []time.Duration
}

func (m *MockTime) Sleep(duration time.Duration) {
	m.sleepDurations = append(m.sleepDurations, duration)
}

func (m *MockTime) GetSleepDurations() []time.Duration {
	return m.sleepDurations
}

func TestSuccessfulBeforeMaxRetries(t *testing.T) {
	mockTime := MockTime{}

	attemptsBeforeSuccess := 3
	currentAttempt := 0

	operation := func() error {
		currentAttempt++
		if currentAttempt >= attemptsBeforeSuccess {
			return nil // Simulate success
		}
		return fmt.Errorf("temporary error")
	}

	shouldRetry := func(err error) bool {
		return err != nil
	}

	err := retry.Retry(operation, shouldRetry, &mockTime)
	if err != nil {
		t.Fatalf("Retry should have succeeded but got error: %v", err)
	}
	if currentAttempt != attemptsBeforeSuccess {
		t.Errorf("Expected %d attempts before success, but got %d", attemptsBeforeSuccess, currentAttempt)
	}
}

func TestRetryTimingAndJitter(t *testing.T) {
	mockTime := MockTime{}

	operation := func() error {
		return fmt.Errorf("rate limited")
	}

	shouldRetry := func(err error) bool {
		return err != nil
	}

	_ = retry.Retry(operation, shouldRetry, &mockTime)
	sleepDurations := mockTime.GetSleepDurations()

	if len(sleepDurations) != retry.MaxRetries-1 {
		t.Fatalf("Expected %d retry attempts, got %d", retry.MaxRetries-1, len(sleepDurations))
	}

	minBackoff := time.Duration(retry.MinBackoff) * time.Millisecond
	maxBackoff := time.Duration(retry.MaxBackoff) * time.Millisecond

	for i, duration := range sleepDurations {
		expectedBackoff := minBackoff * time.Duration(1<<i)
		if expectedBackoff > maxBackoff {
			expectedBackoff = maxBackoff
		}

		if duration < 0 || duration > expectedBackoff {
			t.Errorf("Sleep duration %v on retry %d is outside expected range [0, %v]", duration, i+1, expectedBackoff)
		}
	}
}

func TestExceedingMaxRetries(t *testing.T) {
	mockTime := MockTime{}

	operation := func() error {
		return fmt.Errorf("permanent error")
	}

	shouldRetry := func(err error) bool {
		return err != nil
	}

	err := retry.Retry(operation, shouldRetry, &mockTime)
	if err == nil {
		t.Fatalf("Expected error after maximum retries, but got nil")
	}

	sleepDurations := mockTime.GetSleepDurations()
	if len(sleepDurations) != retry.MaxRetries-1 {
		t.Errorf("Expected %d retries, but got %d", retry.MaxRetries-1, len(sleepDurations))
	}
}

func TestNonRetryableError(t *testing.T) {
	mockTime := MockTime{}

	operation := func() error {
		return fmt.Errorf("non-retryable error")
	}

	shouldRetry := func(err error) bool {
		return false // Never retry
	}

	err := retry.Retry(operation, shouldRetry, &mockTime)
	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}

	sleepDurations := mockTime.GetSleepDurations()
	if len(sleepDurations) != 0 {
		t.Errorf("Expected 0 retries, but got %d", len(sleepDurations))
	}
}
