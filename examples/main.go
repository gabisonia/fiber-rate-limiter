package main

import (
	"github.com/gabisonia/fiber-rate-limiter/middleware"
	"github.com/gabisonia/fiber-rate-limiter/strategies"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

func main() {
	app := fiber.New()

	// Create a strategy (e.g., Token Bucket)
	strategy := strategies.NewSlidingWindowStrategy(2, time.Minute)

	// Create middleware
	limiter := middleware.RateLimitingMiddleware(strategy, clientIdResolver)

	// Global usage
	app.Use(limiter)

	// Route-specific usage
	app.Get("/limited", limiter, func(c *fiber.Ctx) error {
		return c.SendString("This route is rate-limited per endpoint.")
	})

	// Manual usage
	app.Get("/manual", func(c *fiber.Ctx) error {
		clientId := c.IP()

		// Manually apply rate limit logic
		if !strategy.IsRequestAllowed(clientId) {
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded (manual).")
		}

		return c.SendString("Manual rate-limited response.")
	})

	log.Fatal(app.Listen(":3000"))
}

func clientIdResolver(c *fiber.Ctx) string {
	apiKey := c.Get("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Fallback to just IP if API Key is missing
	return c.IP()
}
