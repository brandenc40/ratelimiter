package ratelimiter

import (
	"errors"
	"time"
)

const (
	// NoLimit can be used as the minDuration arg value to NewRateLimiter if no limiting is required.
	NoLimit = time.Duration(-1)
)

var (
	// ErrQueueFull is returned with the queue limit set by WithMaxQueueSize is exceeded.
	ErrQueueFull = errors.New("ratelimiter: queue is full")
)

// RateLimiter provides functionality to block until ready to ensure a rate limit is not exceeded
type RateLimiter interface {
	// Wait blocks until the next call is ready based on the minimum time between calls.
	Wait() error
	// QueueSize returns the current number of queued requests. If WithMaxQueueSize is not set, the result will always be 0.
	QueueSize() uint32
}

type config struct {
	durationBetweenCalls time.Duration
	maxQueue             uint32
}

// NewRateLimiter builds a new rate limiter used to ensure calls adhere to a minimum duration between calls.
func NewRateLimiter(minDuration time.Duration, options ...Option) RateLimiter {
	cfg := config{
		durationBetweenCalls: minDuration,
		maxQueue:             0,
	}
	for _, opt := range options {
		opt.apply(&cfg)
	}
	return newMutexLimiter(cfg)
}

// Option to configure the rate limiter
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
