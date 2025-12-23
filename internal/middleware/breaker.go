package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/pkg/breaker"
	"github.com/test-tt/pkg/logger"
)

// ErrCircuitOpen 熔断器打开错误
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ErrServerError 服务器错误（用于触发熔断）
var ErrServerError = errors.New("server error")

// CircuitBreaker 熔断中间件
// 用于保护下游服务，当错误率过高时自动熔断
func CircuitBreaker(cb *breaker.CircuitBreaker) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 先检查熔断器状态，如果已打开则直接拒绝
		if cb.IsOpen() {
			logger.WarnCtxf(ctx, "circuit breaker is open, rejecting request")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
				"code":    5003,
				"message": "service temporarily unavailable (circuit open)",
			})
			return
		}

		// 执行请求
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next(ctx)

			// 5xx 错误视为失败，触发熔断计数
			if c.Response.StatusCode() >= 500 {
				return nil, ErrServerError
			}
			return nil, nil
		})

		// 如果执行过程中熔断器打开了（错误率达到阈值）
		if err != nil {
			if errors.Is(err, ErrCircuitOpen) || cb.IsOpen() {
				// 如果响应还没发送，则返回熔断响应
				if !c.Response.HasBodyBytes() {
					c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
						"code":    5003,
						"message": "service temporarily unavailable",
					})
				}
			}
			// 其他错误（如 ErrServerError）已经由 handler 处理了响应
		}
	}
}

// CircuitBreakerByPath 基于路径的熔断中间件
// 每个路径使用独立的熔断器
func CircuitBreakerByPath(manager *breaker.Manager) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())
		cb := manager.Get(path)

		// 先检查熔断器状态
		if cb.IsOpen() {
			logger.WarnCtxf(ctx, "circuit breaker is open for path", "path", path)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
				"code":    5003,
				"message": "service temporarily unavailable (circuit open)",
			})
			return
		}

		// 执行请求
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next(ctx)

			// 5xx 错误视为失败
			if c.Response.StatusCode() >= 500 {
				return nil, ErrServerError
			}
			return nil, nil
		})

		if err != nil {
			if errors.Is(err, ErrCircuitOpen) || cb.IsOpen() {
				if !c.Response.HasBodyBytes() {
					c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
						"code":    5003,
						"message": "service temporarily unavailable",
					})
				}
			}
		}
	}
}
