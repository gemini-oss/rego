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

	err := retry.Retry(operation, &mockTime)
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

	_ = retry.Retry(operation, &mockTime)
	sleepDurations := mockTime.GetSleepDurations()

	if len(sleepDurations) != retry.MaxRetries {
		t.Fatalf("Expected %d retry attempts, got %d", retry.MaxRetries, len(sleepDurations))
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

	err := retry.Retry(operation, &mockTime)
	if err == nil {
		t.Fatalf("Expected error after maximum retries, but got nil")
	}

	sleepDurations := mockTime.GetSleepDurations()
	if len(sleepDurations) != retry.MaxRetries {
		t.Errorf("Expected %d retries, but got %d", retry.MaxRetries, len(sleepDurations))
	}
}
