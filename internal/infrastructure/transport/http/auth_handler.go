package http

import (
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authUC usecase.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
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
func (h *AuthHandler) Register(c echo.Context) error {
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Basic validation, more complex validation can be added
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
	}

	err := h.authUC.Register(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		if err == domain.ErrUsernameTooShort || err == domain.ErrPasswordTooShort {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Login handles user login.
func (h *AuthHandler) Login(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	accessToken, refreshToken, err := h.authUC.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to login"})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

