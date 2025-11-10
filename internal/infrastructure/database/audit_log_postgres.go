package database

import (
	"context"
	"internal-dns/internal/domain"
	"internal-dns/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type auditLogPostgresRepository struct {
	db *pgxpool.Pool
}

func NewAuditLogPostgresRepository(db *pgxpool.Pool) repository.AuditLogRepository {
	return &auditLogPostgresRepository{db: db}
}

func (r *auditLogPostgresRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, target_id, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, log.UserID, log.Action, log.TargetID, log.OldValue, log.NewValue)
	return err
}
