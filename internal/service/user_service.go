package service

import (
	"context"
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) usecase.UserUseCase {
	return &userService{userRepo: userRepo}
}

func (s *userService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) UpdateUserStatus(ctx context.Context, id int64, isEnabled bool) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.IsEnabled = isEnabled
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

