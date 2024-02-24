// pkg/internal/tests/common/ratelimit/ratelimit_test.go
package ratelimit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/ratelimit"
)

func setupDynamicMockAPIServer() *httptest.Server {
	var requestCount int
	var resetTimestamp int64

	handler := func(w http.ResponseWriter, r *http.Request) {
		limit := 300
		if time.Now().Unix() >= resetTimestamp {
			requestCount = 0
			resetTimestamp = time.Now().Add(5 * time.Second).Unix()
		}

		requestCount++
		remaining := limit - requestCount

		w.Header().Set("X-Rate-Limit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-Rate-Limit-Reset", strconv.FormatInt(resetTimestamp, 10))

		if remaining < 0 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"ok"}`)
		}
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestNewRateLimiter(t *testing.T) {
	rl := ratelimit.NewRateLimiter(10, 5*time.Second)
	rl.Logger.Verbosity = log.TRACE
	if rl.Limit != 10 || rl.Available != 10 {
		t.Errorf("Expected Limit and Available to be 10, got %d and %d", rl.Limit, rl.Available)
	}
	if rl.Interval != 5*time.Second {
		t.Errorf("Expected CustomResetInterval to be 5 seconds, got %v", rl.Interval)
	}

	rl.Logger.Delete()
}

func TestRateLimiterWithHeaders(t *testing.T) {
	server := setupDynamicMockAPIServer()
	defer server.Close()

	client := &http.Client{}
	rl := ratelimit.NewRateLimiter()
	rl.UsesReset = true
	rl.Logger.Verbosity = log.TRACE

	var rateLimited bool

	for i := 0; i < 900; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		rl.UpdateFromHeaders(resp.Header)
		rl.Wait()

		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	if rateLimited {
		t.Errorf("Expected to not be rate limited, but was")
	}

	if rl.Available < 0 {
		t.Errorf("Available requests should not be negative, got %d", rl.Available)
	}

	rl.Logger.Delete()
}

func TestRateLimiterNoHeaders(t *testing.T) {
	server := setupDynamicMockAPIServer()
	defer server.Close()

	client := &http.Client{}
	rl := ratelimit.NewRateLimiter(120, 5*time.Second)
	rl.Logger.Verbosity = log.TRACE

	var rateLimited bool
	requestsToMake := rl.Available * 3 // Set available requests above the limit

	for i := 0; i < requestsToMake; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		rl.Logger.Print(i)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		rl.Wait()
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	if rateLimited {
		t.Errorf("Expected to not be rate limited, but was")
	}

	if rl.Available < int(float64(rl.Limit)*0.05) {
		t.Errorf("Available requests dropped too low, got %d", rl.Available)
	}

	rl.Logger.Delete()
}

func TestRateLimiterReset(t *testing.T) {
	resetTime := time.Now().Add(10 * time.Second).Unix() // Reset after 10 seconds
	rl := ratelimit.NewRateLimiter(5)
	rl.Logger.Verbosity = log.TRACE
	rl.ResetTimestamp = resetTime
	rl.Start()
	defer rl.Stop()

	time.Sleep(5 * time.Second) // Wait for the reset time to pass

	if rl.Available != 5 {
		t.Errorf("Expected rate limit to reset, but available is %d", rl.Available)
	}

	rl.Logger.Delete()
}

func TestUpdateFromHeaders(t *testing.T) {
	rl := ratelimit.NewRateLimiter(5)
	rl.Logger.Verbosity = log.TRACE
	rl.UsesReset = true
	rl.Start()
	defer rl.Stop()

	resetTime := time.Now().Add(1 * time.Minute).Unix()
	headers := http.Header{}
	headers.Set("X-Rate-Limit-Reset", strconv.FormatInt(resetTime, 10))
	headers.Set("X-Rate-Limit-Remaining", "3")

	rl.UpdateFromHeaders(headers)

	if rl.ResetTimestamp != resetTime {
		t.Errorf("Expected ResetTimestamp to be %d, got %d", resetTime, rl.ResetTimestamp)
	}
	if rl.Available != 3 {
		t.Errorf("Expected Available to be 3, got %d", rl.Available)
	}

	rl.Logger.Delete()
}
