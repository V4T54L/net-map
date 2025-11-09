package domain

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTooShort = errors.New("username must be at least 3 characters long")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrInvalidRole      = errors.New("invalid user role")
)

// UserRole defines the type for user roles
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User represents a user in the system.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         UserRole
	IsEnabled    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new user instance and validates its properties.
func NewUser(username, password string, role UserRole) (*User, error) {
	if len(username) < 3 {
		return nil, ErrUsernameTooShort
	}
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}
	if role != RoleUser && role != RoleAdmin {
		return nil, ErrInvalidRole
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsEnabled:    true, // Users are enabled by default
	}, nil
}

// ValidatePassword checks if the provided password matches the user's hashed password.
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// SanitizeEmail is a placeholder for email validation logic.
func SanitizeEmail(email string) (string, error) {
	// Simple regex for email validation
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return "", errors.New("invalid email format")
	}
	return email, nil
}
```
```go
