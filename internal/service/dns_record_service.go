package service

import (
	"context"
	"errors"

	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase"
	"internal-dns/pkg/bloomfilter"
)

type dnsRecordService struct {
	dnsRepo     repository.DNSRecordRepository
	bloomFilter bloomfilter.Filter
	cache       cache.DNSRecordCache
}

// NewDNSRecordService creates a new DNSRecordUseCase implementation.
func NewDNSRecordService(dnsRepo repository.DNSRecordRepository, bf bloomfilter.Filter, cache cache.DNSRecordCache) usecase.DNSRecordUseCase {
	return &dnsRecordService{
		dnsRepo:     dnsRepo,
		bloomFilter: bf,
		cache:       cache,
	}
}

func (s *dnsRecordService) CreateRecord(ctx context.Context, userID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error) {
	exists, err := s.bloomFilter.Test(ctx, domainName)
	if err != nil {
		// Log the error but proceed, as bloom filter is probabilistic
	}
	if exists {
		// Potential duplicate, check DB
		_, err := s.dnsRepo.FindByDomainName(ctx, domainName)
		if err == nil {
			return nil, repository.ErrDuplicateDomainName
		}
		if !errors.Is(err, repository.ErrDNSRecordNotFound) {
			return nil, err
		}
	}

	record, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}

	if err := s.dnsRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// Add to bloom filter after successful DB insertion
	_ = s.bloomFilter.Add(ctx, record.DomainName)

	return record, nil
}

func (s *dnsRecordService) GetRecordByID(ctx context.Context, userID int64, recordID int64) (*domain.DNSRecord, error) {
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	// Security check: user can only get their own records
	if record.UserID != userID {
		return nil, repository.ErrDNSRecordNotFound
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
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	if record.UserID != userID {
		return nil, repository.ErrDNSRecordNotFound
	}

	originalDomainName := record.DomainName

	updatedRecord, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}

	record.DomainName = updatedRecord.DomainName
	record.Value = updatedRecord.Value
	record.Type = updatedRecord.Type

	if err := s.dnsRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	// Invalidate cache for the old domain name
	if err := s.cache.Delete(ctx, originalDomainName); err != nil {
		// Log this error, but don't fail the operation
	}
	if originalDomainName != record.DomainName {
		if err := s.cache.Delete(ctx, record.DomainName); err != nil {
			// Log this error
		}
	}

	return record, nil
}

func (s *dnsRecordService) DeleteRecord(ctx context.Context, userID int64, recordID int64) error {
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record.UserID != userID {
		return repository.ErrDNSRecordNotFound
	}

	if err := s.dnsRepo.Delete(ctx, recordID); err != nil {
		return err
	}

	// Invalidate cache
	if err := s.cache.Delete(ctx, record.DomainName); err != nil {
		// Log this error, but don't fail the operation
	}

	return nil
}

func (s *dnsRecordService) ResolveDomain(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	return s.dnsRepo.FindByDomainName(ctx, domainName)
}

