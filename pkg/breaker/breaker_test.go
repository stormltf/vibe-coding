package breaker

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &Config{
			Name:         "test",
			MaxRequests:  3,
			Interval:     5 * time.Second,
			Timeout:      10 * time.Second,
			FailureRatio: 0.6,
			MinRequests:  5,
		}
		cb := New(cfg)
		if cb == nil {
			t.Error("expected non-nil circuit breaker")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		cb := New(nil)
		if cb == nil {
			t.Error("expected non-nil circuit breaker with default config")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("test-breaker")

	if cfg.Name != "test-breaker" {
		t.Errorf("Name = %v, want 'test-breaker'", cfg.Name)
	}
	if cfg.MaxRequests != 5 {
		t.Errorf("MaxRequests = %v, want 5", cfg.MaxRequests)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("Interval = %v, want 10s", cfg.Interval)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.FailureRatio != 0.5 {
		t.Errorf("FailureRatio = %v, want 0.5", cfg.FailureRatio)
	}
	if cfg.MinRequests != 10 {
		t.Errorf("MinRequests = %v, want 10", cfg.MinRequests)
	}
}

func TestExecute_Success(t *testing.T) {
	cb := New(&Config{
		Name:         "test",
		MaxRequests:  5,
		Interval:     time.Second,
		Timeout:      time.Second,
		FailureRatio: 0.5,
		MinRequests:  3,
	})

	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result != "success" {
		t.Errorf("Execute() = %v, want 'success'", result)
	}
}

func TestExecute_Failure(t *testing.T) {
	cb := New(&Config{
		Name:         "test",
		MaxRequests:  5,
		Interval:     time.Second,
		Timeout:      time.Second,
		FailureRatio: 0.5,
		MinRequests:  3,
	})

	expectedErr := errors.New("test error")
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}
}

func TestIsOpen(t *testing.T) {
	cb := New(&Config{
		Name:         "test",
		MaxRequests:  1,
		Interval:     100 * time.Millisecond,
		Timeout:      100 * time.Millisecond,
		FailureRatio: 0.5,
		MinRequests:  2,
	})

	// Initially should be closed
	if cb.IsOpen() {
		t.Error("circuit breaker should be closed initially")
	}

	// Cause failures to trip the breaker
	testErr := errors.New("test error")
	for i := 0; i < 5; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, testErr
		})
	}

	// Should be open after failures
	if !cb.IsOpen() {
		t.Error("circuit breaker should be open after failures")
	}
}

func TestManager(t *testing.T) {
	m := NewManager(&Config{
		Name:         "default",
		MaxRequests:  5,
		Interval:     time.Second,
		Timeout:      time.Second,
		FailureRatio: 0.5,
		MinRequests:  10,
	})

	t.Run("get creates new breaker", func(t *testing.T) {
		cb := m.Get("test1")
		if cb == nil {
			t.Error("expected non-nil circuit breaker")
		}
	})

	t.Run("get returns same breaker", func(t *testing.T) {
		cb1 := m.Get("test2")
		cb2 := m.Get("test2")
		if cb1 != cb2 {
			t.Error("expected same circuit breaker instance")
		}
	})

	t.Run("different names create different breakers", func(t *testing.T) {
		cb1 := m.Get("test3")
		cb2 := m.Get("test4")
		if cb1 == cb2 {
			t.Error("expected different circuit breaker instances")
		}
	})

	t.Run("nil config uses default", func(t *testing.T) {
		m2 := NewManager(nil)
		cb := m2.Get("test")
		if cb == nil {
			t.Error("expected non-nil circuit breaker")
		}
	})
}

func TestManagerExecute(t *testing.T) {
	m := NewManager(nil)

	result, err := m.Execute("test", func() (interface{}, error) {
		return 42, nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result != 42 {
		t.Errorf("Execute() = %v, want 42", result)
	}
}

func TestGlobalBreaker(t *testing.T) {
	cb := GetBreaker("global-test")
	if cb == nil {
		t.Error("expected non-nil circuit breaker")
	}

	result, err := Execute("global-test-2", func() (interface{}, error) {
		return "ok", nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result != "ok" {
		t.Errorf("Execute() = %v, want 'ok'", result)
	}
}
