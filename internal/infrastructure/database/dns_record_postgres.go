package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"
)

type dnsRecordPostgresRepository struct {
	db *pgxpool.Pool
}

func NewDNSRecordPostgresRepository(db *pgxpool.Pool) repository.DNSRecordRepository {
	return &dnsRecordPostgresRepository{db: db}
}

func (r *dnsRecordPostgresRepository) Create(ctx context.Context, record *domain.DNSRecord) error {
	query := `INSERT INTO dns_records (user_id, domain_name, type, value)
              VALUES ($1, $2, $3, $4)
              RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, record.UserID, record.DomainName, record.Type, record.Value).
		Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrDuplicateDomainName
		}
		return err
	}
	return nil
}

func (r *dnsRecordPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.DNSRecord, error) {
	query := `SELECT id, user_id, domain_name, type, value, created_at, updated_at
              FROM dns_records WHERE id = $1`
	record := &domain.DNSRecord{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&record.ID, &record.UserID, &record.DomainName, &record.Type,
		&record.Value, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrDNSRecordNotFound
		}
		return nil, err
	}
	return record, nil
}

func (r *dnsRecordPostgresRepository) FindByDomainName(ctx context.Context, domainName string) (*domain.DNSRecord, error) {
	query := `SELECT id, user_id, domain_name, type, value, created_at, updated_at
              FROM dns_records WHERE domain_name = $1`
	record := &domain.DNSRecord{}
	err := r.db.QueryRow(ctx, query, domainName).Scan(
		&record.ID, &record.UserID, &record.DomainName, &record.Type,
		&record.Value, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrDNSRecordNotFound
		}
		return nil, err
	}
	return record, nil
}

func (r *dnsRecordPostgresRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*domain.DNSRecord, error) {
	query := `SELECT id, user_id, domain_name, type, value, created_at, updated_at
              FROM dns_records
              WHERE user_id = $1
              ORDER BY created_at DESC
              LIMIT $2 OFFSET $3`
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.DNSRecord
	for rows.Next() {
		record := &domain.DNSRecord{}
		err := rows.Scan(
			&record.ID, &record.UserID, &record.DomainName, &record.Type,
			&record.Value, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *dnsRecordPostgresRepository) CountByUserID(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM dns_records WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *dnsRecordPostgresRepository) Update(ctx context.Context, record *domain.DNSRecord) error {
	query := `UPDATE dns_records
              SET domain_name = $1, type = $2, value = $3, updated_at = NOW()
              WHERE id = $4
              RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, record.DomainName, record.Type, record.Value, record.ID).Scan(&record.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrDNSRecordNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrDuplicateDomainName
		}
		return err
	}
	return nil
}

func (r *dnsRecordPostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM dns_records WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrDNSRecordNotFound
	}
	return nil
}

func (r *dnsRecordPostgresRepository) GetAllDomainNames(ctx context.Context) ([]string, error) {
	query := `SELECT domain_name FROM dns_records`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domainName string
		if err := rows.Scan(&domainName); err != nil {
			return nil, err
		}
		domains = append(domains, domainName)
	}
	return domains, nil
}

