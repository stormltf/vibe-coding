package middleware

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/pkg/breaker"
)

// CircuitBreaker 熔断中间件
// 用于保护下游服务，当错误率过高时自动熔断
func CircuitBreaker(cb *breaker.CircuitBreaker) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next(ctx)

			// 5xx 错误视为失败
			if c.Response.StatusCode() >= 500 {
				return nil, http.ErrAbortHandler
			}
			return nil, nil
		})

		if err != nil {
			// 熔断器开启，返回服务不可用
			if cb.IsOpen() {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
					"code":    5003,
					"message": "service temporarily unavailable",
				})
				return
			}
		}
	}
}

// CircuitBreakerByPath 基于路径的熔断中间件
// 每个路径使用独立的熔断器
func CircuitBreakerByPath(manager *breaker.Manager) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())
		cb := manager.Get(path)

		_, err := cb.Execute(func() (interface{}, error) {
			c.Next(ctx)

			// 5xx 错误视为失败
			if c.Response.StatusCode() >= 500 {
				return nil, http.ErrAbortHandler
			}
			return nil, nil
		})

		if err != nil {
			if cb.IsOpen() {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, map[string]interface{}{
					"code":    5003,
					"message": "service temporarily unavailable",
				})
				return
			}
		}
	}
}
