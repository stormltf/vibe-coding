package middleware

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Timeout  time.Duration // 超时时间
	Response interface{}   // 超时响应
}

// DefaultTimeoutConfig 默认配置
func DefaultTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		Timeout: 30 * time.Second,
		Response: map[string]interface{}{
			"code":    4008,
			"message": "request timeout",
		},
	}
}

// Timeout 请求超时中间件
// 注意：此中间件通过 context 传递超时信号，业务代码需要检查 ctx.Done() 来响应超时
// 超时后会立即返回响应，但底层 handler 可能仍在执行（需要业务代码配合检查 context）
func Timeout(cfg *TimeoutConfig) app.HandlerFunc {
	if cfg == nil {
		cfg = DefaultTimeoutConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// 创建带超时的 context
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()

		// 用于通知处理完成
		done := make(chan struct{})
		// 原子标记：是否已超时
		var timedOut int32

		go func() {
			defer func() {
				// 防止 panic 导致 goroutine 泄露，recovery 中间件会处理实际的 panic
				_ = recover()
				close(done)
			}()

			// 执行下一个 handler
			c.Next(ctx)
		}()

		select {
		case <-done:
			// 正常完成
			return
		case <-ctx.Done():
			// 标记超时
			atomic.StoreInt32(&timedOut, 1)

			// 超时处理
			if ctx.Err() == context.DeadlineExceeded {
				// 立即返回超时响应（不等待 goroutine）
				// 注意：后台 goroutine 可能继续执行，但通过 context 已发送取消信号
				c.AbortWithStatusJSON(http.StatusRequestTimeout, cfg.Response)
			}
			return
		}
	}
}

// TimeoutWithDuration 使用指定时间的超时中间件
func TimeoutWithDuration(timeout time.Duration) app.HandlerFunc {
	return Timeout(&TimeoutConfig{
		Timeout: timeout,
		Response: map[string]interface{}{
			"code":    4008,
			"message": "request timeout",
		},
	})
}

// IsTimeout 检查 context 是否已超时（供业务代码使用）
func IsTimeout(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return ctx.Err() == context.DeadlineExceeded
	default:
		return false
	}
}
