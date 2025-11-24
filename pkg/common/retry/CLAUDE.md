# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `retry` package provides a robust retry mechanism with exponential backoff and cryptographically secure jitter. It's designed to handle transient failures gracefully, preventing thundering herd problems while maximizing the chance of successful operation completion.

## Architecture

### Core Components

1. **Retry Function**:
   ```go
   func Retry(operation func() error, shouldRetry func(error) bool, time Time) error
   ```
   - Executes operation up to 5 times
   - Uses custom predicate to determine retryability
   - Returns on success or non-retryable error

2. **Backoff Strategy**:
   - Exponential: 500ms → 1s → 2s → 3s (capped)
   - Jitter: 0 to backoff milliseconds (cryptographically secure)
   - Formula: `min(500 * 2^retryCount, 3000) + random(0, backoff)`

3. **Time Interface**:
   ```go
   type Time interface {
       Sleep(time.Duration)
   }
   ```
   - Allows dependency injection for testing
   - `RealTime` for production, mock for tests

### Configuration

```go
const (
    MaxRetries = 5    // Total attempts (1 initial + 4 retries)
    MinBackoff = 500  // Starting backoff in milliseconds
    MaxBackoff = 3000 // Maximum backoff in milliseconds
)
```

## Development Tasks

### Basic Usage

```go
import "github.com/gemini-oss/rego/pkg/common/retry"

err := retry.Retry(
    // Operation to retry
    func() error {
        return doSomethingThatMightFail()
    },
    // Retry predicate
    func(err error) bool {
        return err != nil && isTransientError(err)
    },
    // Time implementation
    retry.RealTime{},
)
```

### HTTP Request Example

```go
// From requests package
err := retry.Retry(
    func() error {
        resp, body, err = client.doRequest(ctx, method, url)
        return err
    },
    func(err error) bool {
        if err == nil {
            return false
        }
        if resp == nil {
            return true // Network error
        }
        return IsRetryableStatusCode(resp.StatusCode)
    },
    retry.RealTime{},
)
```

### Testing with Mock Time

```go
type MockTime struct {
    sleepCalls []time.Duration
}

func (m *MockTime) Sleep(d time.Duration) {
    m.sleepCalls = append(m.sleepCalls, d)
}

// In test
mockTime := &MockTime{}
err := retry.Retry(operation, shouldRetry, mockTime)
// Assert on mockTime.sleepCalls
```

## Important Notes

- Uses cryptographically secure random for jitter (via crypt package)
- No sleep after final attempt (fail fast on last try)
- Gracefully handles jitter generation failures
- Thread-safe - can be used concurrently
- Always returns the last error if all retries fail

## Best Practices

1. **Idempotency**: Only retry idempotent operations
2. **Clear Predicates**: Be specific about retryable errors
3. **Context Awareness**: Check context cancellation in operations
4. **Logging**: Log retry attempts for debugging
5. **Error Types**: Consider which errors are truly transient

## Common Patterns

### Retryable HTTP Status Codes

```go
// From requests package
retryableCodes := []int{408, 429, 500, 502, 503, 504}
```

### Network Errors

```go
shouldRetry := func(err error) bool {
    // Retry on timeout or connection errors
    var netErr net.Error
    return errors.As(err, &netErr) && netErr.Temporary()
}
```

### Database Operations

```go
shouldRetry := func(err error) bool {
    // Retry on deadlock or connection issues
    return isDeadlock(err) || isConnectionError(err)
}
```

## Common Pitfalls

1. **Non-Idempotent Operations**: Don't retry operations with side effects
2. **Infinite Retries**: The package prevents this with MaxRetries
3. **Missing Context**: Operations should respect context cancellation
4. **Too Aggressive**: Consider if 5 retries is appropriate
5. **Synchronization**: Jitter prevents this, but be aware of retry storms
