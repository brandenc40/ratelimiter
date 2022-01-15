package ratelimiter_test

import (
	"time"

	"github.com/brandenc40/ratelimiter"
)

func Example() {
	rl := ratelimiter.New(
		10*time.Millisecond,               // 10ms between calls (100 rps)
		ratelimiter.WithMaxQueueSize(100), // max of 100 requests queued up before failure
	)

	for i := 0; i < 100; i++ {
		if err := rl.Wait(); err != nil {
			// handle err
		}
		// do some rate limited functionality
	}
}
