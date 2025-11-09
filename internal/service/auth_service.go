package service

import (
	"context"
	"errors"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/token"
)

type authService struct {
	userRepo       repository.UserRepository
	tokenGenerator token.Generator
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, tokenGenerator token.Generator) usecase.AuthUseCase {
	return &authService{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
	}
}

func (s *authService) Register(ctx context.Context, username, password string) error {
	// Check if user already exists
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return repository.ErrUserAlreadyExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return err // Some other unexpected error
	}

	// Create new user
	user, err := domain.NewUser(username, password, domain.RoleUser)
	if err != nil {
		return err
	}

	return s.userRepo.Create(ctx, user)
}

func (s *authService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	if !user.ValidatePassword(password) {
		return "", "", repository.ErrUserNotFound // Use same error to prevent username enumeration
	}

	accessToken, err = s.tokenGenerator.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.tokenGenerator.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
```
```go
