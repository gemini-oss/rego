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
	retry.SetRandomSeed(0)
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
	retry.SetRandomSeed(0) // Set a fixed seed for consistent jitter
	mockTime := MockTime{}

	operation := func() error {
		return fmt.Errorf("rate limited")
	}

	_ = retry.Retry(operation, &mockTime)
	sleepDurations := mockTime.GetSleepDurations()

	if len(sleepDurations) != retry.MaxRetries {
		t.Fatalf("Expected %d retry attempts, got %d", retry.MaxRetries, len(sleepDurations))
	}

	// Verbose logging for each retry interval
	for i, duration := range sleepDurations {
		t.Logf("Retry %d: Slept for %v", i+1, duration)
	}
}

func TestExceedingMaxRetries(t *testing.T) {
	retry.SetRandomSeed(0)
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
