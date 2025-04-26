package strategies

type RateLimitStrategy interface {
	IsRequestAllowed(clientId string) bool
}
