package service

import (
	"context"
	"errors"
	"testing"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"

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

func (m *MockUserRepository) FindAll(ctx context.Context) ([]*domain.User, error) { // Added FindAll
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error { // Added Update
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockTokenGenerator is a mock of token.Generator
type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateAccessToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGenerator) GenerateRefreshToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGenerator) ValidateToken(tokenString string) (*domain.CustomClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomClaims), args.Error(1)
}

// MockAuditLogRepository is a mock of AuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenGenerator := new(MockTokenGenerator)
	mockAuditRepo := new(MockAuditLogRepository)
	authService := NewAuthService(mockUserRepo, mockTokenGenerator, mockAuditRepo) // Changed service initialization
	ctx := context.Background()

	username := "testuser"
	password := "password123"

	t.Run("Successful Registration", func(t *testing.T) {
		mockUserRepo.On("FindByUsername", ctx, username).Return(nil, repository.ErrUserNotFound).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once() // Added audit log expectation

		err := authService.Register(ctx, username, password)
		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t) // Assert audit log
	})

	t.Run("User Already Exists", func(t *testing.T) {
		existingUser, _ := domain.NewUser(username, password, domain.RoleUser)
		mockUserRepo.On("FindByUsername", ctx, username).Return(existingUser, nil).Once()

		err := authService.Register(ctx, username, password)
		assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertNotCalled(t, "Create") // No audit log on early exit
	})

	t.Run("Registration Fails on Create", func(t *testing.T) {
		dbError := errors.New("database error")
		mockUserRepo.On("FindByUsername", ctx, username).Return(nil, repository.ErrUserNotFound).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(dbError).Once()

		err := authService.Register(ctx, username, password)
		assert.ErrorIs(t, err, dbError)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertNotCalled(t, "Create") // No audit log on DB error
	})
}

func TestAuthService_Login(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenGenerator := new(MockTokenGenerator)
	mockAuditRepo := new(MockAuditLogRepository)
	authService := NewAuthService(mockUserRepo, mockTokenGenerator, mockAuditRepo) // Changed service initialization
	ctx := context.Background()

	username := "testuser"
	password := "password123"
	user, _ := domain.NewUser(username, password, domain.RoleUser)
	user.ID = 1 // Set an ID for audit logs

	t.Run("Successful Login", func(t *testing.T) {
		mockUserRepo.On("FindByUsername", ctx, username).Return(user, nil).Once()
		mockTokenGenerator.On("GenerateAccessToken", user).Return("access_token", nil).Once()
		mockTokenGenerator.On("GenerateRefreshToken", user).Return("refresh_token", nil).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once() // Added audit log expectation

		accessToken, refreshToken, err := authService.Login(ctx, username, password)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		mockUserRepo.AssertExpectations(t)
		mockTokenGenerator.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t) // Assert audit log
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockUserRepo.On("FindByUsername", ctx, "unknownuser").Return(nil, repository.ErrUserNotFound).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once() // Added audit log expectation for failed login

		_, _, err := authService.Login(ctx, "unknownuser", password)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t) // Assert audit log
	})

	t.Run("Incorrect Password", func(t *testing.T) {
		mockUserRepo.On("FindByUsername", ctx, username).Return(user, nil).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once() // Added audit log expectation for failed login

		_, _, err := authService.Login(ctx, username, "wrongpassword")
		assert.ErrorIs(t, err, repository.ErrUserNotFound) // Should return same error as not found
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t) // Assert audit log
	})
}

