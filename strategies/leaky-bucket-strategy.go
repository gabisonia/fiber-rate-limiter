package strategies

import (
	"math"
	"sync"
	"time"
)

type LeakyBucketStrategy struct {
	LeakRate   float64
	BucketSize float64
	clients    map[string]*leakyBucketState
	mutex      sync.Mutex
}

type leakyBucketState struct {
	QueuedRequests float64
	LastLeak       time.Time
}

// NewLeakyBucketStrategy creates a new Leaky Bucket rate limiting strategy.
//
// Parameters:
//   - leakRate: number of requests that leak out (are processed) per second.
//   - bucketSize: maximum number of queued requests allowed in the bucket.
//
// Returns:
//   - *LeakyBucketStrategy: a pointer to a new instance of the strategy.
//
// This strategy allows requests to be queued up to bucketSize, and processes them
// at a steady leakRate. If the bucket is full, new requests are denied.
func NewLeakyBucketStrategy(leakRate, bucketSize float64) *LeakyBucketStrategy {
	return &LeakyBucketStrategy{
		LeakRate:   leakRate,
		BucketSize: bucketSize,
		clients:    make(map[string]*leakyBucketState),
	}
}

func (strategy *LeakyBucketStrategy) IsRequestAllowed(clientId string) bool {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	state, exists := strategy.clients[clientId]
	if !exists {
		state = &leakyBucketState{QueuedRequests: 0, LastLeak: now}
		strategy.clients[clientId] = state
	}

	elapsed := now.Sub(state.LastLeak).Seconds()
	leaked := elapsed * strategy.LeakRate
	state.QueuedRequests = math.Max(0, state.QueuedRequests-leaked)
	state.LastLeak = now

	if state.QueuedRequests < strategy.BucketSize {
		state.QueuedRequests++
		return true
	}

	return false
}
