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
				logger.Errorf("panic recovered",
					"error", err,
					"path", string(c.Path()),
					"stack", stack,
				)
				c.Abort()
				response.ErrorWithStatus(c, http.StatusInternalServerError, 500, "Internal Server Error")
			}
		}()
		c.Next(ctx)
	}
}
