package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	// AllowedOrigins 允许的来源列表，支持通配符如 "*.example.com"
	// 空列表表示禁止所有跨域请求
	AllowedOrigins []string
	// AllowedMethods 允许的 HTTP 方法
	AllowedMethods []string
	// AllowedHeaders 允许的请求头
	AllowedHeaders []string
	// AllowCredentials 是否允许携带凭证
	AllowCredentials bool
	// MaxAge 预检请求缓存时间（秒）
	MaxAge int
}

// DefaultCORSConfig 默认 CORS 配置（安全模式）
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{}, // 默认不允许任何跨域
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// DevCORSConfig 开发环境 CORS 配置（宽松模式）
func DevCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// CORS 跨域中间件（使用默认安全配置）
func CORS() app.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 带配置的 CORS 中间件
func CORSWithConfig(cfg *CORSConfig) app.HandlerFunc {
	if cfg == nil {
		cfg = DefaultCORSConfig()
	}

	// 预处理配置
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.GetHeader("Origin"))

		// 检查 origin 是否在允许列表中
		allowed := false
		matchedOrigin := ""

		if origin != "" {
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if matchOrigin(origin, allowedOrigin) {
					allowed = true
					matchedOrigin = origin
					break
				}
			}
		}

		// 如果不允许，不设置 CORS 头
		if !allowed {
			if string(c.Method()) == "OPTIONS" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next(ctx)
			return
		}

		// 设置 CORS 响应头
		c.Response.Header.Set("Access-Control-Allow-Origin", matchedOrigin)
		c.Response.Header.Set("Access-Control-Allow-Methods", methods)
		c.Response.Header.Set("Access-Control-Allow-Headers", headers)
		c.Response.Header.Set("Access-Control-Max-Age", string(rune(cfg.MaxAge)))

		if cfg.AllowCredentials {
			c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		}

		// 添加 Vary 头，确保代理正确缓存
		c.Response.Header.Add("Vary", "Origin")

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next(ctx)
	}
}

// matchOrigin 检查 origin 是否匹配模式
// 支持通配符: "*.example.com" 匹配 "sub.example.com"
// 支持端口通配符: "http://localhost:*" 匹配 "http://localhost:3000"
func matchOrigin(origin, pattern string) bool {
	// 精确匹配
	if origin == pattern {
		return true
	}

	// 端口通配符匹配 (如 http://localhost:*)
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(origin, prefix) {
			// 确保后面是数字（端口）
			rest := strings.TrimPrefix(origin, prefix)
			if rest != "" && isNumeric(rest) {
				return true
			}
		}
	}

	// 子域名通配符匹配 (如 *.example.com)
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		if strings.HasSuffix(origin, suffix) {
			return true
		}
	}

	return false
}

// isNumeric 检查字符串是否全为数字
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
