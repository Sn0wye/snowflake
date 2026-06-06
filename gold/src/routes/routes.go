package routes

import (
	"net/http"

	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/src/service"
	"github.com/Sn0wye/snowflake/gold/src/controllers"
	"github.com/Sn0wye/snowflake/gold/src/reconciliation"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func BindFlakeRoutes(app *fiber.App, db *gorm.DB, j *jwt.JWT, jwtMiddleware fiber.Handler, services *service.ServiceFactory) {
	router := app.Group("/account/flake", jwtMiddleware)

	controller := controllers.NewFlakeController(db, j, services.Flake)

	router.Post("/", controller.CreateFlake)
	router.Get("/", controller.GetFlakes)
	router.Delete("/:id", controller.DeleteFlake)

	app.Get("/account/flake/lookup", controller.PublicLookupFlake)
}

func BindBalanceRoutes(app *fiber.App, db *gorm.DB, j *jwt.JWT, jwtMiddleware fiber.Handler, services *service.ServiceFactory) {
	router := app.Group("/account/balance", jwtMiddleware)

	controller := controllers.NewBalanceController(db, j, services.Balance)

	router.Get("/", controller.GetBalance)
	router.Get("/history", controller.GetBalanceHistory)
}

func BindTransactionRoutes(app *fiber.App, db *gorm.DB, j *jwt.JWT, jwtMiddleware fiber.Handler, rateLimitMiddleware fiber.Handler, services *service.ServiceFactory) {
	router := app.Group("/account/transactions", jwtMiddleware, rateLimitMiddleware)

	controller := controllers.NewTransactionsController(db, j, services.Transaction)

	router.Post("/transfer", controller.CreateTransaction)
	router.Get("/", controller.GetTransactions)
	router.Get("/:id", controller.GetTransactionByID)
	router.Post("/deposit", controller.Deposit)
}

func BindAdminRoutes(app *fiber.App, jwtMiddleware fiber.Handler, rateLimitMiddleware fiber.Handler, log *logger.Logger, job *reconciliation.Job) {
	router := app.Group("/account/admin", jwtMiddleware, rateLimitMiddleware)

	router.Post("/reconcile", func(c *fiber.Ctx) error {
		go job.Run()
		return c.Status(http.StatusAccepted).JSON(fiber.Map{
			"message": "Reconciliation started",
		})
	})
}
