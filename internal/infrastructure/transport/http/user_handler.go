package http

import (
	"net/http"

	"internal-dns/internal/usecase"

	"github.com/labstack/echo/v4"
)

// UserHandler handles HTTP requests related to users.
type UserHandler struct {
	authUC usecase.AuthUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(authUC usecase.AuthUseCase) *UserHandler {
	return &UserHandler{authUC: authUC}
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register handles user registration.
func (h *UserHandler) Register(c echo.Context) error {
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Basic validation
	if len(req.Username) < 3 || len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username must be at least 3 characters and password at least 8 characters"})
	}

	err := h.authUC.Register(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		// This should be more granular in a real app (e.g., check for duplicate username)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Login handles user login.
func (h *UserHandler) Login(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	accessToken, refreshToken, err := h.authUC.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
```
```go
