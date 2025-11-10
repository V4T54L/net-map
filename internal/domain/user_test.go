package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		role     UserRole
		wantErr  error
	}{
		{"Valid User", "testuser", "password123", RoleUser, nil},
		{"Valid Admin", "adminuser", "password123", RoleAdmin, nil},
		{"Short Username", "us", "password123", RoleUser, ErrUsernameTooShort},
		{"Short Password", "testuser", "pass", RoleUser, ErrPasswordTooShort},
		{"Invalid Role", "testuser", "password123", "guest", ErrInvalidRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.password, tt.role)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.role, user.Role)
				assert.True(t, user.IsEnabled)
				assert.NotEmpty(t, user.PasswordHash)
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	password := "strong-password-123"
	user, err := NewUser("validator", password, RoleUser)
	assert.NoError(t, err)
	assert.NotNil(t, user)

	assert.True(t, user.ValidatePassword(password), "Password should be valid")
	assert.False(t, user.ValidatePassword("wrong-password"), "Password should be invalid")
}
