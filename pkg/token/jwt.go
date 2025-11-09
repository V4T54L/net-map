package token

import (
	"fmt"
	"time"

	"internal-dns/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// Generator defines the interface for creating tokens.
type Generator interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken(user *domain.User) (string, error)
}

// jwtGenerator is a JWT token generator.
type jwtGenerator struct {
	secretKey string
}

// NewJWTGenerator creates a new JWT generator.
func NewJWTGenerator(secretKey string) Generator {
	return &jwtGenerator{secretKey: secretKey}
}

// CustomClaims defines the custom claims for the JWT.
type CustomClaims struct {
	UserID int64             `json:"user_id"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a new access token for a user.
func (g *jwtGenerator) GenerateAccessToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // 1 hour
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

// GenerateRefreshToken creates a new refresh token for a user.
func (g *jwtGenerator) GenerateRefreshToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}
```
```go
