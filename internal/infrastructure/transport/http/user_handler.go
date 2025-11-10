package http

import (
	"errors"
	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/transport/http/middleware"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// UserResponse is a DTO for user data sent to clients.
type UserResponse struct {
	ID        int64           `json:"id"`
	Username  string          `json:"username"`
	Role      domain.UserRole `json:"role"`
	IsEnabled bool            `json:"isEnabled"` // Changed to camelCase
	CreatedAt time.Time       `json:"createdAt"` // Changed to camelCase
	UpdatedAt time.Time       `json:"updatedAt"` // Changed to camelCase
}

func toUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		IsEnabled: user.IsEnabled,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toUserResponseList(users []*domain.User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, user := range users {
		res[i] = toUserResponse(user)
	}
	return res
}

// UpdateUserStatusRequest defines the payload for updating a user's status.
type UpdateUserStatusRequest struct {
	IsEnabled *bool `json:"isEnabled" validate:"required"` // Changed to camelCase
}

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userUC usecase.UserUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// ListUsers godoc
// @Summary List all users
// @Description Retrieves a list of all users. (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c echo.Context) error {
	users, err := h.userUC.ListUsers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve users"}) // Changed error response
	}
	return c.JSON(http.StatusOK, toUserResponseList(users))
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Retrieves a single user by their ID. (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"}) // Changed error response
	}

	user, err := h.userUC.GetUserByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) { // Using errors.Is
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"}) // Changed error response
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user"}) // Changed error response
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

// UpdateUserStatus godoc
// @Summary Update a user's status
// @Description Enables or disables a user account. (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param status body UpdateUserStatusRequest true "Update Status Payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users/{id}/status [put]
func (h *UserHandler) UpdateUserStatus(c echo.Context) error {
	actor, ok := c.Get(string(middleware.UserContextKey)).(*domain.User) // Added actor retrieval
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid actor in context"}) // Changed error response
	}

	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64) // Renamed id to targetUserID
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"}) // Changed error response
	}

	var req UpdateUserStatusRequest                              // Changed to var declaration
	if err := c.Bind(&req); err != nil || req.IsEnabled == nil { // Combined bind and nil check
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload, 'isEnabled' is required"}) // Changed error response
	}

	updatedUser, err := h.userUC.UpdateUserStatus(c.Request().Context(), actor.ID, targetUserID, *req.IsEnabled) // Changed signature
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) { // Using errors.Is
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"}) // Changed error response
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user status"}) // Changed error response
	}

	return c.JSON(http.StatusOK, toUserResponse(updatedUser))
}
