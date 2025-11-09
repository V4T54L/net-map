package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
)

// MockDNSRecordRepository is a mock implementation of DNSRecordRepository
type MockDNSRecordRepository struct {
	mock.Mock
}

func (m *MockDNSRecordRepository) Create(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}
func (m *MockDNSRecordRepository) FindByID(ctx context.Context, id int64) (*domain.DNSRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}
func (m *MockDNSRecordRepository) FindByDomainName(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	args := m.Called(ctx, domainName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}
func (m *MockDNSRecordRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DNSRecord), args.Error(1)
}
func (m *MockDNSRecordRepository) Update(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}
func (m *MockDNSRecordRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockDNSRecordRepository) CountByUserID(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *MockDNSRecordRepository) GetAllDomainNames(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockBloomFilter is a mock implementation of bloomfilter.Filter
type MockBloomFilter struct {
	mock.Mock
}

func (m *MockBloomFilter) Add(ctx context.Context, value string) error {
	args := m.Called(ctx, value)
	return args.Error(0)
}
func (m *MockBloomFilter) Test(ctx context.Context, value string) (bool, error) {
	args := m.Called(ctx, value)
	return args.Bool(0), args.Error(1)
}
func (m *MockBloomFilter) AddMulti(ctx context.Context, values []string) error {
	args := m.Called(ctx, values)
	return args.Error(0)
}

func TestDNSRecordService_CreateRecord(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockDNSRecordRepository)
	mockBF := new(MockBloomFilter)
	service := NewDNSRecordService(mockRepo, mockBF)

	domainName := "test.service.local"
	value := "10.0.0.1"
	recordType := domain.A

	t.Run("Success", func(t *testing.T) {
		mockBF.On("Test", ctx, domainName).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.DNSRecord")).Return(nil).Once()
		mockBF.On("Add", ctx, domainName).Return(nil).Once()

		record, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, domainName, record.DomainName)
		mockRepo.AssertExpectations(t)
		mockBF.AssertExpectations(t)
	})

	t.Run("Duplicate detected by Bloom Filter", func(t *testing.T) {
		mockBF.On("Test", ctx, domainName).Return(true, nil).Once()

		_, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrDuplicateDomainName)
		mockBF.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Duplicate detected by Database", func(t *testing.T) {
		mockBF.On("Test", ctx, domainName).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.DNSRecord")).Return(repository.ErrDuplicateDomainName).Once()
		// Service should sync bloom filter on DB duplicate error
		mockBF.On("Add", ctx, domainName).Return(nil).Once()

		_, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrDuplicateDomainName)
		mockRepo.AssertExpectations(t)
		mockBF.AssertExpectations(t)
	})

	t.Run("Invalid Domain Data", func(t *testing.T) {
		_, err := service.CreateRecord(ctx, 1, "invalid", value, recordType)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomainName)
	})

	t.Run("Database Create Error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockBF.On("Test", ctx, domainName).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.DNSRecord")).Return(dbErr).Once()

		_, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockRepo.AssertExpectations(t)
		mockBF.AssertExpectations(t)
	})
}

