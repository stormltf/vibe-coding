package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &Config{
			Secret:     "test-secret",
			Issuer:     "test-issuer",
			ExpireTime: time.Hour,
		}
		j := New(cfg)
		if j.config != cfg {
			t.Error("expected config to be set")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		j := New(nil)
		if j.config == nil {
			t.Error("expected default config")
		}
		if j.config.Issuer != "test-tt" {
			t.Errorf("expected default issuer, got %s", j.config.Issuer)
		}
	})
}

func TestGenerateAndParseToken(t *testing.T) {
	j := New(&Config{
		Secret:     "test-secret-key-for-testing",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	userID := uint64(12345)
	username := "testuser"

	// Generate token
	token, err := j.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Parse token
	claims, err := j.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}
	if claims.Issuer != "test" {
		t.Errorf("Issuer = %v, want %v", claims.Issuer, "test")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	j := New(&Config{
		Secret:     "test-secret",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	tests := []struct {
		name      string
		token     string
		wantError error
	}{
		{"empty token", "", ErrTokenMalformed},
		{"malformed token", "not.a.valid.token", ErrTokenMalformed},
		{"random string", "randomstring", ErrTokenMalformed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := j.ParseToken(tt.token)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	j1 := New(&Config{
		Secret:     "secret-1",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	j2 := New(&Config{
		Secret:     "secret-2",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	// Generate token with j1
	token, err := j1.GenerateToken(1, "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Try to parse with j2 (different secret)
	_, err = j2.ParseToken(token)
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	j := New(&Config{
		Secret:     "test-secret",
		Issuer:     "test",
		ExpireTime: -time.Hour, // Already expired
	})

	token, err := j.GenerateToken(1, "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = j.ParseToken(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	j := New(&Config{
		Secret:     "test-secret",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	// Generate initial token
	token1, err := j.GenerateToken(1, "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Wait to ensure different issued time (JWT uses second precision)
	time.Sleep(time.Second + 100*time.Millisecond)

	// Refresh token
	token2, err := j.RefreshToken(token1)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	// Tokens may be identical if generated in same second, so we just verify
	// the refreshed token is valid and has correct user info
	claims, err := j.ParseToken(token2)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("UserID = %v, want 1", claims.UserID)
	}
	if claims.Username != "user" {
		t.Errorf("Username = %v, want 'user'", claims.Username)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	j := New(&Config{
		Secret:     "test-secret",
		Issuer:     "test",
		ExpireTime: time.Hour,
	})

	_, err := j.RefreshToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Secret == "" {
		t.Error("expected non-empty secret")
	}
	if cfg.Issuer != "test-tt" {
		t.Errorf("Issuer = %v, want 'test-tt'", cfg.Issuer)
	}
	if cfg.ExpireTime != 24*time.Hour {
		t.Errorf("ExpireTime = %v, want 24h", cfg.ExpireTime)
	}
}
