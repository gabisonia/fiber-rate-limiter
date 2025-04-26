package strategies

import (
	"sync"
	"testing"
	"time"
)

// 1) Sequential single‐user: leakRate=0 → first allowed, second denied.
func TestSingleUserSequential_LeakyBucket(t *testing.T) {
	bucketSize := 1.0
	leakRate := 0.0
	s := NewLeakyBucketStrategy(leakRate, bucketSize)
	client := "userA"

	if !s.IsRequestAllowed(client) {
		t.Error("first request: expected allowed, got denied")
	}
	if s.IsRequestAllowed(client) {
		t.Error("second request: expected denied, got allowed")
	}
}

// 2) Reset after exactly one window: fill once, wait, then allow again.
func TestSingleUserLeakReset_LeakyBucket(t *testing.T) {
	bucketSize := 1.0
	window := 100 * time.Millisecond
	leakRate := bucketSize / window.Seconds() // ensures full drain in one window
	s := NewLeakyBucketStrategy(leakRate, bucketSize)
	client := "userA"

	// 1) fill the bucket
	if !s.IsRequestAllowed(client) {
		t.Fatal("initial request: expected allowed")
	}

	// 2) wait one full window + tiny jitter so it leaks completely
	time.Sleep(window + 10*time.Millisecond)

	// 3) now bucket should be empty → allow again
	if !s.IsRequestAllowed(client) {
		t.Error("after leak reset: expected allowed, got denied")
	}
}

// 3) Concurrent single-user: leakRate=0 → exactly bucketSize successes out of 100.
func TestConcurrentSingleUser_LeakyBucket(t *testing.T) {
	bucketSize := 50.0
	leakRate := 0.0
	s := NewLeakyBucketStrategy(leakRate, bucketSize)
	client := "userA"

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0

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

	if float64(allowed) != bucketSize {
		t.Errorf("concurrent single user: expected %v allowed, got %v", bucketSize, allowed)
	}
}

// 4) Concurrent multiple-users: leakRate=0 → each user gets bucketSize successes.
func TestConcurrentMultipleUsers_LeakyBucket(t *testing.T) {
	bucketSize := 20.0
	leakRate := 0.0
	s := NewLeakyBucketStrategy(leakRate, bucketSize)

	users := []string{"userA", "userB"}
	var wg sync.WaitGroup
	results := map[string]float64{"alice": 0, "bob": 0}
	var mu sync.Mutex

	for _, u := range users {
		for i := 0; i < int(bucketSize*2); i++ {
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
		if results[u] != bucketSize {
			t.Errorf("user %q: expected %v allowed, got %v", u, bucketSize, results[u])
		}
	}
}
