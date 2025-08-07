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

type Claims struct {
	jwt.RegisteredClaims
	Type string
}

func GenerateTokens(uid string) (JwtToken, error) {
	now := time.Now()

	accessCl := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			Issuer:    "todo-sample-app",
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   uid,
		},
		Type: "access",
	}
	refreshCl := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			Issuer:    "todo-sample-app",
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   uid,
		},
		Type: "refresh",
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

func (c *Claims) IsExpired() bool {
	return c.ExpiresAt.Before(time.Now())
}

func ParseAccessToken(raw string) (*Claims, error) {
	return parseToken(raw, "access")
}

func ParseRefreshToken(raw string) (*Claims, error) {
	return parseToken(raw, "refresh")
}

func parseToken(raw, wantType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid && claims.Type == wantType {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
