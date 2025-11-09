package token

import (
	"testing"
	"time"

	"internal-dns/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTGenerator(t *testing.T) {
	secret := "test-secret-key"
	generator := NewJWTGenerator(secret)

	user := &domain.User{
		ID:   1,
		Role: domain.RoleUser,
	}

	t.Run("Generate Access Token", func(t *testing.T) {
		tokenString, err := generator.GenerateAccessToken(user)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		// Parse and validate
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		assert.NoError(t, err)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(*CustomClaims)
		assert.True(t, ok)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Role, claims.Role)
		assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
	})

	t.Run("Generate Refresh Token", func(t *testing.T) {
		tokenString, err := generator.GenerateRefreshToken(user)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		// Parse and validate
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		assert.NoError(t, err)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(*CustomClaims)
		assert.True(t, ok)
		assert.Equal(t, user.ID, claims.UserID)
		assert.WithinDuration(t, time.Now().Add(time.Hour*24*7), claims.ExpiresAt.Time, 5*time.Second)
	})
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	generator := NewJWTGenerator("benchmark-secret")
	user := &domain.User{ID: 1, Role: domain.RoleAdmin}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateAccessToken(user)
	}
}
```
