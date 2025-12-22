package middleware

import (
	"context"
	"time"
	"unsafe"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/pkg/logger"
)

// AccessLog 记录请求日志
func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		c.Next(ctx)

		// 使用零拷贝转换减少内存分配
		logger.InfoCtxf(ctx, "access",
			"status", c.Response.StatusCode(),
			"method", b2s(c.Method()),
			"path", b2s(c.Path()),
			"latency", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	}
}

// b2s converts byte slice to string without memory allocation
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
