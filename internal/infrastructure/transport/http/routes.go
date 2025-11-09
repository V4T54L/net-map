package http

import (
	"github.com/labstack/echo/v4"

	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/transport/http/middleware"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/token"
)

func RegisterRoutes(
	e *echo.Echo, // Changed from *echo.Group to *echo.Echo
	authUC usecase.AuthUseCase,
	userUC usecase.UserUseCase,
	dnsUC usecase.DNSRecordUseCase, // Added from attempted
	userRepo repository.UserRepository,
	tokenGenerator token.Generator,
) {
	// Handlers
	authHandler := NewAuthHandler(authUC)
	userHandler := NewUserHandler(userUC)
	dnsHandler := NewDNSRecordHandler(dnsUC) // Added from attempted

	// Middleware
	jwtMiddleware := middleware.NewJWTMiddleware(tokenGenerator, userRepo)

	// Group for v1 API
	v1 := e.Group("/api/v1") // Created v1 group from main Echo instance

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// Admin routes
	adminGroup := v1.Group("/admin", jwtMiddleware.Auth(domain.RoleAdmin)) // Middleware applied directly to group
	adminGroup.GET("/users", userHandler.ListUsers)
	adminGroup.GET("/users/:id", userHandler.GetUser)
	adminGroup.PATCH("/users/:id/status", userHandler.UpdateUserStatus)

	// DNS Record routes (for authenticated users)
	dnsGroup := v1.Group("/dns-records", jwtMiddleware.Auth(domain.RoleUser, domain.RoleAdmin)) // Added from attempted
	dnsGroup.POST("", dnsHandler.CreateRecord)
	dnsGroup.GET("", dnsHandler.ListRecords)
	dnsGroup.GET("/:id", dnsHandler.GetRecord)
	dnsGroup.PUT("/:id", dnsHandler.UpdateRecord)
	dnsGroup.DELETE("/:id", dnsHandler.DeleteRecord)
}
