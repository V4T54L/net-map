package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	RPS     float64
	Burst   int
	TTL     time.Duration
	Enabled bool
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func RateLimiter(config RateLimiterConfig) echo.MiddlewareFunc {
	if !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	// Periodically clean up old visitors
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > config.TTL {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(rate.Limit(config.RPS), config.Burst)
				v = &visitor{limiter: limiter}
				visitors[ip] = v
			}
			v.lastSeen = time.Now()
			mu.Unlock()

			if !v.limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"message": "Too many requests"})
			}

			return next(c)
		}
	}
}

