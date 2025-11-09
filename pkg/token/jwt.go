package token

import (
	"errors"
	"fmt"
	"internal-dns/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generator defines the interface for JWT token generation and validation.
type Generator interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken(user *domain.User) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error)
}

// jwtGenerator implements the Generator interface.
type jwtGenerator struct {
	secretKey string
}

// CustomClaims contains custom JWT claims.
type CustomClaims struct {
	UserID int64           `json:"user_id"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTGenerator creates a new JWT generator.
func NewJWTGenerator(secretKey string) Generator {
	return &jwtGenerator{secretKey: secretKey}
}

// GenerateAccessToken generates a new access token for a user.
func (g *jwtGenerator) GenerateAccessToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // 1 hour
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

// GenerateRefreshToken generates a new refresh token for a user.
func (g *jwtGenerator) GenerateRefreshToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

// ValidateToken validates a JWT token string.
func (g *jwtGenerator) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

