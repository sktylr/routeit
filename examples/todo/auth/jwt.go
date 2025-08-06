package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessTTL  = 15 * time.Minute
	refreshTTL = 14 * 24 * time.Hour
	// Ordinarily this would not be hardcoded
	secretKey = []byte("super-secret-key")
)

type JwtToken struct {
	AccessToken  string
	RefreshToken string
}

func GenerateTokens(uid string) (JwtToken, error) {
	now := time.Now()

	accessCl := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		Issuer:    "todo-sample-app",
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   uid,
	}
	refreshCl := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
		Issuer:    "todo-sample-app",
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   uid,
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS384, accessCl)
	accessT, err := access.SignedString(secretKey)
	if err != nil {
		return JwtToken{}, err
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS384, refreshCl)
	refreshT, err := refresh.SignedString(secretKey)
	if err != nil {
		return JwtToken{}, nil
	}

	return JwtToken{AccessToken: accessT, RefreshToken: refreshT}, nil
}
