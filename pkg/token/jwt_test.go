package token

import (
	"internal-dns/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTGenerator(t *testing.T) {
	secretKey := "test-secret-key"
	generator := NewJWTGenerator(secretKey)
	user := &domain.User{
		ID:       1,
		Username: "testuser",
		Role:     domain.RoleUser,
	}

	t.Run("GenerateAccessToken", func(t *testing.T) {
		tokenString, err := generator.GenerateAccessToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		claims, err := generator.ValidateToken(tokenString)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Role, claims.Role)
		assert.Equal(t, user.Username, claims.Subject)
		assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, time.Second*5)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		tokenString, err := generator.GenerateRefreshToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		claims, err := generator.ValidateToken(tokenString)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Role, claims.Role)
		assert.Equal(t, user.Username, claims.Subject)
		assert.WithinDuration(t, time.Now().Add(time.Hour*24*7), claims.ExpiresAt.Time, time.Second*5)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		// Valid token
		validToken, err := generator.GenerateAccessToken(user)
		require.NoError(t, err)
		_, err = generator.ValidateToken(validToken)
		assert.NoError(t, err)

		// Invalid signature
		otherGenerator := NewJWTGenerator("different-secret")
		invalidToken, err := otherGenerator.GenerateAccessToken(user)
		require.NoError(t, err)
		_, err = generator.ValidateToken(invalidToken)
		assert.Error(t, err)

		// Malformed token
		_, err = generator.ValidateToken("not.a.real.token")
		assert.Error(t, err)
	})
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	generator := NewJWTGenerator("benchmark-secret")
	user := &domain.User{
		ID:       1,
		Username: "benchuser",
		Role:     domain.RoleAdmin,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateAccessToken(user)
	}
}
