package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	config := RateLimiterConfig{
		RPS:     1,
		Burst:   1,
		TTL:     1 * time.Minute,
		Enabled: true,
	}
	limiterMiddleware := RateLimiter(config)

	// First request should be allowed
	err := limiterMiddleware(handler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Second request immediately after should be denied
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = limiterMiddleware(handler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Wait for the rate limiter to allow another request
	time.Sleep(1 * time.Second)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = limiterMiddleware(handler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiter_Disabled(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	config := RateLimiterConfig{
		Enabled: false,
	}
	limiterMiddleware := RateLimiter(config)

	// All requests should be allowed
	for i := 0; i < 5; i++ {
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		err := limiterMiddleware(handler)(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}
	config := RateLimiterConfig{
		RPS:     1000,
		Burst:   1000,
		TTL:     1 * time.Minute,
		Enabled: true,
	}
	limiterMiddleware := RateLimiter(config)
	mw := limiterMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			_ = mw(c)
		}
	})
}

