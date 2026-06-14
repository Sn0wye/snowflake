// Package metrics wires Prometheus RED (rate, errors, duration) and Go runtime
// metrics into a Fiber app and exposes them at GET /metrics.
package metrics

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

// Register attaches the Prometheus middleware to app and serves the collected
// metrics (HTTP request rate, error rate, latency histograms plus Go runtime
// metrics) at /metrics on the app's existing port.
//
// Must be called before binding routes so every handler is instrumented.
func Register(app *fiber.App, serviceName string) {
	fp := fiberprometheus.New(serviceName)
	fp.RegisterAt(app, "/metrics")
	app.Use(fp.Middleware)
}
