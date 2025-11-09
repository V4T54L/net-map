package http

import (
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// UserResponse is a DTO for user data sent to clients.
type UserResponse struct {
	ID        int64            `json:"id"`
	Username  string           `json:"username"`
	Role      domain.UserRole  `json:"role"`
	IsEnabled bool             `json:"is_enabled"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
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
	IsEnabled *bool `json:"is_enabled"`
}

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userUC usecase.UserUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// ListUsers handles requests to list all users.
func (h *UserHandler) ListUsers(c echo.Context) error {
	users, err := h.userUC.ListUsers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve users"})
	}
	return c.JSON(http.StatusOK, toUserResponseList(users))
}

// GetUser handles requests to get a single user by ID.
func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	user, err := h.userUC.GetUserByID(c.Request().Context(), id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user"})
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}

// UpdateUserStatus handles requests to enable or disable a user.
func (h *UserHandler) UpdateUserStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	req := new(UpdateUserStatusRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.IsEnabled == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "'is_enabled' field is required"})
	}

	user, err := h.userUC.UpdateUserStatus(c.Request().Context(), id, *req.IsEnabled)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user status"})
	}

	return c.JSON(http.StatusOK, toUserResponse(user))
}
