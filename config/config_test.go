package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Env: "dev",
		Server: &ServerConfig{
			Port: 8888,
		},
		MySQL: &MySQLConfig{
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: 30 * time.Minute,
		},
		Redis: &RedisConfig{
			PoolSize:     100,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		JWT: &JWTConfig{
			Secret: "a-very-long-secret-key-for-testing-purposes-32chars",
		},
		RateLimit: &RateLimitConfig{
			Rate:  100,
			Burst: 200,
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"port too high", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: &ServerConfig{Port: tt.port},
			}

			err := Validate(cfg)
			if err == nil {
				t.Error("expected error for invalid port")
			}
			if !strings.Contains(err.Error(), "server.port") {
				t.Errorf("error should mention server.port: %v", err)
			}
		})
	}
}

func TestValidate_MySQLConfig(t *testing.T) {
	t.Run("max_open_conns too high", func(t *testing.T) {
		cfg := &Config{
			MySQL: &MySQLConfig{
				MaxOpenConns: 501,
				MaxIdleConns: 10,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for max_open_conns > 500")
		}
	})

	t.Run("max_idle_conns exceeds max_open_conns", func(t *testing.T) {
		cfg := &Config{
			MySQL: &MySQLConfig{
				MaxOpenConns: 50,
				MaxIdleConns: 100,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error when max_idle_conns > max_open_conns")
		}
	})

	t.Run("negative conn_max_lifetime", func(t *testing.T) {
		cfg := &Config{
			MySQL: &MySQLConfig{
				MaxOpenConns:    50,
				MaxIdleConns:    10,
				ConnMaxLifetime: -1 * time.Second,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for negative conn_max_lifetime")
		}
	})
}

func TestValidate_RedisConfig(t *testing.T) {
	t.Run("pool_size too high", func(t *testing.T) {
		cfg := &Config{
			Redis: &RedisConfig{
				PoolSize:     1001,
				DialTimeout:  time.Second,
				ReadTimeout:  time.Second,
				WriteTimeout: time.Second,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for pool_size > 1000")
		}
	})

	t.Run("zero dial_timeout", func(t *testing.T) {
		cfg := &Config{
			Redis: &RedisConfig{
				PoolSize:     100,
				DialTimeout:  0,
				ReadTimeout:  time.Second,
				WriteTimeout: time.Second,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for zero dial_timeout")
		}
	})
}

func TestValidate_JWTConfig(t *testing.T) {
	t.Run("empty secret", func(t *testing.T) {
		cfg := &Config{
			JWT: &JWTConfig{
				Secret: "",
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for empty JWT secret")
		}
	})

	t.Run("default insecure secret in prod", func(t *testing.T) {
		cfg := &Config{
			Env: "prod",
			JWT: &JWTConfig{
				Secret: defaultInsecureSecret,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for default secret in production")
		}
	})

	t.Run("short secret in prod", func(t *testing.T) {
		cfg := &Config{
			Env: "prod",
			JWT: &JWTConfig{
				Secret: "short-secret",
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for short secret in production")
		}
	})

	t.Run("short secret in dev is ok", func(t *testing.T) {
		cfg := &Config{
			Env: "dev",
			JWT: &JWTConfig{
				Secret: "short-secret",
			},
		}

		err := Validate(cfg)
		if err != nil {
			t.Errorf("short secret should be allowed in dev: %v", err)
		}
	})
}

func TestValidate_RateLimitConfig(t *testing.T) {
	t.Run("zero rate", func(t *testing.T) {
		cfg := &Config{
			RateLimit: &RateLimitConfig{
				Rate:  0,
				Burst: 100,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for zero rate")
		}
	})

	t.Run("negative burst", func(t *testing.T) {
		cfg := &Config{
			RateLimit: &RateLimitConfig{
				Rate:  100,
				Burst: -1,
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected error for negative burst")
		}
	})
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"dev", true},
		{"", true},
		{"prod", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			if got := cfg.IsDev(); got != tt.want {
				t.Errorf("IsDev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProd(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"prod", true},
		{"dev", false},
		{"", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			if got := cfg.IsProd(); got != tt.want {
				t.Errorf("IsProd() = %v, want %v", got, tt.want)
			}
		})
	}
}
