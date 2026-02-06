package routes

import (
	"snowflake/pkg/config"
	"snowflake/pkg/jwt"
	"snowflake/pkg/logger"
	"snowflake/pkg/messaging"
	"snowflake/src/controllers"
	"snowflake/src/db"

	"github.com/gofiber/fiber/v2"
)

func BindAuthRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/auth")

	controller := controllers.NewAuthController(db, jwt, rmq)

	router.Post("/login", controller.Login)
	router.Post("/register", controller.Register)

	router.Get("/profile", jwtMiddleware, controller.Profile)
}
