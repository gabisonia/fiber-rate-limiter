package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// fakeStrategy allows deterministic allowance and retry delays.
type fakeStrategy struct {
	allow bool
	wait  time.Duration
}

func (f fakeStrategy) IsRequestAllowed(clientId string) bool { return f.allow }
func (f fakeStrategy) RetryAfter(clientId string) time.Duration {
	return f.wait
}

func TestMiddlewareSetsRetryAfterOnLimit(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimitingMiddleware(fakeStrategy{allow: false, wait: 2 * time.Second}, func(*fiber.Ctx) string { return "client" }))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "2" {
		t.Fatalf("expected Retry-After=2, got %q", got)
	}
}

func TestMiddlewareOmitsRetryAfterWhenZero(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimitingMiddleware(fakeStrategy{allow: false, wait: 0}, func(*fiber.Ctx) string { return "client" }))

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "" {
		t.Fatalf("expected no Retry-After header, got %q", got)
	}
}

func TestMiddlewarePassesThroughWhenAllowed(t *testing.T) {
	app := fiber.New()
	app.Use(RateLimitingMiddleware(fakeStrategy{allow: true, wait: 0}, func(*fiber.Ctx) string { return "client" }))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get(fiber.HeaderRetryAfter); got != "" {
		t.Fatalf("expected no Retry-After header, got %q", got)
	}
}
