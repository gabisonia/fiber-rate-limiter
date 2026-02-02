package strategies

import (
	"sync"
	"time"
)

type SlidingWindowStrategy struct {
	Limit      int
	WindowSize time.Duration
	clients    map[string][]time.Time
	mutex      sync.Mutex
}

// NewSlidingWindowStrategy creates a new Sliding Window rate limiting strategy.
//
// Parameters:
//   - limit: maximum number of allowed requests within the sliding window.
//   - windowSize: duration of the sliding time window.
//
// Returns:
//   - *SlidingWindowStrategy: a pointer to a new instance of the strategy.
//
// This strategy limits requests based on the number of requests within a rolling
// time window. It allows more accurate rate limiting compared to fixed windows by
// continuously evaluating the request timestamps.
func NewSlidingWindowStrategy(limit int, windowSize time.Duration) *SlidingWindowStrategy {
	return &SlidingWindowStrategy{
		Limit:      limit,
		WindowSize: windowSize,
		clients:    make(map[string][]time.Time),
	}
}

func (strategy *SlidingWindowStrategy) IsRequestAllowed(clientId string) bool {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	timestamps, exists := strategy.clients[clientId]
	if !exists {
		timestamps = []time.Time{}
	}

	// Filter out old timestamps
	filtered := timestamps[:0]
	for _, t := range timestamps {
		if now.Sub(t) < strategy.WindowSize {
			filtered = append(filtered, t)
		}
	}
	timestamps = filtered

	if len(timestamps) < strategy.Limit {
		timestamps = append(timestamps, now)
		strategy.clients[clientId] = timestamps
		return true
	}

	strategy.clients[clientId] = timestamps
	return false
}

// RetryAfter returns how long until the oldest timestamp expires and capacity frees up.
func (strategy *SlidingWindowStrategy) RetryAfter(clientId string) time.Duration {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	timestamps := strategy.clients[clientId]
	if len(timestamps) == 0 {
		return 0
	}

	// refresh timestamps to drop stale entries
	filtered := timestamps[:0]
	for _, t := range timestamps {
		if now.Sub(t) < strategy.WindowSize {
			filtered = append(filtered, t)
		}
	}
	strategy.clients[clientId] = filtered

	if len(filtered) < strategy.Limit {
		return 0
	}

	oldest := filtered[0]
	wait := oldest.Add(strategy.WindowSize).Sub(now)
	if wait < 0 {
		return 0
	}
	return wait
}
