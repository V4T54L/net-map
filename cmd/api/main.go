package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"internal-dns/internal/infrastructure/database"
	"internal-dns/internal/infrastructure/transport/http"
	"internal-dns/internal/repository"
	"internal-dns/internal/service"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/token"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database connection
	dbpool, err := database.NewPostgresPool(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	// Initialize components
	userRepo := database.NewUserPostgresRepository(dbpool)
	tokenGenerator := token.NewJWTGenerator(os.Getenv("JWT_SECRET_KEY"))
	authService := service.NewAuthService(userRepo, tokenGenerator)

	// Setup Echo server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Setup routes
	apiGroup := e.Group("/api/v1")
	http.RegisterRoutes(apiGroup, authService)

	// Start server
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
```
```go
