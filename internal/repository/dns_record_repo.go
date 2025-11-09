package repository

import (
	"context"
	"errors"

	"internal-dns/internal/domain"
)

var (
	ErrDNSRecordNotFound   = errors.New("dns record not found")
	ErrDuplicateDomainName = errors.New("a record with this domain name already exists")
)

type DNSRecordRepository interface {
	Create(ctx context.Context, record *domain.DNSRecord) error
	FindByID(ctx context.Context, id int64) (*domain.DNSRecord, error)
	FindByDomainName(ctx context.Context, domainName string) (*domain.DNSRecord, error)
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, error)
	Update(ctx context.Context, record *domain.DNSRecord) error
	Delete(ctx context.Context, id int64) error
	CountByUserID(ctx context.Context, userID int64) (int, error)
	GetAllDomainNames(ctx context.Context) ([]string, error)
}

