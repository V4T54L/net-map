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
	"internal-dns/internal/util" // Added from attempted
	"internal-dns/pkg/token"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Database connection
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	connString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	dbPool, err := database.NewPostgresPool(context.Background(), connString)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established")

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET_KEY must be set")
	}

	// Dependency Injection
	userRepo := database.NewUserPostgresRepository(dbPool)
	tokenGenerator := token.NewJWTGenerator(jwtSecret)
	authService := service.NewAuthService(userRepo, tokenGenerator)
	userService := service.NewUserService(userRepo) // Added from attempted

	// Setup Echo HTTP server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	apiGroup := e.Group("/api/v1")
	http.RegisterRoutes(apiGroup, authService, userService, userRepo, tokenGenerator) // Updated signature

	// Start server
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		e.Logger.Fatal(err)
	}
}
