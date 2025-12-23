package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
)

func newTestEngine() *route.Engine {
	opt := config.NewOptions([]config.Option{})
	return route.NewEngine(opt)
}

// TestRequestID 测试请求 ID 中间件
func TestRequestID(t *testing.T) {
	r := newTestEngine()
	r.Use(RequestID())
	r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
		requestID := GetRequestID(c)
		c.String(http.StatusOK, requestID)
	})

	t.Run("generates new request ID", func(t *testing.T) {
		w := ut.PerformRequest(r, http.MethodGet, "/test", nil)
		assert.DeepEqual(t, http.StatusOK, w.Code)

		// Response should have X-Request-ID header
		reqID := w.Header().Get(RequestIDKey)
		assert.True(t, reqID != "")
		assert.True(t, len(reqID) == 36) // UUID format
	})

	t.Run("uses existing request ID", func(t *testing.T) {
		customID := "custom-request-id-12345"
		w := ut.PerformRequest(r, http.MethodGet, "/test", nil,
			ut.Header{Key: RequestIDKey, Value: customID})

		assert.DeepEqual(t, http.StatusOK, w.Code)
		assert.DeepEqual(t, customID, w.Header().Get(RequestIDKey))
		assert.DeepEqual(t, customID, w.Body.String())
	})
}

// Note: TestRecovery requires logger initialization which is not available in unit tests.
// The Recovery middleware is tested through integration tests instead.

// TestCORS 测试 CORS 中间件
func TestCORS(t *testing.T) {
	t.Run("default config blocks unknown origins", func(t *testing.T) {
		r := newTestEngine()
		r.Use(CORS()) // 默认安全配置
		r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
			c.String(http.StatusOK, "ok")
		})

		w := ut.PerformRequest(r, http.MethodGet, "/test", nil,
			ut.Header{Key: "Origin", Value: "http://example.com"})

		assert.DeepEqual(t, http.StatusOK, w.Code)
		// 默认配置不设置 CORS 头（安全模式）
		assert.DeepEqual(t, "", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("dev config allows localhost", func(t *testing.T) {
		r := newTestEngine()
		r.Use(CORSWithConfig(DevCORSConfig()))
		r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
			c.String(http.StatusOK, "ok")
		})

		w := ut.PerformRequest(r, http.MethodGet, "/test", nil,
			ut.Header{Key: "Origin", Value: "http://localhost:3000"})

		assert.DeepEqual(t, http.StatusOK, w.Code)
		assert.DeepEqual(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("handles preflight for allowed origin", func(t *testing.T) {
		r := newTestEngine()
		r.Use(CORSWithConfig(DevCORSConfig()))
		r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
			c.String(http.StatusOK, "ok")
		})

		w := ut.PerformRequest(r, http.MethodOptions, "/test", nil,
			ut.Header{Key: "Origin", Value: "http://localhost:8080"},
			ut.Header{Key: "Access-Control-Request-Method", Value: "POST"})

		assert.DeepEqual(t, http.StatusNoContent, w.Code)
	})

	t.Run("blocks preflight for disallowed origin", func(t *testing.T) {
		r := newTestEngine()
		r.Use(CORS()) // 默认安全配置
		r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
			c.String(http.StatusOK, "ok")
		})

		w := ut.PerformRequest(r, http.MethodOptions, "/test", nil,
			ut.Header{Key: "Origin", Value: "http://malicious.com"},
			ut.Header{Key: "Access-Control-Request-Method", Value: "POST"})

		assert.DeepEqual(t, http.StatusForbidden, w.Code)
	})
}

// TestTimeout 测试超时中间件
func TestTimeout(t *testing.T) {
	r := newTestEngine()
	r.Use(Timeout(&TimeoutConfig{Timeout: 100 * time.Millisecond}))
	r.GET("/fast", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("fast request succeeds", func(t *testing.T) {
		w := ut.PerformRequest(r, http.MethodGet, "/fast", nil)
		assert.DeepEqual(t, http.StatusOK, w.Code)
	})

	// Note: Timeout behavior for slow requests requires integration testing
	// as ut.PerformRequest uses synchronous HTTP handling
}

// TestTimeoutWithNilConfig 测试 nil 配置
func TestTimeoutWithNilConfig(t *testing.T) {
	r := newTestEngine()
	r.Use(Timeout(nil)) // Should use default config
	r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "ok")
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)
	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// TestDefaultTimeoutConfig 测试默认超时配置
func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	assert.DeepEqual(t, 30*time.Second, cfg.Timeout)
	assert.NotNil(t, cfg.Response)
}

// TestIsTimeout 测试超时检查函数
func TestIsTimeout(t *testing.T) {
	t.Run("not timed out", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, IsTimeout(ctx))
	})

	t.Run("timed out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)
		assert.True(t, IsTimeout(ctx))
	})
}

// TestTimeoutWithDuration 测试使用 Duration 的超时中间件
func TestTimeoutWithDuration(t *testing.T) {
	r := newTestEngine()
	r.Use(TimeoutWithDuration(50 * time.Millisecond))
	r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "ok")
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)
	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// TestAccessLogConfig 测试访问日志配置
func TestAccessLogConfig(t *testing.T) {
	cfg := DefaultAccessLogConfig()

	assert.DeepEqual(t, 1.0, cfg.SampleRate)
	assert.DeepEqual(t, time.Second, cfg.SlowThreshold)
	assert.True(t, len(cfg.SkipPaths) > 0)
}

// TestRateLimiterConfig 测试限流配置
func TestRateLimiterConfig(t *testing.T) {
	cfg := DefaultRateLimiterConfig()

	assert.True(t, cfg.Rate > 0)
	assert.True(t, cfg.Burst > 0)
}

// TestIPRateLimiter 测试 IP 限流器
func TestIPRateLimiter(t *testing.T) {
	cfg := &RateLimiterConfig{
		Rate:  10,
		Burst: 20,
	}
	limiter := NewIPRateLimiter(cfg)
	defer limiter.Stop()

	t.Run("creates limiter for IP", func(t *testing.T) {
		l := limiter.GetLimiter("192.168.1.1")
		assert.NotNil(t, l)
	})

	t.Run("returns same limiter for same IP", func(t *testing.T) {
		l1 := limiter.GetLimiter("192.168.1.2")
		l2 := limiter.GetLimiter("192.168.1.2")
		assert.True(t, l1 == l2)
	})

	t.Run("tracks size correctly", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			limiter.GetLimiter("10.0.0." + string(rune('0'+i)))
		}
		assert.True(t, limiter.Size() >= 5)
	})
}

// TestIPRateLimiterWithNilConfig 测试 nil 配置
func TestIPRateLimiterWithNilConfig(t *testing.T) {
	limiter := NewIPRateLimiter(nil)
	defer limiter.Stop()

	l := limiter.GetLimiter("192.168.1.1")
	assert.NotNil(t, l)
}

// TestGetRequestIDEmpty 测试空请求 ID
func TestGetRequestIDEmpty(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(ctx context.Context, c *app.RequestContext) {
		requestID := GetRequestID(c)
		c.String(http.StatusOK, requestID)
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)
	assert.DeepEqual(t, http.StatusOK, w.Code)
	assert.DeepEqual(t, "", w.Body.String())
}
