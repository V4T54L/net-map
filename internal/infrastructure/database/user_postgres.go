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

func NewUserPostgresRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, password_hash, role, is_enabled) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Role, user.IsEnabled).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// A more robust implementation would check for specific constraint violations
		return err
	}
	return nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, is_enabled, created_at, updated_at 
              FROM users WHERE username = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsEnabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, is_enabled, created_at, updated_at 
              FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsEnabled, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) FindAll(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, username, password_hash, role, is_enabled, created_at, updated_at 
              FROM users ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsEnabled, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users 
              SET username = $1, role = $2, is_enabled = $3, updated_at = NOW() 
              WHERE id = $4
              RETURNING updated_at`
	err := r.db.QueryRow(ctx, query, user.Username, user.Role, user.IsEnabled, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrUserNotFound
		}
		return err
	}
	return nil
}
