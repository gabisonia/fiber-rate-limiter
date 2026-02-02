package strategies

import (
	"sync"
	"testing"
	"time"
)

// Sequential, single-user: allow up to Limit, then deny.
func TestSingleUserSequential_SlidingWindow(t *testing.T) {
	limit := 3
	window := 100 * time.Millisecond
	s := NewSlidingWindowStrategy(limit, window)
	client := "userA"

	// First `limit` requests should be allowed
	for i := 0; i < limit; i++ {
		if !s.IsRequestAllowed(client) {
			t.Errorf("seq %d: expected allowed, got denied", i+1)
		}
	}
	// Next one should be denied
	if s.IsRequestAllowed(client) {
		t.Error("seq over limit: expected denied, got allowed")
	}
}

// After window slides past the oldest timestamp, you can send again.
func TestSingleUserWindowReset_SlidingWindow(t *testing.T) {
	limit := 2
	window := 50 * time.Millisecond
	s := NewSlidingWindowStrategy(limit, window)
	client := "userA"

	// exhaust the limit
	for i := 0; i < limit; i++ {
		_ = s.IsRequestAllowed(client)
	}
	if s.IsRequestAllowed(client) {
		t.Fatal("should be denied immediately after hitting limit")
	}

	// wait for window to expire so the first timestamp drops out
	time.Sleep(window + 10*time.Millisecond)

	// now we should be allowed again up to `limit`
	for i := 0; i < limit; i++ {
		if !s.IsRequestAllowed(client) {
			t.Errorf("after reset %d: expected allowed", i+1)
		}
	}
}

// Concurrent calls for the same user should only succeed `limit` times.
func TestConcurrentSingleUser_SlidingWindow(t *testing.T) {
	limit := 50
	window := 100 * time.Millisecond
	s := NewSlidingWindowStrategy(limit, window)
	client := "userA"

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0

	// spin up 100 goroutines all calling IsRequestAllowed
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if s.IsRequestAllowed(client) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed != limit {
		t.Errorf("concurrent single user: expected %d allowed, got %d", limit, allowed)
	}
}

// Two users in parallel: each gets its own sliding window.
func TestConcurrentMultipleUsers_SlidingWindow(t *testing.T) {
	limit := 20
	window := 100 * time.Millisecond
	s := NewSlidingWindowStrategy(limit, window)

	users := []string{"userA", "userB"}
	var wg sync.WaitGroup
	results := make(map[string]int)
	var mu sync.Mutex

	for _, u := range users {
		// each user fires 2×limit goroutines
		for i := 0; i < 2*limit; i++ {
			wg.Add(1)
			go func(user string) {
				defer wg.Done()
				if s.IsRequestAllowed(user) {
					mu.Lock()
					results[user]++
					mu.Unlock()
				}
			}(u)
		}
	}
	wg.Wait()

	for _, u := range users {
		if got := results[u]; got != limit {
			t.Errorf("user %q: expected %d allowed, got %d", u, limit, got)
		}
	}
}

// The oldest timestamp should slide out, freeing capacity without waiting for a full window reset.
func TestSlidingWindowDropsOldest(t *testing.T) {
	limit := 3
	window := 60 * time.Millisecond
	s := NewSlidingWindowStrategy(limit, window)
	client := "userA"

	// fill the window
	for i := 0; i < limit; i++ {
		if !s.IsRequestAllowed(client) {
			t.Fatalf("prime %d: expected allowed", i+1)
		}
	}
	if s.IsRequestAllowed(client) {
		t.Fatal("extra request should be denied while window still full")
	}

	// wait just long enough for the first timestamp to age out
	time.Sleep(70 * time.Millisecond)

	if !s.IsRequestAllowed(client) {
		t.Fatal("after oldest dropped: expected allowed")
	}
}
