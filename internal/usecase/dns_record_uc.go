package usecase

import (
	"context"

	"internal-dns/internal/domain"
)

type DNSRecordUseCase interface {
	CreateRecord(ctx context.Context, userID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error)
	GetRecordByID(ctx context.Context, userID int64, recordID int64) (*domain.DNSRecord, error)
	ListRecordsByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, int, error)
	UpdateRecord(ctx context.Context, userID int64, recordID int64, domainName, value string, recordType domain.RecordType) (*domain.DNSRecord, error)
	DeleteRecord(ctx context.Context, userID int64, recordID int64) error
}

