package middleware

import (
	"github.com/gabisonia/fiber-rate-limiter/strategies"
	"github.com/gofiber/fiber/v2"
)

// RateLimitingMiddleware creates a Fiber middleware that applies rate limiting
// using the provided strategy and client ID resolver function.
//
// Parameters:
//   - strategy: RateLimitStrategy that defines how rate limits are enforced.
//   - clientIdResolver: function to extract a unique client ID from the request.
//
// Returns:
//   - fiber.Handler: the middleware function that checks rate limits.
//
// If the client exceeds the allowed rate, the middleware responds with HTTP 429.
// Otherwise, it passes the request to the next handler.
func RateLimitingMiddleware(strategy strategies.RateLimitStrategy, clientIdResolver func(*fiber.Ctx) string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientId := clientIdResolver(c)

		if !strategy.IsRequestAllowed(clientId) {
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded.")
		}

		return c.Next()
	}
}
