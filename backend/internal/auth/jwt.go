package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sakkurohilla/kineticops/backend/config"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Generate JWT access token
func GenerateJWT(userID int64, username string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	cfg := config.Load()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// Validate JWT and parse claims
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	cfg := config.Load()
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
