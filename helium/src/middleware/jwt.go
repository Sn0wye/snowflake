package middleware

import (
	"fmt"

	"github.com/getsnowflake/snowflake/helium/pkg/exceptions"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func JWTMiddleware(conf *viper.Viper, logger *logger.Logger) fiber.Handler {
	j := jwt.NewJwt(conf)
	return func(ctx *fiber.Ctx) error {
		tokenString := ctx.Get("Authorization")
		if tokenString == "" {
			fmt.Println("No token provided")
			return exceptions.Unauthorized(ctx)
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			fmt.Println("Invalid token provided")
			return exceptions.Unauthorized(ctx)
		}

		ctx.Locals("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		return ctx.Next()
	}
}

func recoveryLoggerFunc(ctx *fiber.Ctx, logger *logger.Logger) {
	userInfo := ctx.Locals("claims").(*jwt.Claims)
	logger.NewContext(ctx, zap.String("UserId", userInfo.Subject))
}
