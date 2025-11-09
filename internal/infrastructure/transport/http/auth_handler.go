package http

import (
	"errors"
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authUC usecase.AuthUseCase
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"` // Changed min=8 to min=6
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`  // Changed to camelCase
	RefreshToken string `json:"refreshToken"` // Changed to camelCase
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Registration Info"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Basic validation, more complex validation can be added
	if len(req.Username) < 3 || len(req.Password) < 6 { // Updated password length check
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username must be at least 3 characters and password at least 6 characters"})
	}

	err := h.authUC.Register(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) { // Using errors.Is
			return c.JSON(http.StatusConflict, map[string]string{"error": "Username already exists"}) // Refined error message
		}
		if errors.Is(err, domain.ErrUsernameTooShort) || errors.Is(err, domain.ErrPasswordTooShort) { // Using errors.Is
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Login godoc
// @Summary Log in a user
// @Description Authenticates a user and returns access and refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login Credentials"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	accessToken, refreshToken, err := h.authUC.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) { // Using errors.Is
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"}) // Refined error message
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to login"})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

