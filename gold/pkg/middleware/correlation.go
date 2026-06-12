package middleware

import (
	"regexp"

	"github.com/getsnowflake/snowflake/gold/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const CorrelationIDKey = "correlation_id"

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func validCorrelationID(id string) bool {
	return uuidV4Re.MatchString(id)
}

func CorrelationMiddleware(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Correlation-ID")
		if !validCorrelationID(id) {
			id = uuid.New().String()
		}
		c.Locals(CorrelationIDKey, id)
		c.Set("X-Correlation-ID", id)
		log.NewContext(c, zap.String("correlation_id", id))
		return c.Next()
	}
}

func GetCorrelationID(c *fiber.Ctx) string {
	id, _ := c.Locals(CorrelationIDKey).(string)
	return id
}
