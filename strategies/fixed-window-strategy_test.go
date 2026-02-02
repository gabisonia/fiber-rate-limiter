package strategies

import (
	"sync"
	"testing"
	"time"
)

// Test sequential single-user behavior: allow up to limit, then deny.
func TestSingleUserSequential(t *testing.T) {
	limit := 3
	window := 100 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)

	client := "userA"

	// First `limit` requests should be allowed
	for i := 0; i < limit; i++ {
		if !s.IsRequestAllowed(client) {
			t.Errorf("request %d: expected allowed, got denied", i+1)
		}
	}

	// Next request should be denied
	if s.IsRequestAllowed(client) {
		t.Error("request over limit: expected denied, got allowed")
	}
}

// Test that after the window expires, the counter resets.
func TestSingleUserWindowReset(t *testing.T) {
	limit := 2
	window := 50 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)

	client := "userA"
	// exhaust the limit
	for i := 0; i < limit; i++ {
		_ = s.IsRequestAllowed(client)
	}
	if s.IsRequestAllowed(client) {
		t.Fatal("should be denied immediately after hitting limit")
	}

	// wait for window to roll over
	time.Sleep(window + 10*time.Millisecond)

	// now we should be allowed again up to `limit`
	for i := 0; i < limit; i++ {
		if !s.IsRequestAllowed(client) {
			t.Errorf("after reset, request %d: expected allowed", i+1)
		}
	}
}

// Test concurrent usage: many goroutines for the same user.
func TestConcurrentSingleUser(t *testing.T) {
	limit := 50
	window := 100 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)

	const client = "userA"
	var wg sync.WaitGroup
	allowed := 0
	var mu sync.Mutex

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

	// only `limit` of them should have succeeded
	if allowed != limit {
		t.Errorf("concurrent single user: expected %d allowed, got %d", limit, allowed)
	}
}

// Test concurrent usage: two different users in parallel.
func TestConcurrentMultipleUsers(t *testing.T) {
	limit := 20
	window := 100 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)

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

// Ensure the counter rolls over cleanly across multiple consecutive windows.
func TestMultipleWindowRollovers(t *testing.T) {
	limit := 2
	window := 30 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)
	client := "userA"

	for cycle := 0; cycle < 3; cycle++ {
		for i := 0; i < limit; i++ {
			if !s.IsRequestAllowed(client) {
				t.Fatalf("cycle %d request %d: expected allowed", cycle+1, i+1)
			}
		}
		if s.IsRequestAllowed(client) {
			t.Fatalf("cycle %d over limit: expected denied", cycle+1)
		}
		time.Sleep(window + 5*time.Millisecond)
	}
}

// RetryAfter should indicate remaining time in the current window.
func TestRetryAfter_FixedWindow(t *testing.T) {
	limit := 1
	window := 40 * time.Millisecond
	s := NewFixedWindowStrategy(limit, window)
	client := "userA"

	if !s.IsRequestAllowed(client) {
		t.Fatal("first request should be allowed")
	}
	if s.RetryAfter(client) <= 0 {
		t.Fatal("expected positive retry-after when over limit")
	}
	time.Sleep(window + 5*time.Millisecond)
	if s.RetryAfter(client) != 0 {
		t.Fatal("expected zero retry-after after window elapsed")
	}
}
