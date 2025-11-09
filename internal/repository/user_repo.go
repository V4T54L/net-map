package repository

import (
	"context"
	"errors"

	"internal-dns/internal/domain"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this username already exists")
)

// UserRepository defines the interface for user data storage.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
}
```
```go
