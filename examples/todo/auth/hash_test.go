package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple password", "Password123", false},
		{"empty password", "", false},
		{"long password", string(make([]byte, 72)), false},
		{"too long password", string(make([]byte, 73)), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.input)

			if (err != nil) != tc.wantErr {
				t.Fatalf("HashPassword(%q) unexpected error = %v", tc.input, err)
			}
			if tc.wantErr {
				return
			}

			if len(hash) == 0 {
				t.Errorf("Expected non-empty hash for %q", tc.input)
			}
			if hash == tc.input {
				t.Errorf(`HashPassword(%q) = %q, did not expect to equal input`, tc.input, hash)
			}
		})
	}

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		hash1, _ := HashPassword("Password123")
		hash2, _ := HashPassword("password123")
		if hash1 == hash2 {
			t.Errorf("Hashes should differ for different inputs")
		}
	})
}

func TestComparePassword(t *testing.T) {
	hash, err := HashPassword("CorrectHorseBatteryStaple")
	if err != nil {
		t.Fatalf("unexpected error while hashing comparison password: %s", err)
	}

	tests := []struct {
		name     string
		password string
		hashed   string
		want     bool
	}{
		{"matching password", "CorrectHorseBatteryStaple", hash, true},
		{"wrong password", "WrongHorse", hash, false},
		{"empty password", "", hash, false},
		{"empty hash", "CorrectHorseBatteryStaple", "", false},
		{"garbage hash", "whatever", "not-a-valid-hash", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComparePassword(tc.hashed, tc.password)
			if got != tc.want {
				t.Errorf("ComparePassword(...) = %v, want %v", got, tc.want)
			}
		})
	}
}
