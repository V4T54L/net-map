package http

import (
	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/transport/http/middleware"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/token"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(
	g *echo.Group,
	authUC usecase.AuthUseCase,
	userUC usecase.UserUseCase,
	userRepo repository.UserRepository,
	tokenGenerator token.Generator,
) {
	// Handlers
	authHandler := NewAuthHandler(authUC) // NewAuthHandler is assumed to be created elsewhere
	userHandler := NewUserHandler(userUC)

	// Middleware
	jwtMiddleware := middleware.NewJWTMiddleware(tokenGenerator, userRepo)

	// Public routes
	authGroup := g.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// Admin routes
	adminGroup := g.Group("/admin")
	adminGroup.Use(jwtMiddleware.Auth(domain.RoleAdmin))
	adminGroup.GET("/users", userHandler.ListUsers)
	adminGroup.GET("/users/:id", userHandler.GetUser)
	adminGroup.PATCH("/users/:id/status", userHandler.UpdateUserStatus)
}
