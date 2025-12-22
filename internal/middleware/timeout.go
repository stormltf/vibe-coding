package middleware

import (
	"context"
	"net/http"
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

		go func() {
			c.Next(ctx)
			close(done)
		}()

		select {
		case <-done:
			// 正常完成
			return
		case <-ctx.Done():
			// 超时
			if ctx.Err() == context.DeadlineExceeded {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, cfg.Response)
			}
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
