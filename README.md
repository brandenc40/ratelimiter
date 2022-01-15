# Ratelimiter

Rate limiter used to ensure a minimum duration between executions. Additionally supports the optional limit of max queue size. 
This can be used to ensure programs don't bottleneck due to having too many requests queued by the ratelimiter at a given time.

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