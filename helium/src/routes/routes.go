package routes

import (
	"github.com/getsnowflake/snowflake/helium/pkg/config"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/src/controllers"
	"github.com/getsnowflake/snowflake/helium/src/db"

	"github.com/gofiber/fiber/v2"
)

func BindAuthRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	db := db.GetDB()
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	router := app.Group("/auth")

	controller := controllers.NewAuthController(db, jwt, rmq, log)

	router.Post("/login", controller.Login)
	router.Post("/register", controller.Register)
	router.Post("/refresh", controller.Refresh)
	router.Post("/logout", controller.Logout)

	router.Get("/profile", jwtMiddleware, controller.Profile)

	oauthController := controllers.NewOAuthController(db, jwt, rmq, conf, log)

	router.Get("/oauth/google", oauthController.Redirect)
	router.Get("/oauth/google/callback", oauthController.Callback)
}
