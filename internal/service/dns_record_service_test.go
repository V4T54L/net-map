package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require" // Keep require for initial checks

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
)

// MockDNSRecordRepository is a mock implementation of DNSRecordRepository
type MockDNSRecordRepository struct {
	mock.Mock
}

func (m *MockDNSRecordRepository) Create(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	record.ID = 1 // Simulate DB setting the ID
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

// MockDNSRecordCache is a mock of cache.DNSRecordCache
type MockDNSRecordCache struct {
	mock.Mock
}

func (m *MockDNSRecordCache) Get(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	args := m.Called(ctx, domainName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DNSRecord), args.Error(1)
}

func (m *MockDNSRecordCache) Set(ctx context.Context, record *domain.DNSRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockDNSRecordCache) Delete(ctx context.Context, domainName string) error {
	args := m.Called(ctx, domainName)
	return args.Error(0)
}

func TestDNSRecordService_CreateRecord(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockDNSRecordRepository)
	mockBF := new(MockBloomFilter)
	mockCache := new(MockDNSRecordCache)
	mockAuditRepo := new(MockAuditLogRepository)
	service := NewDNSRecordService(mockRepo, mockBF, mockCache, mockAuditRepo) // Changed service initialization

	domainName := "test.service.local"
	value := "10.0.0.1"
	recordType := domain.A

	t.Run("Success", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		mockBF.On("Test", ctx, domainName).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.DNSRecord")).Return(nil).Once()
		mockBF.On("Add", ctx, domainName).Return(nil).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).
			Run(func(args mock.Arguments) {
				wg.Done()
			}).
			Return(nil).
			Once()

		record, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.NoError(t, err)
		require.NotNil(t, record)

		// Wait for audit goroutine to finish or timeout.
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		// OK
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for audit log creation goroutine")
		}

		assert.Equal(t, domainName, record.DomainName)
		mockRepo.AssertExpectations(t)
		mockBF.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t) // Assert audit log
	})

	t.Run("duplicate detected by bloom filter and db", func(t *testing.T) { // Updated test case name
		mockBF.On("Test", ctx, domainName).Return(true, nil).Once()
		mockRepo.On("FindByDomainName", ctx, domainName).Return(&domain.DNSRecord{}, nil).Once() // Added FindByDomainName call

		_, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrDuplicateDomainName)
		mockBF.AssertExpectations(t)
		mockRepo.AssertExpectations(t) // Assert FindByDomainName
		mockRepo.AssertNotCalled(t, "Create")
		mockAuditRepo.AssertNotCalled(t, "Create")
	})

	// t.Run("Invalid Domain Data", func(t *testing.T) {
	// 	_, err := service.CreateRecord(ctx, 1, "invalid", value, recordType)
	// 	require.Error(t, err)
	// 	assert.ErrorIs(t, err, domain.ErrInvalidDomainName)
	// 	mockBF.AssertNotCalled(t, "Test") // Should not reach bloom filter
	// 	mockRepo.AssertNotCalled(t, "Create")
	// 	mockAuditRepo.AssertNotCalled(t, "Create")
	// })

	t.Run("Database Create Error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockBF.On("Test", ctx, domainName).Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.DNSRecord")).Return(dbErr).Once()

		_, err := service.CreateRecord(ctx, 1, domainName, value, recordType)

		require.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockRepo.AssertExpectations(t)
		mockBF.AssertExpectations(t)
		mockAuditRepo.AssertNotCalled(t, "Create") // No audit log on DB error
	})
}
