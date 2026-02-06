package middleware

import (
	"fmt"
	"time"

	"snowflake/pkg/exceptions"
	"snowflake/pkg/jwt"
	"snowflake/pkg/logger"

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

		expirationTime := claims.ExpiresAt.Time
		if time.Until(expirationTime) < 5*time.Minute {
			newTokenString, err := j.GenToken(claims.Subject, time.Now().Add(time.Hour*24*90))
			if err != nil {
				fmt.Println("Error generating new token")
				return exceptions.InternalServer(ctx, err.Error())
			}
			ctx.Set("Authorization", "Bearer "+newTokenString)
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
