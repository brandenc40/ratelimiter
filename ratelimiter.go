package ratelimiter

import (
	"errors"
	"time"
)

// NoLimit can be used as the minDuration arg value to New if no limiting is required.
const NoLimit = time.Duration(-1)

// ErrQueueFull is returned with the queue limit set by WithMaxQueueSize is exceeded.
var ErrQueueFull = errors.New("ratelimiter: queue is full")

// RateLimiter provides functionality to block until ready to ensure a rate limit is not exceeded.
type RateLimiter interface {
	// Wait blocks until the next call is ready based on the minimum time between calls.
	Wait() error
	// NumQueued returns the current number of queued requests. If WithMaxQueueSize is not set,
	// the result will always be 0.
	NumQueued() uint32
}

type config struct {
	durationBetweenCalls time.Duration
	maxQueue             uint32
}

// New builds a new rate limiter used to ensure calls adhere to a minimum duration between calls.
func New(minDuration time.Duration, options ...Option) RateLimiter {
	cfg := config{
		durationBetweenCalls: minDuration,
		maxQueue:             0,
	}

	for _, opt := range options {
		opt.apply(&cfg)
	}

	return newMutexLimiter(cfg)
}

// PerDuration is a helper function for determining the min duration between calls with a common requests per
// duration syntax.
//
// e.g. for 100qps:
//	ratelimiter.New(
//		ratelimiter.PerDuration(100, time.Second)
//	)
//
func PerDuration(n int, duration time.Duration) time.Duration {
	return duration / time.Duration(n)
}

// Option to configure the rate limiter.
type Option interface {
	apply(*config)
}

// WithMaxQueueSize sets the maximum number of requests that can be queued up. If the queue
// limit is reached, ErrQueueFull will be returned when Wait is called.
func WithMaxQueueSize(maxQueue uint32) Option {
	return maxQueueOption(maxQueue)
}

type maxQueueOption uint32

func (m maxQueueOption) apply(cfg *config) {
	cfg.maxQueue = uint32(m)
}
