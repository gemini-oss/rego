# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `ratelimit` package provides adaptive rate limiting for API requests using a token bucket pattern with intelligent throttling. It supports both fixed-interval and dynamic header-based rate limiting, automatically preventing API rate limit violations across all Rego service integrations.

**Key Features:**
- Token bucket implementation with automatic refill
- Progressive throttling to avoid hitting limits
- Dynamic limit updates from response headers
- Thread-safe concurrent operation
- Integrated jitter to prevent thundering herd
- Automatic integration with requests package
- Background goroutine for periodic resets

## Architecture

### Core Components

1. **RateLimiter Structure**:
   - Token bucket with `Available` counter and `Limit`
   - Time-based reset with `ResetTimestamp`
   - Thread-safe with mutex protection
   - Background goroutine for automatic resets

2. **Throttling Algorithm**:
   - Progressive: Starts throttling at 10% remaining OR 90% usage
   - Scaled wait: Duration proportional to remaining capacity
   - Jitter: Random 0-50ms added to prevent thundering herd
   - Max wait: Capped at 10 seconds
   - Extra caution: Half wait time when <7.5% remaining

### Key Features

- **Adaptive**: Updates limits from HTTP response headers
- **Smooth**: Distributes requests evenly, avoiding bursts
- **Automatic**: Integrated with requests package transparently
- **Configurable**: Fixed or dynamic limits per service

## Development Tasks

### Creating a Rate Limiter

```go
// Fixed limits (order doesn't matter)
rl := ratelimit.NewRateLimiter(120, 1*time.Minute)  // 120/min
rl := ratelimit.NewRateLimiter(1*time.Minute, 120)  // Same result

// Dynamic (header-based) - no arguments
rl := ratelimit.NewRateLimiter()  // Uses response headers
rl.ResetHeaders = true

// Just limit (defaults to 1 minute interval)
rl := ratelimit.NewRateLimiter(100)  // 100/min

// Attach to HTTP client
client.HTTP.RateLimiter = rl
```

### Flexible Constructor

The constructor accepts variadic arguments in any order:
- `int`: Sets the limit
- `time.Duration`: Sets the interval
- No arguments: Creates header-based limiter
- Default interval: 1 minute if not specified

### Supported Headers

The rate limiter automatically reads:
- `X-Rate-Limit-Limit`: Total allowed requests
- `X-Rate-Limit-Remaining`: Requests left in window
- `X-Rate-Limit-Reset`: Unix timestamp for reset
- `Retry-After`: Seconds to wait (stored, not used)

### Wait Algorithm (Corrected)

```go
// Actual implementation
if available < 10% of limit OR used > 90% of limit:
    remainingRatio = available / limit
    scaledWait = remainingRatio * 0.5 * timeUntilReset

    if available < 7.5% of limit:
        scaledWait = scaledWait / 2  // Extra caution

    jitter = random(0, 50ms)
    finalWait = min(scaledWait + jitter, 10s)
    sleep(finalWait)
```

**Key Points:**
- Uses remaining ratio (available/limit), not depletion ratio
- Multiplies by 0.5 to be less aggressive
- Jitter prevents synchronized retries
- 10-second cap prevents excessive waits

## Important Notes

- **Goroutine Cleanup**: Always call `Stop()` to terminate background goroutine
- **Automatic Integration**: requests package calls `Wait()` before each request
- **Header Trust**: Header-based mode relies entirely on server headers
- **Debug Logging**: TRACE level shows wait calculations and decisions
- **Default Interval**: 1 minute if not specified in constructor
- **Thread Safety**: All operations protected by mutex
- **No Manual Decrement**: In header mode, Available is set directly from headers

### Method Reference

```go
// Core methods
Wait()                           // Block until safe to proceed
Decrement()                      // Reduce available count (fixed mode)
UpdateFromHeaders(headers)       // Update limits from response
Stop()                           // Clean up background goroutine

// Fields
Limit           int              // Maximum requests allowed
Available       int              // Current tokens available
Interval        time.Duration    // Reset interval
ResetTimestamp  time.Time        // Next reset time
ResetHeaders    bool             // Use dynamic headers?
RetryAfter      int              // Stored from Retry-After header
```

## Common Patterns

### Service Examples

| Service | Type | Configuration | Notes |
|---------|------|---------------|-------|
| SnipeIT | Fixed | 120/min | API has hard limit |
| Google Admin | Fixed | 2400/min | Per-API limits |
| Google Drive | Fixed | 12000/min | High volume |
| Okta | Dynamic | Headers | Varies by endpoint |
| Atlassian | Dynamic | Headers | Adaptive limits |
| Jamf | Fixed | 100/min | Conservative default |
| Slack | Fixed | 50/min | Tier-based limits |

```go
// Fixed limits (SnipeIT)
rl := NewRateLimiter(120, 1*time.Minute)

// High volume (Google Drive)
rl := NewRateLimiter(12000, 1*time.Minute)

// Dynamic (Okta)
rl := NewRateLimiter()  // No args = header-based
rl.ResetHeaders = true
```

### Integration Flow

1. Service creates rate limiter during init
2. Attaches to `requests.Client.RateLimiter`
3. Each HTTP response updates limits via headers
4. `Wait()` called automatically before next request
5. Transparent throttling without code changes

## Common Pitfalls

1. **Forgetting Stop()**: Leaves goroutine running (memory leak)
2. **Wrong Mode**: Using fixed limits when API provides headers
3. **Too Aggressive**: Setting limits higher than API allows
4. **Time Zones**: Reset timestamps are Unix time (UTC)
5. **Burst Traffic**: Initial requests may hit limits quickly
6. **Multiple Stop()**: Calling Stop() multiple times is safe
7. **Zero Limits**: If headers return 0, no throttling occurs

## Performance Characteristics

- **Memory**: Minimal - just counters and timestamps
- **CPU**: Negligible except during wait calculations
- **Goroutines**: One background ticker per rate limiter
- **Mutex Contention**: Brief locks for counter updates

## Debugging

```go
// Enable trace logging
rl.Log.Verbosity = log.TRACE

// Inspect state
log.Printf("Limit: %d, Available: %d, Reset: %v",
    rl.Limit, rl.Available, rl.ResetTimestamp)

// Force immediate throttling (testing)
rl.Available = 1
rl.Limit = 100
```

## Advanced Usage

### Custom Rate Limiting Strategy
```go
// Implement more aggressive throttling
type AggressiveRateLimiter struct {
    *ratelimit.RateLimiter
}

func (a *AggressiveRateLimiter) Wait() {
    // Start throttling at 50% instead of 10%
    if float64(a.Available)/float64(a.Limit) < 0.5 {
        // Custom wait logic
    }
}
```

### Monitoring Rate Limit Usage
```go
// Track usage percentage
usagePercent := 100 * (1 - float64(rl.Available)/float64(rl.Limit))
if usagePercent > 80 {
    log.Warning("High rate limit usage: %.1f%%", usagePercent)
}
```
