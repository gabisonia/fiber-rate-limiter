package strategies

import (
	"math"
	"sync"
	"testing"
	"time"
)

// Sequential, single-user: consume up to bucket size, then deny.
func TestSingleUserSequential_TokenBucket(t *testing.T) {
	bucketSize := 2.0
	refillRate := 10.0 // tokens per second
	s := NewTokenBucketStrategy(refillRate, bucketSize)
	client := "userA"

	// First `bucketSize` requests allowed
	for i := 0; i < int(bucketSize); i++ {
		if !s.IsRequestAllowed(client) {
			t.Errorf("request %d: expected allowed, got denied", i+1)
		}
	}
	// Next one should be denied
	if s.IsRequestAllowed(client) {
		t.Error("over bucket size: expected denied, got allowed")
	}
}

// After tokens are drained, wait for refill and verify one comes back.
func TestSingleUserRefill_TokenBucket(t *testing.T) {
	bucketSize := 2.0
	refillRate := 10.0 // tokens per second
	s := NewTokenBucketStrategy(refillRate, bucketSize)
	client := "userA"

	// drain the bucket
	for i := 0; i < int(bucketSize); i++ {
		_ = s.IsRequestAllowed(client)
	}
	if s.IsRequestAllowed(client) {
		t.Fatal("should have been denied after draining bucket")
	}

	// wait ~150ms → should refill ~1.5 tokens
	time.Sleep(150 * time.Millisecond)

	// one request allowed (needs ≥1 token)
	if !s.IsRequestAllowed(client) {
		t.Error("after refill: expected allowed, got denied")
	}
	// and immediately after, only ~0.5 tokens remain → deny
	if s.IsRequestAllowed(client) {
		t.Error("after consuming refill: expected denied, got allowed")
	}
}

// Concurrent calls for the same user should only succeed `bucketSize` times.
func TestConcurrentSingleUser_TokenBucket(t *testing.T) {
	bucketSize := 50.0
	refillRate := 1000.0 // high refill, but calls happen near-instantly so refill ≈0
	s := NewTokenBucketStrategy(refillRate, bucketSize)
	client := "userA"

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0

	// fire 100 goroutines
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
		t.Errorf("expected %v allowed, got %v", bucketSize, allowed)
	}
}

// Two users in parallel: each gets their own bucket.
func TestConcurrentMultipleUsers_TokenBucket(t *testing.T) {
	bucketSize := 20.0
	refillRate := 1000.0
	s := NewTokenBucketStrategy(refillRate, bucketSize)

	users := []string{"userA", "userB"}
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := map[string]float64{"alice": 0, "bob": 0}

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
		if math.Abs(results[u]-bucketSize) > 0.0001 {
			t.Errorf("user %q: expected %v allowed, got %v", u, bucketSize, results[u])
		}
	}
}
