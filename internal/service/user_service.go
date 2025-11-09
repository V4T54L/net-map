package service

import (
	"context"
	"log" // Added log import
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase" // Keep usecase import for interface
)

type userService struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository // Added auditRepo
}

func NewUserService(userRepo repository.UserRepository, auditRepo repository.AuditLogRepository) usecase.UserUseCase { // Changed signature, kept usecase interface
	return &userService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

func (s *userService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) UpdateUserStatus(ctx context.Context, actorID, targetUserID int64, isEnabled bool) (*domain.User, error) { // Changed signature
	user, err := s.userRepo.FindByID(ctx, targetUserID) // Use targetUserID
	if err != nil {
		return nil, err
	}

	oldUser := *user // Make a copy for the audit log

	user.IsEnabled = isEnabled
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Fire-and-forget audit log
	go func() {
		auditLog, err := domain.NewAuditLog(actorID, domain.ActionUpdateUserStatus, targetUserID, oldUser, user)
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for user status update: %v", err)
			}
		}
	}()

	return user, nil
}

