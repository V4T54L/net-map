package database

import (
	"context"
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
	"time"
)

type dnsRepoInMemory struct {
	hm map[string]*domain.DNSRecord
}

func NewDNSRecordInMemoryRepository() repository.DNSRecordRepository {
	hm := map[string]*domain.DNSRecord{
		"abc.abc": {
			ID:         1,
			UserID:     1,
			DomainName: "abc.abc",
			Type:       domain.A,
			Value:      "123.123.145.145",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		"pqr.pqr": {
			ID:         2,
			UserID:     1,
			DomainName: "pqr.pqr",
			Type:       domain.CNAME,
			Value:      "123.123.145.145",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		"abc.abc.": {
			ID:         3,
			UserID:     1,
			DomainName: "abc.abc",
			Type:       domain.CNAME,
			Value:      "123.123.145.145",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		"pqr.pqr.": {
			ID:         4,
			UserID:     1,
			DomainName: "pqr.pqr",
			Type:       domain.A,
			Value:      "123.123.145.145",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}
	return &dnsRepoInMemory{hm: hm}
}

func (r *dnsRepoInMemory) Create(ctx context.Context, record *domain.DNSRecord) error {
	return nil
}

func (r *dnsRepoInMemory) FindByID(ctx context.Context, id int64) (*domain.DNSRecord, error) {
	return nil, repository.ErrDNSRecordNotFound
}

func (r *dnsRepoInMemory) FindByDomainName(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	if val, ok := r.hm[domainName]; ok {
		return val, nil
	}
	return nil, repository.ErrDNSRecordNotFound
}

func (r *dnsRepoInMemory) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, error) {
	return nil, repository.ErrDNSRecordNotFound
}

func (r *dnsRepoInMemory) Update(ctx context.Context, record *domain.DNSRecord) error {
	return nil
}

func (r *dnsRepoInMemory) Delete(ctx context.Context, id int64) error {
	return nil
}

func (r *dnsRepoInMemory) CountByUserID(ctx context.Context, userID int64) (int, error) {
	return 0, nil
}

func (r *dnsRepoInMemory) GetAllDomainNames(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
