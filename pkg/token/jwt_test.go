package token

import (
	"internal-dns/internal/domain"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require" // Added from attempted
)

func TestJWTGenerator(t *testing.T) {
	secretKey := "supersecretkey" // Changed variable name
	generator := NewJWTGenerator(secretKey)

	user := &domain.User{
		ID:       1,
		Username: "testuser", // Added from attempted
		Role:     domain.RoleUser,
	}

	t.Run("GenerateAccessToken", func(t *testing.T) { // Renamed test
		tokenString, err := generator.GenerateAccessToken(user)
		require.NoError(t, err) // Changed from assert.NoError
		assert.NotEmpty(t, tokenString)

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		require.NoError(t, err) // Changed from assert.NoError
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(*CustomClaims)
		require.True(t, ok) // Changed from assert.True
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Role, claims.Role)
		assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, time.Second*5) // Changed duration format
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) { // Renamed test
		tokenString, err := generator.GenerateRefreshToken(user)
		require.NoError(t, err) // Changed from assert.NoError
		assert.NotEmpty(t, tokenString)

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		require.NoError(t, err) // Changed from assert.NoError
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(*CustomClaims)
		require.True(t, ok) // Changed from assert.True
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Role, claims.Role)
		assert.WithinDuration(t, time.Now().Add(time.Hour*24*7), claims.ExpiresAt.Time, time.Second*5) // Changed duration format
	})

	t.Run("ValidateToken", func(t *testing.T) { // Added from attempted
		t.Run("Valid Token", func(t *testing.T) {
			tokenString, err := generator.GenerateAccessToken(user)
			require.NoError(t, err)

			claims, err := generator.ValidateToken(tokenString)
			require.NoError(t, err)
			assert.Equal(t, user.ID, claims.UserID)
			assert.Equal(t, user.Role, claims.Role)
			assert.Equal(t, user.Username, claims.Subject)
		})

		t.Run("Invalid Token - Bad Signature", func(t *testing.T) {
			otherGenerator := NewJWTGenerator("differentsecret")
			tokenString, err := otherGenerator.GenerateAccessToken(user)
			require.NoError(t, err)

			_, err = generator.ValidateToken(tokenString)
			assert.Error(t, err)
		})

		t.Run("Invalid Token - Malformed", func(t *testing.T) {
			_, err := generator.ValidateToken("bad.token.string")
			assert.Error(t, err)
		})
	})
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	generator := NewJWTGenerator("benchmarksecret")
	user := &domain.User{ID: 1, Username: "benchuser", Role: domain.RoleUser} // Added Username
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateAccessToken(user)
	}
}
