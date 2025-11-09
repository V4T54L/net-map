package token

import (
	"errors" // Added from attempted
	"time"

	"internal-dns/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// Generator defines an interface for token generation.
type Generator interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken(user *domain.User) (string, error)
	ValidateToken(tokenString string) (*CustomClaims, error) // Added from attempted
}

type jwtGenerator struct {
	secretKey string
}

// CustomClaims extends standard claims with user-specific data.
type CustomClaims struct {
	UserID int64           `json:"user_id"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTGenerator creates a new JWT token generator.
func NewJWTGenerator(secretKey string) Generator {
	return &jwtGenerator{secretKey: secretKey}
}

func (g *jwtGenerator) GenerateAccessToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // 1 hour
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username, // Changed from fmt.Sprintf("%d", user.ID)
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

func (g *jwtGenerator) GenerateRefreshToken(user *domain.User) (string, error) {
	claims := &CustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username, // Changed from fmt.Sprintf("%d", user.ID)
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

func (g *jwtGenerator) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(g.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
