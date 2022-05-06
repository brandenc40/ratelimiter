package ratelimiter_test

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	rl "github.com/brandenc40/ratelimiter"
)

func ExampleRateLimiter() {
	limiter := rl.New(
		// 10ms between calls (100 rps)
		rl.PerDuration(100, time.Second),
	)

	var (
		start    = time.Now()
		nSuccess = 0
		nError   = 0
	)

	for i := 0; i < 100; i++ {
		if err := limiter.Wait(); err != nil {
			// queue is not limited so this should never return an error
			nError++
			continue
		}
		nSuccess++

	}

	elapsed := time.Since(start)

	// 10ms each for 100 requrests == 990ms total (first request is 0ms)
	fmt.Printf("(timeElapsed >= 990ms) == %v\n", elapsed.Milliseconds() >= 990)
	fmt.Println("nSuccess:", nSuccess)
	fmt.Println("nError:", nError)

	// Output:
	// (timeElapsed >= 990ms) == true
	// nSuccess: 100
	// nError: 0
}

func ExampleWithMaxQueueSize() {
	limiter := rl.New(
		// 1 request per second
		rl.PerDuration(1, time.Second),
		// only one can be queued at a time
		rl.WithMaxQueueSize(1),
	)

	// first call is executed immediately and not useful for this example
	_ = limiter.Wait()

	// ensure a single proc handles the goroutines
	runtime.GOMAXPROCS(1)

	startTime := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			if err := limiter.Wait(); err != nil {
				fmt.Println(i, "err:", err)
				return
			}
			fmt.Println(i, "success", time.Since(startTime).Round(time.Second))

		}(i)

		// quick sleep to ensure goroutines are started in order
		time.Sleep(time.Nanosecond)
	}

	wg.Wait()

	// Output:
	// 2 err: ratelimiter: queue is full
	// 3 err: ratelimiter: queue is full
	// 4 err: ratelimiter: queue is full
	// 0 success 1s
	// 1 success 2s
}
