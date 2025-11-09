package service

import (
	"context"
	"errors"
	"testing"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/pkg/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenGen := token.NewJWTGenerator("test-secret")
	authService := NewAuthService(mockRepo, tokenGen)
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	t.Run("Successful Registration", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, username).Return(nil, repository.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := authService.Register(ctx, username, password)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Already Exists", func(t *testing.T) {
		existingUser, _ := domain.NewUser(username, password, domain.RoleUser)
		mockRepo.On("FindByUsername", ctx, username).Return(existingUser, nil).Once()

		err := authService.Register(ctx, username, password)
		assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Registration Fails on Create", func(t *testing.T) {
		dbError := errors.New("database error")
		mockRepo.On("FindByUsername", ctx, username).Return(nil, repository.ErrUserNotFound).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(dbError).Once()

		err := authService.Register(ctx, username, password)
		assert.ErrorIs(t, err, dbError)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	tokenGen := token.NewJWTGenerator("test-secret")
	authService := NewAuthService(mockRepo, tokenGen)
	ctx := context.Background()

	username := "testuser"
	password := "password123"
	user, _ := domain.NewUser(username, password, domain.RoleUser)

	t.Run("Successful Login", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, username).Return(user, nil).Once()

		accessToken, refreshToken, err := authService.Login(ctx, username, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, username).Return(nil, repository.ErrUserNotFound).Once()

		_, _, err := authService.Login(ctx, username, password)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Incorrect Password", func(t *testing.T) {
		mockRepo.On("FindByUsername", ctx, username).Return(user, nil).Once()

		_, _, err := authService.Login(ctx, username, "wrongpassword")
		assert.ErrorIs(t, err, repository.ErrUserNotFound) // Should return same error as not found
		mockRepo.AssertExpectations(t)
	})
}
```
```go
