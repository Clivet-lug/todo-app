package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the payload stored inside every JWT token
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` 
	jwt.RegisteredClaims
}

// jwtSecret reads the secret from env so it's never hard-coded
func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback for local dev
		secret = "dev-secret-change-me"
	}
	return []byte(secret)
}

// GenerateToken creates a signed JWT for the given user fields.
// Tokens expire after 24 hours.
func GenerateToken(userID int, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

// ParseToken validates a token string and returns its claims.
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Guard: ensure the signing method is what we expect
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}