package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthTokenFlow(t *testing.T) {
	uid := "test-user-id"

	tokens, err := GenerateTokens(uid)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	tests := []struct {
		name         string
		rawToken     string
		parser       func(string) (*Claims, error)
		expectType   string
		expectErr    bool
		expectUserID string
	}{
		{
			name:         "valid access token",
			rawToken:     tokens.AccessToken,
			parser:       ParseAccessToken,
			expectType:   "access",
			expectErr:    false,
			expectUserID: uid,
		},
		{
			name:         "valid refresh token",
			rawToken:     tokens.RefreshToken,
			parser:       ParseRefreshToken,
			expectType:   "refresh",
			expectErr:    false,
			expectUserID: uid,
		},
		{
			name:       "access token parsed as refresh token",
			rawToken:   tokens.AccessToken,
			parser:     ParseRefreshToken,
			expectType: "refresh",
			expectErr:  true,
		},
		{
			name:       "refresh token parsed as access token",
			rawToken:   tokens.RefreshToken,
			parser:     ParseAccessToken,
			expectType: "access",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tt.parser(tt.rawToken)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims.Type != tt.expectType {
				t.Errorf("expected type %q, got %q", tt.expectType, claims.Type)
			}
			if claims.Subject != tt.expectUserID {
				t.Errorf("expected subject %q, got %q", tt.expectUserID, claims.Subject)
			}
			if claims.IsExpired() {
				t.Errorf("expected token to not be expired")
			}
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		want      bool
	}{
		{
			name:      "token already expired",
			expiresIn: -1 * time.Minute,
			want:      true,
		},
		{
			name:      "token not expired",
			expiresIn: 5 * time.Minute,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(tt.expiresIn)),
				},
			}
			got := claims.IsExpired()
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
