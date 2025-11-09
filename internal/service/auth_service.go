package service

import (
	"context"
	"errors"
	"log" // Added log import

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase" // Keep usecase import for interface
	"internal-dns/pkg/token"
)

type authService struct {
	userRepo       repository.UserRepository
	tokenGenerator token.Generator
	auditRepo      repository.AuditLogRepository // Added auditRepo
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo repository.UserRepository, tokenGenerator token.Generator, auditRepo repository.AuditLogRepository) usecase.AuthUseCase { // Changed signature, kept usecase interface
	return &authService{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
		auditRepo:      auditRepo,
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

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Fire-and-forget audit log
	go func() {
		auditLog, err := domain.NewAuditLog(user.ID, domain.ActionUserRegister, user.ID, nil, map[string]string{"username": user.Username})
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for user registration: %v", err)
			}
		}
	}()

	return nil
}

func (s *authService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		// Fire-and-forget audit log for failed login
		go func() {
			auditLog, err := domain.NewAuditLog(0, domain.ActionUserLoginFailure, 0, nil, map[string]string{"username": username})
			if err == nil {
				if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
					log.Printf("failed to create audit log for failed login: %v", err)
				}
			}
		}()
		return "", "", repository.ErrUserNotFound // Use same error to prevent username enumeration
	}

	if !user.ValidatePassword(password) {
		// Fire-and-forget audit log for failed login
		go func() {
			auditLog, err := domain.NewAuditLog(user.ID, domain.ActionUserLoginFailure, user.ID, nil, map[string]string{"username": username})
			if err == nil {
				if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
					log.Printf("failed to create audit log for failed login: %v", err)
				}
			}
		}()
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

	// Fire-and-forget audit log for successful login
	go func() {
		auditLog, err := domain.NewAuditLog(user.ID, domain.ActionUserLoginSuccess, user.ID, nil, nil)
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for successful login: %v", err)
			}
		}
	}()

	return accessToken, refreshToken, nil
}

