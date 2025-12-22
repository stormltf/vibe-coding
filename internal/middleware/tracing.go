package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracing 链路追踪中间件
func Tracing(serviceName string) app.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(ctx context.Context, c *app.RequestContext) {

		// 创建 span
		spanName := string(c.Method()) + " " + string(c.Path())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", string(c.Method())),
				attribute.String("http.url", string(c.URI().RequestURI())),
				attribute.String("http.host", string(c.Host())),
				attribute.String("http.user_agent", string(c.UserAgent())),
				attribute.String("net.peer.ip", c.ClientIP()),
			),
		)
		defer span.End()

		// 将 trace ID 设置到响应头
		if span.SpanContext().HasTraceID() {
			c.Response.Header.Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		c.Next(ctx)

		// 记录响应状态
		statusCode := c.Response.StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		if statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// headerCarrier 实现 propagation.TextMapCarrier
type headerCarrier struct {
	c *app.RequestContext
}

func (h *headerCarrier) Get(key string) string {
	return string(h.c.GetHeader(key))
}

func (h *headerCarrier) Set(key string, value string) {
	h.c.Request.Header.Set(key, value)
}

func (h *headerCarrier) Keys() []string {
	var keys []string
	h.c.Request.Header.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})
	return keys
}
