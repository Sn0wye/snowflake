package routes

import (
	"context"
	"net/http"

	"github.com/Sn0wye/snowflake/gold/pkg/health"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BindHealthRoutes(app *fiber.App, db *gorm.DB, rmq *messaging.MessagingService) {
	hs := health.NewService()
	hs.Register("database", health.NewDatabaseChecker(db))
	hs.Register("rabbitmq", health.FuncChecker(func(ctx context.Context) error {
		return rmq.HealthCheck()
	}))

	router := app.Group("/account")
	router.Get("/health", func(c *fiber.Ctx) error {
		status, components := hs.Check(context.Background())
		statusText := "healthy"
		if status == http.StatusServiceUnavailable {
			statusText = "unhealthy"
		}
		return c.Status(status).JSON(fiber.Map{
			"status":     statusText,
			"components": components,
		})
	})
}
