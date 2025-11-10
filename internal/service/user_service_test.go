package service

import (
	"context"
	"internal-dns/internal/domain"

	"github.com/stretchr/testify/mock"
)

// Re-declaring mocks here to keep test files self-contained as per project style.
type MockUserRepoForTest struct {
	mock.Mock
}

func (m *MockUserRepoForTest) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepoForTest) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepoForTest) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepoForTest) FindAll(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}
func (m *MockUserRepoForTest) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockAuditRepoForTest struct {
	mock.Mock
}

func (m *MockAuditRepoForTest) Create(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// func TestUserService_UpdateUserStatus(t *testing.T) {
// 	mockUserRepo := new(MockUserRepoForTest)
// 	mockAuditRepo := new(MockAuditRepoForTest)
// 	s := NewUserService(mockUserRepo, mockAuditRepo)

// 	ctx := context.Background()
// 	actorID := int64(1)
// 	targetUserID := int64(2)
// 	now := time.Now()

// 	userToUpdate := &domain.User{
// 		ID:        targetUserID,
// 		Username:  "testuser",
// 		IsEnabled: true,
// 		CreatedAt: now,
// 		UpdatedAt: now,
// 	}

// 	t.Run("success", func(t *testing.T) {
// 		mockUserRepo.On("FindByID", ctx, targetUserID).Return(userToUpdate, nil).Once()
// 		mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()
// 		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

// 		updatedUser, err := s.UpdateUserStatus(ctx, actorID, targetUserID, false)

// 		assert.NoError(t, err)
// 		assert.NotNil(t, updatedUser)
// 		assert.False(t, updatedUser.IsEnabled)
// 		mockUserRepo.AssertExpectations(t)
// 		mockAuditRepo.AssertExpectations(t)
// 	})

// 	t.Run("user not found", func(t *testing.T) {
// 		mockUserRepo.On("FindByID", ctx, targetUserID).Return(nil, repository.ErrUserNotFound).Once()

// 		_, err := s.UpdateUserStatus(ctx, actorID, targetUserID, false)

// 		assert.ErrorIs(t, err, repository.ErrUserNotFound)
// 		mockUserRepo.AssertExpectations(t)
// 	})

// 	t.Run("update fails", func(t *testing.T) {
// 		dbErr := errors.New("db error")
// 		// Create a fresh user object for this test run to avoid data races
// 		userForTest := &domain.User{
// 			ID:        targetUserID,
// 			Username:  "testuser",
// 			IsEnabled: true,
// 			CreatedAt: now,
// 			UpdatedAt: now,
// 		}
// 		mockUserRepo.On("FindByID", ctx, targetUserID).Return(userForTest, nil).Once()
// 		mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(dbErr).Once()

// 		_, err := s.UpdateUserStatus(ctx, actorID, targetUserID, false)

// 		assert.ErrorIs(t, err, dbErr)
// 		mockUserRepo.AssertExpectations(t)
// 	})
// }
