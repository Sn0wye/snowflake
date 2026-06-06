package middleware

import (
	"sync"
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

type userRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*slidingWindow
	rate    int
}

type slidingWindow struct {
	timestamps []time.Time
}

func newUserRateLimiter(rate int) *userRateLimiter {
	return &userRateLimiter{
		buckets: make(map[string]*slidingWindow),
		rate:    rate,
	}
}

func (l *userRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	window := now.Add(-time.Second)

	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &slidingWindow{}
		l.buckets[key] = bucket
	}

	valid := bucket.timestamps[:0]
	for _, ts := range bucket.timestamps {
		if ts.After(window) {
			valid = append(valid, ts)
		}
	}
	bucket.timestamps = valid

	if len(valid) >= l.rate {
		return false
	}

	bucket.timestamps = append(bucket.timestamps, now)
	return true
}

func UserRateLimitMiddleware(conf *viper.Viper) fiber.Handler {
	if !conf.GetBool("rate_limit.enabled") {
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	rate := conf.GetInt("rate_limit.transactional")
	if rate <= 0 {
		rate = 10
	}

	limiter := newUserRateLimiter(rate)

	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*jwt.Claims)
		if !ok || claims.Subject == "" {
			return c.Next()
		}

		if !limiter.allow(claims.Subject) {
			c.Set("Retry-After", "1")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success":     false,
				"message":     "Too many requests. Try again in 1 second.",
				"retry_after": 1,
			})
		}
		return c.Next()
	}
}
