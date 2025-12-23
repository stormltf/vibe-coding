package middleware

import (
	"context"
	"math/rand"
	"time"
	"unsafe"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/pkg/logger"
)

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	// SampleRate 采样率 (0.0-1.0)，1.0 表示记录所有请求
	SampleRate float64
	// SlowThreshold 慢请求阈值，超过此时间总是记录
	SlowThreshold time.Duration
	// SkipPaths 跳过记录的路径（如健康检查）
	SkipPaths []string
}

// DefaultAccessLogConfig 默认配置
func DefaultAccessLogConfig() *AccessLogConfig {
	return &AccessLogConfig{
		SampleRate:    1.0,         // 默认记录所有
		SlowThreshold: time.Second, // 1秒以上视为慢请求
		SkipPaths:     []string{"/ping", "/health", "/metrics"},
	}
}

// AccessLog 记录请求日志（默认配置）
func AccessLog() app.HandlerFunc {
	return AccessLogWithConfig(nil)
}

// AccessLogWithConfig 带配置的访问日志中间件
func AccessLogWithConfig(cfg *AccessLogConfig) app.HandlerFunc {
	if cfg == nil {
		cfg = DefaultAccessLogConfig()
	}

	// 预处理跳过路径为 map 以提高查找效率
	skipPathsMap := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipPathsMap[p] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		c.Next(ctx)

		// 计算请求耗时
		latency := time.Since(start)
		path := b2s(c.Path())

		// 跳过特定路径
		if skipPathsMap[path] {
			return
		}

		// 慢请求总是记录
		isSlow := latency >= cfg.SlowThreshold

		// 采样判断：慢请求总是记录，否则按采样率记录
		// 注：这里使用 math/rand 而非 crypto/rand，因为日志采样不需要加密级别随机性
		if !isSlow && cfg.SampleRate < 1.0 {
			if rand.Float64() > cfg.SampleRate { //nolint:gosec // 日志采样不需要加密随机数
				return // 跳过此请求的日志
			}
		}

		// 根据状态码选择日志级别
		status := c.Response.StatusCode()
		logFunc := logger.InfoCtxf
		if status >= 500 {
			logFunc = logger.ErrorCtxf
		} else if status >= 400 {
			logFunc = logger.WarnCtxf
		}

		// 记录日志
		logFunc(ctx, "access",
			"status", status,
			"method", b2s(c.Method()),
			"path", path,
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"slow", isSlow,
		)
	}
}

// b2s converts byte slice to string without memory allocation
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
