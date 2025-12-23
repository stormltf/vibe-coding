package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	"github.com/test-tt/pkg/logger"
)

const RequestIDKey = "X-Request-ID"

// RequestID 为每个请求生成唯一 ID，并注入到 context 中
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		requestID := string(c.GetHeader(RequestIDKey))
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Response.Header.Set(RequestIDKey, requestID)

		// 将 logid 注入到 context 中，便于日志追踪
		ctx = logger.ContextWithLogID(ctx, requestID)
		c.Next(ctx)
	}
}

// GetRequestID 从 RequestContext 获取请求 ID
func GetRequestID(c *app.RequestContext) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
