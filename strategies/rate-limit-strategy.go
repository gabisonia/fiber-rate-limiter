package strategies

import "time"

type RateLimitStrategy interface {
	// IsRequestAllowed returns true if the request should proceed.
	IsRequestAllowed(clientId string) bool
	// RetryAfter returns how long a caller should wait before retrying.
	// If zero, the request can be retried immediately.
	RetryAfter(clientId string) time.Duration
}
