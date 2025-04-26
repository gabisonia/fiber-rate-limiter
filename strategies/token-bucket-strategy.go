package strategies

import (
	"math"
	"sync"
	"time"
)

type TokenBucketStrategy struct {
	RefillRate float64
	BucketSize float64
	clients    map[string]*tokenBucketState
	mutex      sync.Mutex
}

type tokenBucketState struct {
	Tokens     float64
	LastRefill time.Time
}

// NewTokenBucketStrategy creates a new Token Bucket rate limiting strategy.
//
// Parameters:
//   - refillRate: number of tokens added to the bucket per second.
//   - bucketSize: maximum number of tokens the bucket can hold.
//
// Returns:
//   - *TokenBucketStrategy: a pointer to a new instance of the strategy.
//
// This strategy allows bursts of requests up to bucketSize and then refills tokens
// at a steady rate defined by refillRate. A request consumes one token.
// If no tokens are available, the request is denied.
func NewTokenBucketStrategy(refillRate, bucketSize float64) *TokenBucketStrategy {
	return &TokenBucketStrategy{
		RefillRate: refillRate,
		BucketSize: bucketSize,
		clients:    make(map[string]*tokenBucketState),
	}
}

func (strategy *TokenBucketStrategy) IsRequestAllowed(clientId string) bool {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	state, exists := strategy.clients[clientId]
	if !exists {
		state = &tokenBucketState{Tokens: strategy.BucketSize, LastRefill: now}
		strategy.clients[clientId] = state
	}

	elapsed := now.Sub(state.LastRefill).Seconds()
	state.Tokens = math.Min(strategy.BucketSize, state.Tokens+(elapsed*strategy.RefillRate))
	state.LastRefill = now

	if state.Tokens >= 1 {
		state.Tokens--
		return true
	}

	return false
}
