package usecase

import (
	"context"
	"internal-dns/internal/domain"
)

// UserUseCase defines the business logic for user management.
type UserUseCase interface {
	ListUsers(ctx context.Context) ([]*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateUserStatus(ctx context.Context, id int64, isEnabled bool) (*domain.User, error)
}

