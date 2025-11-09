package service

import (
	"context"
	"fmt"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/bloomfilter"
)

type dnsRecordService struct {
	dnsRepo     repository.DNSRecordRepository
	bloomFilter bloomfilter.Filter
	// auditRepo repository.AuditLogRepository // To be added later
}

func NewDNSRecordService(dnsRepo repository.DNSRecordRepository, bf bloomfilter.Filter) usecase.DNSRecordUseCase {
	return &dnsRecordService{
		dnsRepo:     dnsRepo,
		bloomFilter: bf,
	}
}

func (s *dnsRecordService) CreateRecord(ctx context.Context, userID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error) {
	// 1. Check bloom filter for potential duplicates
	exists, err := s.bloomFilter.Test(ctx, domainName)
	if err != nil {
		// Log the error but proceed, as bloom filter is probabilistic
		fmt.Printf("Warning: Bloom filter check failed: %v\n", err)
	}
	if exists {
		// If bloom filter says it exists, it might be a duplicate.
		// The database will give the final confirmation.
		// We can return a specific error here to hint at a potential duplicate.
		return nil, repository.ErrDuplicateDomainName
	}

	// 2. Create domain object (which includes validation)
	record, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}

	// 3. Persist to database
	if err := s.dnsRepo.Create(ctx, record); err != nil {
		if err == repository.ErrDuplicateDomainName {
			// Sync bloom filter if DB confirms duplicate
			_ = s.bloomFilter.Add(ctx, domainName)
		}
		return nil, err
	}

	// 4. Add to bloom filter on success
	if err := s.bloomFilter.Add(ctx, domainName); err != nil {
		// Log error, but don't fail the operation
		fmt.Printf("Warning: Failed to add domain to bloom filter: %v\n", err)
	}

	// TODO: Add audit log entry

	return record, nil
}

func (s *dnsRecordService) GetRecordByID(ctx context.Context, userID int64, recordID int64) (*domain.DNSRecord, error) {
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	// Authorization check: user can only get their own records
	if record.UserID != userID {
		return nil, repository.ErrDNSRecordNotFound // Obscure error for security
	}

	return record, nil
}

func (s *dnsRecordService) ListRecordsByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	records, err := s.dnsRepo.FindByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.dnsRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (s *dnsRecordService) UpdateRecord(ctx context.Context, userID int64, recordID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error) {
	// 1. Fetch existing record
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	// 2. Authorization check
	if record.UserID != userID {
		return nil, repository.ErrDNSRecordNotFound
	}

	// 3. Validate new data
	updatedRecord, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}
	updatedRecord.ID = record.ID
	updatedRecord.CreatedAt = record.CreatedAt

	// 4. Persist update
	if err := s.dnsRepo.Update(ctx, updatedRecord); err != nil {
		return nil, err
	}

	// TODO: Add audit log entry

	return updatedRecord, nil
}

func (s *dnsRecordService) DeleteRecord(ctx context.Context, userID int64, recordID int64) error {
	// 1. Fetch existing record to check ownership
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return err
	}

	// 2. Authorization check
	if record.UserID != userID {
		return repository.ErrDNSRecordNotFound
	}

	// 3. Perform deletion
	if err := s.dnsRepo.Delete(ctx, recordID); err != nil {
		return err
	}

	// Note: We don't remove from the bloom filter as it's not supported.
	// It will be rebuilt periodically.

	// TODO: Add audit log entry

	return nil
}

