package database

import (
	"context"
	"errors"

	"internal-dns/internal/domain"
	"internal-dns/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userPostgresRepository struct {
	db *pgxpool.Pool
}

// NewUserPostgresRepository creates a new repository for user data.
func NewUserPostgresRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (username, password_hash, role, is_enabled)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `
	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Role, user.IsEnabled).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		// Handle potential unique constraint violation
		return err
	}
	return nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
        SELECT id, username, password_hash, role, is_enabled, created_at, updated_at
        FROM users
        WHERE username = $1
    `
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsEnabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
        SELECT id, username, password_hash, role, is_enabled, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsEnabled, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
```
```go
