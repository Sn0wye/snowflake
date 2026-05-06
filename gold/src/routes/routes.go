package routes

import (
	"net/http"

	"github.com/Sn0wye/snowflake/gold/pkg/config"
	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/src/controllers"
	"github.com/Sn0wye/snowflake/gold/src/db"
	"github.com/Sn0wye/snowflake/gold/src/reconciliation"

	"github.com/gofiber/fiber/v2"
)

func BindFlakeRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/flake", jwtMiddleware)

	controller := controllers.NewFlakeController(db, jwt)

	router.Post("/", controller.CreateFlake)
	router.Get("/", controller.GetFlakes)
	router.Delete("/:id", controller.DeleteFlake)
	router.Get("/lookup", controller.PublicLookupFlake)
}

func BindBalanceRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/balance", jwtMiddleware)

	controller := controllers.NewBalanceController(db, jwt)

	router.Get("/", controller.GetBalance)
	router.Get("/history", controller.GetBalanceHistory)
}

func BindTransactionRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/transactions", jwtMiddleware)

	controller := controllers.NewTransactionsController(db, jwt, rmq, log)

	router.Post("/transfer", controller.CreateTransaction)
	router.Get("/", controller.GetTransactions)
	router.Get("/:id", controller.GetTransactionByID)
	router.Post("/deposit", controller.Deposit)
}

func BindHealthRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		database := db.GetDB()
		sqlDB, err := database.DB()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "unavailable",
			})
		}
		if err := sqlDB.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "unreachable",
			})
		}
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":   "healthy",
			"database": "ok",
		})
	})
}

func BindAdminRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, job *reconciliation.Job) {
	router := app.Group("/account/admin", jwtMiddleware)

	// On-demand reconciliation trigger (JWT required; no admin-role check yet)
	router.Post("/reconcile", func(c *fiber.Ctx) error {
		go job.Run()
		return c.Status(http.StatusAccepted).JSON(fiber.Map{
			"message": "Reconciliation started",
		})
	})
}
