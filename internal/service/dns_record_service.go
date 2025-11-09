package service

import (
	"context"
	"errors"
	"log" // Added log import

	"internal-dns/internal/domain"
	"internal-dns/internal/infrastructure/cache"
	"internal-dns/internal/repository"
	"internal-dns/internal/usecase" // Keep usecase import for interface
	"internal-dns/pkg/bloomfilter"
)

type dnsRecordService struct {
	dnsRepo     repository.DNSRecordRepository
	bloomFilter bloomfilter.Filter
	cache       cache.DNSRecordCache
	auditRepo   repository.AuditLogRepository // Added auditRepo
}

// NewDNSRecordService creates a new DNSRecordUseCase implementation.
func NewDNSRecordService(dnsRepo repository.DNSRecordRepository, bf bloomfilter.Filter, cache cache.DNSRecordCache, auditRepo repository.AuditLogRepository) usecase.DNSRecordUseCase { // Changed signature, kept usecase interface
	return &dnsRecordService{
		dnsRepo:     dnsRepo,
		bloomFilter: bf,
		cache:       cache,
		auditRepo:   auditRepo,
	}
}

func (s *dnsRecordService) CreateRecord(ctx context.Context, userID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error) {
	// 1. Check Bloom Filter first
	exists, err := s.bloomFilter.Test(ctx, domainName)
	if err != nil {
		// Log error but proceed, as DB is the source of truth
		log.Printf("Bloom filter check failed: %v", err) // Added log
	}
	if exists {
		// 2. If Bloom filter hits, check the database
		_, err := s.dnsRepo.FindByDomainName(ctx, domainName)
		if err == nil {
			return nil, repository.ErrDuplicateDomainName
		}
		if !errors.Is(err, repository.ErrDNSRecordNotFound) {
			return nil, err // A different database error occurred
		}
	}

	// 3. Create the domain entity (which includes validation)
	record, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}

	// 4. Persist to the database
	if err := s.dnsRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// 5. Add to Bloom Filter
	if err := s.bloomFilter.Add(ctx, record.DomainName); err != nil {
		// Log error, but don't fail the operation
		log.Printf("Failed to add domain to Bloom filter: %v", err) // Added log
	}

	// 6. Fire-and-forget audit log
	go func() {
		auditLog, err := domain.NewAuditLog(userID, domain.ActionCreateDNSRecord, record.ID, nil, record)
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for DNS record creation: %v", err)
			}
		}
	}()

	return record, nil
}

func (s *dnsRecordService) GetRecordByID(ctx context.Context, userID int64, recordID int64) (*domain.DNSRecord, error) {
	record, err := s.dnsRepo.FindByID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	// Security check: user can only get their own records
	if record.UserID != userID {
		return nil, repository.ErrDNSRecordNotFound // Hide existence from other users
	}

	return record, nil
}

func (s *dnsRecordService) ListRecordsByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, int, error) {
	if page < 1 { // Keep original validation
		page = 1
	}
	if pageSize < 1 || pageSize > 100 { // Keep original validation
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
	// 1. Verify ownership and get the old record
	oldRecord, err := s.GetRecordByID(ctx, userID, recordID) // Use GetRecordByID for ownership check
	if err != nil {
		return nil, err
	}

	// 2. Create a new domain entity for validation
	updatedRecord, err := domain.NewDNSRecord(userID, domainName, value, recordType)
	if err != nil {
		return nil, err
	}
	updatedRecord.ID = recordID         // Preserve original ID
	updatedRecord.CreatedAt = oldRecord.CreatedAt // Preserve original creation time

	// 3. Persist the update
	if err := s.dnsRepo.Update(ctx, updatedRecord); err != nil {
		return nil, err
	}

	// 4. Invalidate cache
	if err := s.cache.Delete(ctx, oldRecord.DomainName); err != nil {
		log.Printf("Failed to delete old domain from cache: %v", err) // Added log
	}
	if oldRecord.DomainName != updatedRecord.DomainName {
		if err := s.cache.Delete(ctx, updatedRecord.DomainName); err != nil {
			log.Printf("Failed to delete new domain from cache: %v", err) // Added log
		}
	}

	// 5. Fire-and-forget audit log
	go func() {
		auditLog, err := domain.NewAuditLog(userID, domain.ActionUpdateDNSRecord, recordID, oldRecord, updatedRecord)
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for DNS record update: %v", err)
			}
		}
	}()

	return updatedRecord, nil
}

func (s *dnsRecordService) DeleteRecord(ctx context.Context, userID int64, recordID int64) error {
	// 1. Verify ownership and get the record to be deleted
	record, err := s.GetRecordByID(ctx, userID, recordID) // Use GetRecordByID for ownership check
	if err != nil {
		return err
	}

	// 2. Delete from the database
	if err := s.dnsRepo.Delete(ctx, recordID); err != nil {
		return err
	}

	// 3. Invalidate cache
	if err := s.cache.Delete(ctx, record.DomainName); err != nil {
		log.Printf("Failed to delete domain from cache: %v", err) // Added log
	}

	// 4. Fire-and-forget audit log
	go func() {
		auditLog, err := domain.NewAuditLog(userID, domain.ActionDeleteDNSRecord, recordID, record, nil)
		if err == nil {
			if err := s.auditRepo.Create(context.Background(), auditLog); err != nil {
				log.Printf("failed to create audit log for DNS record deletion: %v", err)
			}
		}
	}()

	return nil
}

func (s *dnsRecordService) ResolveDomain(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	return s.dnsRepo.FindByDomainName(ctx, domainName)
}

