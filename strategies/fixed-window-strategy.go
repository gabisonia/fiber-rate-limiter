package strategies

import (
	"sync"
	"time"
)

type FixedWindowStrategy struct {
	Limit      int
	WindowSize time.Duration
	clients    map[string]*fixedWindowState
	mutex      sync.Mutex
}

type fixedWindowState struct {
	WindowStart  time.Time
	RequestCount int
}

// NewFixedWindowStrategy creates a new Fixed Window rate limiting strategy.
//
// Parameters:
//   - limit: maximum number of allowed requests per window.
//   - windowSize: duration of the fixed time window.
//
// Returns:
//   - *FixedWindowStrategy: a pointer to a new instance of the strategy.
//
// This strategy limits requests by counting them within fixed, non-overlapping time windows.
func NewFixedWindowStrategy(limit int, windowSize time.Duration) *FixedWindowStrategy {
	return &FixedWindowStrategy{
		Limit:      limit,
		WindowSize: windowSize,
		clients:    make(map[string]*fixedWindowState),
	}
}

func (strategy *FixedWindowStrategy) IsRequestAllowed(clientId string) bool {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	state, exists := strategy.clients[clientId]
	if !exists {
		state = &fixedWindowState{WindowStart: now, RequestCount: 0}
		strategy.clients[clientId] = state
	}

	if now.After(state.WindowStart.Add(strategy.WindowSize)) {
		state.WindowStart = now
		state.RequestCount = 0
	}

	if state.RequestCount < strategy.Limit {
		state.RequestCount++
		return true
	}

	return false
}

// RetryAfter returns the remaining time in the current window before another
// request would be allowed.
func (strategy *FixedWindowStrategy) RetryAfter(clientId string) time.Duration {
	now := time.Now()
	strategy.mutex.Lock()
	defer strategy.mutex.Unlock()

	state, exists := strategy.clients[clientId]
	if !exists {
		return 0
	}

	// If the window has already rolled, allow immediately.
	windowEnd := state.WindowStart.Add(strategy.WindowSize)
	if now.After(windowEnd) {
		return 0
	}

	return windowEnd.Sub(now)
}
