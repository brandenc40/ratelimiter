# Ratelimiter

Rate limiter used to ensure a minimum duration between executions. 

Additionally supports the optional limit of max queue size. This can be used to ensure 
programs don't bottleneck due to having too many requests queued by the ratelimiter at a given time.


## Package Interface

```go 
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
	// NumQueued returns the current number of queued requests. If WithMaxQueueSize is not set, the result will always be 0.
	NumQueued() uint32
}
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