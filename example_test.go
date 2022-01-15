package ratelimiter_test

import (
	"fmt"
	"time"

	"github.com/brandenc40/ratelimiter"
)

func Example() {
	rl := ratelimiter.New(
		10*time.Millisecond,               // 10ms between calls (100 rps)
		ratelimiter.WithMaxQueueSize(100), // max of 100 requests queued up before failure
	)

	var (
		start    = time.Now()
		nSuccess = 0
		nError   = 0
	)
	for i := 0; i < 100; i++ {
		if err := rl.Wait(); err != nil {
			nError++
		} else {
			nSuccess++
		}
	}

	elapsed := time.Since(start)

	// 10ms each for 100 requrests == 990ms total (first request is 0ms)
	fmt.Printf("(timeElapsed >= 990ms) == %v\n", elapsed.Milliseconds() >= 950)
	fmt.Println("nSuccess:", nSuccess)
	fmt.Println("nError:", nError)

	// Output:
	// (timeElapsed >= 990ms) == true
	// nSuccess: 100
	// nError: 0
}
