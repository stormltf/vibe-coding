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
	t.Run("refresh too early", func(t *testing.T) {
		// Token 有 24 小时有效期，远超 MinRefreshWindow (2小时)
		j := New(&Config{
			Secret:     "test-secret",
			Issuer:     "test",
			ExpireTime: 24 * time.Hour,
		})

		token1, err := j.GenerateToken(1, "user")
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		// 尝试刷新（应该失败，因为剩余时间 24h > MinRefreshWindow 2h）
		_, err = j.RefreshToken(token1)
		if !errors.Is(err, ErrRefreshTooEarly) {
			t.Errorf("expected ErrRefreshTooEarly, got %v", err)
		}
	})

	t.Run("force refresh", func(t *testing.T) {
		j := New(&Config{
			Secret:     "test-secret",
			Issuer:     "test",
			ExpireTime: 24 * time.Hour,
		})

		token1, err := j.GenerateToken(1, "user")
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		// 强制刷新应该成功（不检查剩余时间）
		token2, err := j.ForceRefreshToken(token1)
		if err != nil {
			t.Fatalf("ForceRefreshToken() error = %v", err)
		}

		claims, err := j.ParseToken(token2)
		if err != nil {
			t.Fatalf("ParseToken() error = %v", err)
		}

		if claims.UserID != 1 {
			t.Errorf("UserID = %v, want 1", claims.UserID)
		}
	})

	t.Run("refresh near expiry", func(t *testing.T) {
		// Token 有 1 小时有效期，小于 MinRefreshWindow (2小时)
		j := New(&Config{
			Secret:     "test-secret",
			Issuer:     "test",
			ExpireTime: time.Hour,
		})

		token1, err := j.GenerateToken(1, "user")
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		// 这个 token 剩余时间 1h < MinRefreshWindow 2h，应该允许刷新
		token2, err := j.RefreshToken(token1)
		if err != nil {
			t.Fatalf("RefreshToken() error = %v", err)
		}

		claims, err := j.ParseToken(token2)
		if err != nil {
			t.Fatalf("ParseToken() error = %v", err)
		}

		if claims.UserID != 1 {
			t.Errorf("UserID = %v, want 1", claims.UserID)
		}
	})
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

	// 安全改进：默认配置返回空密钥，强制用户显式配置
	if cfg.Secret != "" {
		t.Error("expected empty secret (security: force explicit configuration)")
	}
	if cfg.Issuer != "test-tt" {
		t.Errorf("Issuer = %v, want 'test-tt'", cfg.Issuer)
	}
	if cfg.ExpireTime != 24*time.Hour {
		t.Errorf("ExpireTime = %v, want 24h", cfg.ExpireTime)
	}
}

func TestValidateConfig(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		err := ValidateConfig(nil)
		if !errors.Is(err, ErrSecretNotConfigured) {
			t.Errorf("expected ErrSecretNotConfigured, got %v", err)
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		cfg := &Config{Secret: ""}
		err := ValidateConfig(cfg)
		if !errors.Is(err, ErrSecretNotConfigured) {
			t.Errorf("expected ErrSecretNotConfigured, got %v", err)
		}
	})

	t.Run("short secret", func(t *testing.T) {
		cfg := &Config{Secret: "short"}
		err := ValidateConfig(cfg)
		if err == nil {
			t.Error("expected error for short secret")
		}
	})

	t.Run("valid secret", func(t *testing.T) {
		cfg := &Config{Secret: "this-is-a-very-long-secret-key-for-testing"}
		err := ValidateConfig(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
