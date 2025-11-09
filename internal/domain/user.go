package domain

import (
	"errors"
	"internal-dns/internal/util" // Added from attempted
	"regexp"
	"time"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User represents a user in the system.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Exclude from JSON responses
	Role         UserRole  `json:"role"`
	IsEnabled    bool      `json:"is_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var (
	ErrUsernameTooShort = errors.New("username must be at least 3 characters long")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	ErrInvalidRole      = errors.New("invalid user role")
)

// NewUser creates a new User instance with validation.
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

	hashedPassword, err := util.HashPassword(password) // Changed to use util
	if err != nil {
		return nil, err
	}

	return &User{
		Username:     username,
		PasswordHash: hashedPassword,
		Role:         role,
		IsEnabled:    true, // Users are enabled by default
	}, nil
}

// ValidatePassword checks if the provided password matches the user's hashed password.
func (u *User) ValidatePassword(password string) bool {
	return util.CheckPasswordHash(password, u.PasswordHash) // Changed to use util
}

// SanitizeEmail is a placeholder for email validation.
func SanitizeEmail(email string) (string, error) {
	// A very basic email regex for demonstration
	if !regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(email) {
		return "", errors.New("invalid email format")
	}
	return email, nil
}
