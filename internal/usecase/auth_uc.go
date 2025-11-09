package usecase

import (
	"context"
)

// AuthUseCase defines the interface for authentication-related operations.
type AuthUseCase interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
}

