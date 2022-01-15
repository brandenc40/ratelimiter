package ratelimiter

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// compile time interface validation.
var _ RateLimiter = (*mutexLimiter)(nil)

func newMutexLimiter(config config) RateLimiter {
	return &mutexLimiter{
		mu:              sync.Mutex{},
		nextCall:        time.Now(),
		durBetweenCalls: config.durationBetweenCalls,
		maxQueue:        config.maxQueue,
		queued:          0,
	}
}

type mutexLimiter struct {
	mu sync.Mutex

	nextCall        time.Time
	durBetweenCalls time.Duration
	maxQueue        uint32
	queued          uint32
}

func (r *mutexLimiter) NumQueued() uint32 {
	return atomic.LoadUint32(&r.queued)
}

func (r *mutexLimiter) Wait() error {
	// check if not rate limited
	if r.durBetweenCalls <= 0 {
		return nil
	}

	// enqueue the request, error is returned if queue is full
	if err := r.enqueue(); err != nil {
		return err
	}

	// wait fur mutex
	r.mu.Lock()

	// remove from queue since the lock was acquired
	r.dequeue()

	now := time.Now()

	// get the duration still required
	// between now and the next call.
	sleepDuration := r.nextCall.Sub(now)
	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
		// if we had to wait for this call, add the wait duration as well as the
		// duration between calls to enable us to reuse the time.Now() above.
		r.nextCall = now.Add(r.durBetweenCalls + sleepDuration)
	} else {
		r.nextCall = now.Add(r.durBetweenCalls)
	}

	// free mutex for next call
	r.mu.Unlock()

	return nil
}

func (r *mutexLimiter) enqueue() error {
	// check if unlimited queue size
	if r.maxQueue == 0 {
		return nil
	}

	for swapped := false; !swapped; {
		// atomically read the current queue size
		curQueued := r.NumQueued()

		// if queue is full, remove from queue and return error
		if curQueued == math.MaxUint32 || curQueued >= r.maxQueue {
			return ErrQueueFull
		}

		// attempt to queue request, false will be returned if the value at &r.queued is
		// not equal to the expected curQueued value. If false is returned, the loop will
		// start over and attempt again.
		swapped = atomic.CompareAndSwapUint32(&r.queued, curQueued, curQueued+1)
	}

	return nil
}

func (r *mutexLimiter) dequeue() {
	// check if unlimited queue size
	if r.maxQueue == 0 {
		return
	}

	// decrement the queue count
	atomic.AddUint32(&r.queued, ^uint32(0))
}
