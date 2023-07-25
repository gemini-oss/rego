// pkg/common/ratelimit/ratelimit.go
package ratelimit

import (
	"time"
)

type RateLimiter struct {
	Rate   int           // maximum number of requests
	Bucket chan struct{} // a bucket holds tokens
	Ticker *time.Ticker
}

func NewRateLimiter(rate int) *RateLimiter {
	bucket := make(chan struct{}, rate)
	return &RateLimiter{Rate: rate, Bucket: bucket, Ticker: time.NewTicker(time.Second / time.Duration(rate))}
}

func (rl *RateLimiter) Start() {
	for range rl.Ticker.C {
		rl.Bucket <- struct{}{}
	}
}

func (rl *RateLimiter) Stop() {
	rl.Ticker.Stop()
}

func (rl *RateLimiter) UpdateRate(rate int) {
	rl.Stop()
	rl.Rate = rate
	rl.Ticker = time.NewTicker(time.Second / time.Duration(rl.Rate))
	go rl.Start()
}

func (rl *RateLimiter) Wait() {
	<-rl.Bucket
}
