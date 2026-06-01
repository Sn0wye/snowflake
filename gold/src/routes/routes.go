package routes

import (
	"net/http"

	"github.com/Sn0wye/snowflake/gold/pkg/config"
	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/pkg/service"
	"github.com/Sn0wye/snowflake/gold/src/controllers"
	"github.com/Sn0wye/snowflake/gold/src/db"
	"github.com/Sn0wye/snowflake/gold/src/reconciliation"

	"github.com/gofiber/fiber/v2"
)

func BindFlakeRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService, services *service.ServiceFactory) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/flake", jwtMiddleware)

	controller := controllers.NewFlakeController(db, jwt, services.Flake)

	router.Post("/", controller.CreateFlake)
	router.Get("/", controller.GetFlakes)
	router.Delete("/:id", controller.DeleteFlake)

	app.Get("/account/flake/lookup", controller.PublicLookupFlake)
}

func BindBalanceRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService, services *service.ServiceFactory) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/balance", jwtMiddleware)

	controller := controllers.NewBalanceController(db, jwt, services.Balance)

	router.Get("/", controller.GetBalance)
	router.Get("/history", controller.GetBalanceHistory)
}

func BindTransactionRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService, services *service.ServiceFactory) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/account/transactions", jwtMiddleware)

	controller := controllers.NewTransactionsController(db, jwt, services.Transaction)

	router.Post("/transfer", controller.CreateTransaction)
	router.Get("/", controller.GetTransactions)
	router.Get("/:id", controller.GetTransactionByID)
	router.Post("/deposit", controller.Deposit)
}

func BindAdminRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, job *reconciliation.Job) {
	router := app.Group("/account/admin", jwtMiddleware)

	router.Post("/reconcile", func(c *fiber.Ctx) error {
		go job.Run()
		return c.Status(http.StatusAccepted).JSON(fiber.Map{
			"message": "Reconciliation started",
		})
	})
}
