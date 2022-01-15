# Ratelimiter [![Go Reference](https://pkg.go.dev/badge/github.com/brandenc40/ratelimiter#example-package.svg)](https://pkg.go.dev/github.com/brandenc40/ratelimiter#example-package)

Rate limiter used to ensure a minimum duration between executions. 

Additionally supports the optional limit of max queue size. This can be used to ensure 
programs don't bottleneck due to having too many requests queued by the ratelimiter at a given time.


## Package Interface

```go 
// NoLimit can be used as the minDuration arg value to New if no limiting is required.
const NoLimit = time.Duration(-1)

// ErrQueueFull is returned with the queue limit set by WithMaxQueueSize is exceeded.
var ErrQueueFull = errors.New("ratelimiter: queue is full")

// RateLimiter provides functionality to block until ready to ensure a rate limit is not exceeded.
type RateLimiter interface {
    // Wait blocks until the next call is ready based on the minimum time between calls.
    Wait() error
    // NumQueued returns the current number of queued requests. If WithMaxQueueSize is not set, the result will always be 0.
    NumQueued() uint32
}

// New builds a new rate limiter used to ensure calls adhere to a minimum duration between calls.
func New(minDuration time.Duration, options ...Option) RateLimiter

// WithMaxQueueSize sets the maximum number of requests that can be queued up. If the queue
// limit is reached, ErrQueueFull will be returned when Wait is called.
func WithMaxQueueSize(maxQueue uint32) Option
```

## Usage Example

```go
package main

import (
    "time"

    "github.com/brandenc40/ratelimiter"
)

func main() {
    rl := ratelimiter.New(
        10*time.Millisecond,               // 10ms between calls (100 rps)
        ratelimiter.WithMaxQueueSize(100), // (optional) max of 100 requests queued up before failure
    )

    for i := 0; i < 100; i++ {
        if err := rl.Wait(); err != nil {
            // handle err
        }
    // do some rate limited functionality
    }
}
```
