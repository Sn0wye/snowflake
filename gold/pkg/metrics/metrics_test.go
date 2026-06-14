package metrics_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsnowflake/snowflake/gold/pkg/metrics"
	"github.com/gofiber/fiber/v2"
)

func TestRegisterExposesMetricsEndpoint(t *testing.T) {
	app := fiber.New()
	metrics.Register(app, "gold")

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Generate a request so RED metrics have data.
	if _, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ping", nil)); err != nil {
		t.Fatalf("ping request: %v", err)
	}

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/metrics", nil))
	if err != nil {
		t.Fatalf("metrics request: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	got := string(body)

	// Request rate / error rate counter (carries method, status_code, path labels).
	if !strings.Contains(got, "http_requests_total") {
		t.Fatalf("expected http_requests_total in metrics, got:\n%s", got)
	}
	// Latency histogram.
	if !strings.Contains(got, "http_request_duration_seconds") {
		t.Fatalf("expected http_request_duration_seconds histogram in metrics, got:\n%s", got)
	}
}
