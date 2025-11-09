package http

import (
	"internal-dns/internal/usecase"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes sets up the HTTP routes for the application.
func RegisterRoutes(g *echo.Group, authUC usecase.AuthUseCase) {
	userHandler := NewUserHandler(authUC)

	authGroup := g.Group("/auth")
	authGroup.POST("/register", userHandler.Register)
	authGroup.POST("/login", userHandler.Login)
}
```
```go
