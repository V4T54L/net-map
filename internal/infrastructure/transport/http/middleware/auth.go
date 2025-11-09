package middleware

import (
	"context"
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/pkg/token"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type contextKey string

const UserContextKey = contextKey("user")

type JWTMiddleware struct {
	tokenGenerator token.Generator
	userRepo       repository.UserRepository
}

func NewJWTMiddleware(tg token.Generator, ur repository.UserRepository) *JWTMiddleware {
	return &JWTMiddleware{
		tokenGenerator: tg,
		userRepo:       ur,
	}
}

func (m *JWTMiddleware) Auth(requiredRoles ...domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
			}
			tokenString := parts[1]

			claims, err := m.tokenGenerator.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
			}

			// Check if user exists and is enabled
			user, err := m.userRepo.FindByID(c.Request().Context(), claims.UserID)
			if err != nil || !user.IsEnabled {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found or disabled"})
			}

			// Check role
			if len(requiredRoles) > 0 {
				isAllowed := false
				for _, role := range requiredRoles {
					if user.Role == role {
						isAllowed = true
						break
					}
				}
				if !isAllowed {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
				}
			}

			// Store user in context
			ctx := context.WithValue(c.Request().Context(), UserContextKey, user)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

