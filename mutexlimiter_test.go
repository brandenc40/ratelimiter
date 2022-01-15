package ratelimiter

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestRateLimiter_Wait(t *testing.T) {
	r := newMutexLimiter(config{
		durationBetweenCalls: time.Millisecond / 2,
		maxQueue:             0,
	})
	t.Run("async", func(t *testing.T) {
		start := time.Now()
		var eg errgroup.Group
		for i := 0; i < 1000; i++ {
			eg.Go(func() error {
				return r.Wait()
			})
		}
		err := eg.Wait()
		if err != nil {
			t.Errorf("should not return an error")
		}
		duration := time.Since(start)
		expectedDur := 497 * time.Millisecond // slightly below 500ms to prevent flaky-ness
		if duration < expectedDur {
			t.Errorf("test took %v, expected %v", duration, expectedDur)
		}
	})
	t.Run("sync", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			err := r.Wait()
			if err != nil {
				t.Errorf("should not return an error")
			}
		}
		duration := time.Since(start)
		expectedDur := 497 * time.Millisecond // slightly below 500ms to prevent flaky-ness
		if duration < expectedDur {
			t.Errorf("test took %v, expected %v", duration, expectedDur)
		}
	})
}

func TestRateLimiter_Wait_NoLimit(t *testing.T) {
	r := newMutexLimiter(config{
		durationBetweenCalls: NoLimit,
		maxQueue:             0,
	})
	start := time.Now()
	var eg errgroup.Group
	for i := 0; i < 1000; i++ {
		eg.Go(func() error {
			return r.Wait()
		})
	}
	err := eg.Wait()
	if err != nil {
		t.Errorf("should not return an error")
	}
	duration := time.Since(start)
	expectedMaxDur := 1 * time.Millisecond
	if duration > expectedMaxDur {
		t.Errorf("test took %v, expected < %v", duration, expectedMaxDur)
	}

}

func TestRateLimiter_Wait_QueueError(t *testing.T) {
	r := newMutexLimiter(config{
		durationBetweenCalls: time.Millisecond / 2,
		maxQueue:             10,
	})
	var eg errgroup.Group
	for i := 0; i < 20; i++ {
		eg.Go(func() error {
			return r.Wait()
		})
	}
	err := eg.Wait()
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestRateLimiter_Wait_QueueNoError(t *testing.T) {
	r := newMutexLimiter(config{
		durationBetweenCalls: time.Millisecond / 2,
		maxQueue:             20,
	})
	var eg errgroup.Group
	runtime.GOMAXPROCS(10)
	for i := 0; i < 19; i++ {
		eg.Go(func() error {
			return r.Wait()
		})
	}
	err := eg.Wait()
	if err != nil {
		t.Errorf("error not expected")
	}
}

func BenchmarkRateLimiter_Wait_NoLimiter(b *testing.B) {
	r := newMutexLimiter(config{
		durationBetweenCalls: NoLimit,
		maxQueue:             math.MaxUint32,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Wait()
	}
}

func BenchmarkRateLimiter_Wait_WithQueueLimit(b *testing.B) {
	r := newMutexLimiter(config{
		durationBetweenCalls: time.Nanosecond,
		maxQueue:             math.MaxUint32,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Wait()
	}
}

func BenchmarkRateLimiter_Wait_NoQueueLimit(b *testing.B) {
	r := newMutexLimiter(config{
		durationBetweenCalls: time.Nanosecond,
		maxQueue:             0,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Wait()
	}
}

// test case inspired by https://github.com/uber-go/ratelimit
func BenchmarkRateLimiter(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	for _, procs := range []int{16} {
		runtime.GOMAXPROCS(procs)
		for name, limiter := range map[string]RateLimiter{
			"mutex": newMutexLimiter(config{
				durationBetweenCalls: time.Nanosecond,
				maxQueue:             65536,
			}),
		} {
			for ng := 1; ng < 16; ng++ {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 16; ng < 128; ng += 8 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 128; ng < 512; ng += 16 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 512; ng < 1024; ng += 32 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 1024; ng < 2048; ng += 64 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 2048; ng < 4096; ng += 128 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 4096; ng < 16384; ng += 512 {
				runner(b, name, procs, ng, limiter)
			}
			for ng := 16384; ng < 65536; ng += 2048 {
				runner(b, name, procs, ng, limiter)
			}
		}
	}
}

func runner(b *testing.B, name string, procs int, ng int, limiter RateLimiter) bool {
	return b.Run(fmt.Sprintf("type:%s-procs:%d-goroutines:%d", name, procs, ng), func(b *testing.B) {
		var wg sync.WaitGroup
		var trigger int32
		atomic.StoreInt32(&trigger, 1)
		n := b.N
		batchSize := n / ng
		if batchSize == 0 {
			batchSize = n
		}
		for n > 0 {
			wg.Add(1)
			batch := min(n, batchSize)
			n -= batch
			go func(quota int) {
				for atomic.LoadInt32(&trigger) == 1 {
					// wait until trigger is switched to 1
					runtime.Gosched()
				}
				for i := 0; i < quota; i++ {
					_ = limiter.Wait()
				}
				wg.Done()
			}(batch)
		}

		b.StartTimer()
		atomic.CompareAndSwapInt32(&trigger, 1, 0)
		wg.Wait()
		b.StopTimer()
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
