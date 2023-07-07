// pkg/common/ratelimit/ratelimit.go
package ratelimit

import (
	"time"
)

type RateLimiter struct {
	rate   int           // maximum number of requests
	bucket chan struct{} // a bucket holds tokens
}

func NewRateLimiter(rate int) *RateLimiter {
	bucket := make(chan struct{}, rate)
	return &RateLimiter{rate: rate, bucket: bucket}
}

func (rl *RateLimiter) Start() {
	ticker := time.NewTicker(time.Second / time.Duration(rl.rate))
	defer ticker.Stop()
	for range ticker.C {
		rl.bucket <- struct{}{}
	}
}

func (rl *RateLimiter) Wait() {
	<-rl.bucket
}
