package middleware

import (
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
			logger.Warn("no token provided")
			return exceptions.Unauthorized(ctx)
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.Warn("invalid token", zap.Error(err))
			return exceptions.Unauthorized(ctx)
		}

		ctx.Locals("claims", claims)
		setUserLogContext(ctx, logger)
		return ctx.Next()
	}
}

func setUserLogContext(ctx *fiber.Ctx, logger *logger.Logger) {
	userInfo := ctx.Locals("claims").(*jwt.Claims)
	logger.NewContext(ctx, zap.String("UserId", userInfo.Subject))
}
