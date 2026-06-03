package routes

import (
	"github.com/getsnowflake/snowflake/helium/pkg/config"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
	"github.com/getsnowflake/snowflake/helium/pkg/service"
	"github.com/getsnowflake/snowflake/helium/src/controllers"
	"github.com/getsnowflake/snowflake/helium/src/db"

	"github.com/gofiber/fiber/v2"
)

func BindAuthRoutes(app *fiber.App, jwtMiddleware fiber.Handler, log *logger.Logger, rmq *messaging.MessagingService) {
	database := db.GetDB()
	conf := config.GetConfig()
	j := jwt.NewJwt(conf)

	repos := repository.NewFactory()
	services := service.NewFactory(repos, j, rmq)

	router := app.Group("/auth")

	authController := controllers.NewAuthController(database, j, services.Auth, log)

	router.Post("/login", authController.Login)
	router.Post("/register", authController.Register)
	router.Post("/refresh", authController.Refresh)
	router.Post("/logout", authController.Logout)

	router.Get("/profile", jwtMiddleware, authController.Profile)

	oauthController := controllers.NewOAuthController(database, j, conf, services.OAuth, log)

	router.Get("/oauth/google", oauthController.Redirect)
	router.Get("/oauth/google/callback", oauthController.Callback)
}
