package middleware

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/response"
)

// Recovery 捕获 panic 并返回 500 错误
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())

				// 获取请求详细信息
				method := string(c.Method())
				path := string(c.Path())
				query := string(c.URI().QueryString())
				ip := c.ClientIP()
				userAgent := string(c.UserAgent())
				requestID := string(c.GetHeader("X-Request-ID"))

				// 记录完整的错误信息
				logger.Errorf("panic recovered",
					"error", err,
					"method", method,
					"path", path,
					"query", query,
					"ip", ip,
					"user_agent", userAgent,
					"request_id", requestID,
					"stack", stack,
				)

				// 只有在响应未开始时才写入错误响应
				if !c.Response.HasBodyBytes() {
					c.Abort()
					response.ErrorWithStatus(c, http.StatusInternalServerError, 500, "Internal Server Error")
				} else {
					// 响应已开始，只能中止
					c.Abort()
				}
			}
		}()
		c.Next(ctx)
	}
}
