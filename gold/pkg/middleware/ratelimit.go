package middleware

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type limitEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

type userRateLimiter struct {
	entries sync.Map
	rateVal int
}

func newUserRateLimiter(r int) *userRateLimiter {
	l := &userRateLimiter{rateVal: r}
	go l.reap(1 * time.Minute)
	return l
}

func (l *userRateLimiter) allow(key string) bool {
	now := time.Now().UnixNano()

	v, _ := l.entries.LoadOrStore(key, &limitEntry{
		limiter: rate.NewLimiter(rate.Limit(l.rateVal), l.rateVal),
	})
	entry := v.(*limitEntry)
	entry.lastSeen.Store(now)

	return entry.limiter.Allow()
}

func (l *userRateLimiter) reap(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-interval).UnixNano()
		l.entries.Range(func(key, value any) bool {
			entry := value.(*limitEntry)
			if entry.lastSeen.Load() < cutoff {
				l.entries.Delete(key)
			}
			return true
		})
	}
}

func UserRateLimitMiddleware(conf *viper.Viper) fiber.Handler {
	if !conf.GetBool("rate_limit.enabled") {
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	r := conf.GetInt("rate_limit.transactional")
	if r <= 0 {
		r = 10
	}

	limiter := newUserRateLimiter(r)

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
