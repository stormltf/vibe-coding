package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"golang.org/x/time/rate"
)

// SecurityHeaders 安全响应头中间件
// 设置常见的安全相关 HTTP 响应头
func SecurityHeaders() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.URI().Path())

		// 静态文件和前端页面使用宽松的安全策略
		// 自动识别：根路径、.html 结尾、/static 开头、/swagger 开头
		isStaticOrPage := path == "/" || path == "/favicon.ico" ||
			(len(path) >= 5 && path[len(path)-5:] == ".html") ||
			(len(path) >= 7 && path[:7] == "/static") ||
			(len(path) >= 8 && path[:8] == "/swagger")

		// 防止 MIME 类型嗅探
		c.Response.Header.Set("X-Content-Type-Options", "nosniff")

		// 防止点击劫持
		c.Response.Header.Set("X-Frame-Options", "DENY")

		// XSS 保护（现代浏览器已内置，但仍建议设置）
		c.Response.Header.Set("X-XSS-Protection", "1; mode=block")

		// 引用来源策略
		c.Response.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if isStaticOrPage {
			// 前端页面：允许加载本站资源，允许连接到 Agent 服务
			c.Response.Header.Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' http://localhost:3001; frame-ancestors 'none'")
		} else {
			// API 端点：严格 CSP
			c.Response.Header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			// 缓存控制（仅 API 响应不缓存）
			c.Response.Header.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			c.Response.Header.Set("Pragma", "no-cache")
			c.Response.Header.Set("Expires", "0")
		}

		// 权限策略（禁用不需要的浏览器特性）
		c.Response.Header.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next(ctx)
	}
}

// HSTSMiddleware HSTS 中间件（仅用于 HTTPS 生产环境）
// maxAge: HSTS 有效期（秒），建议 31536000（1年）
func HSTSMiddleware(maxAge int) app.HandlerFunc {
	hstsValue := "max-age=" + string(rune(maxAge)) + "; includeSubDomains; preload"
	return func(ctx context.Context, c *app.RequestContext) {
		// 仅在 HTTPS 连接时设置 HSTS
		if string(c.URI().Scheme()) == "https" {
			c.Response.Header.Set("Strict-Transport-Security", hstsValue)
		}
		c.Next(ctx)
	}
}

// AuthRateLimiter 认证端点专用限流器
type AuthRateLimiter struct {
	ips     map[string]*authLimiterEntry
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
	ttl     time.Duration
	maxSize int
}

type authLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	authLimiter     *AuthRateLimiter
	authLimiterOnce sync.Once
)

// getAuthRateLimiter 获取认证限流器单例
func getAuthRateLimiter() *AuthRateLimiter {
	authLimiterOnce.Do(func() {
		authLimiter = &AuthRateLimiter{
			ips:     make(map[string]*authLimiterEntry),
			rate:    rate.Limit(10.0 / 60.0), // 每分钟 10 次
			burst:   5,                       // 突发 5 次
			ttl:     10 * time.Minute,
			maxSize: 10000,
		}
		go authLimiter.cleanup()
	})
	return authLimiter
}

// cleanup 定期清理过期条目
func (a *AuthRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()
		for ip, entry := range a.ips {
			if now.Sub(entry.lastSeen) > a.ttl {
				delete(a.ips, ip)
			}
		}
		a.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (a *AuthRateLimiter) Allow(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	if entry, exists := a.ips[ip]; exists {
		entry.lastSeen = now
		return entry.limiter.Allow()
	}

	// 检查是否需要清理
	if len(a.ips) >= a.maxSize {
		// 简单清理：删除最老的条目
		var oldestIP string
		var oldestTime time.Time
		for ip, entry := range a.ips {
			if oldestIP == "" || entry.lastSeen.Before(oldestTime) {
				oldestIP = ip
				oldestTime = entry.lastSeen
			}
		}
		if oldestIP != "" {
			delete(a.ips, oldestIP)
		}
	}

	// 创建新条目
	entry := &authLimiterEntry{
		limiter:  rate.NewLimiter(a.rate, a.burst),
		lastSeen: now,
	}
	a.ips[ip] = entry
	return entry.limiter.Allow()
}

// AuthRateLimit 认证端点限流中间件
// 专门用于登录、注册等认证端点，防止暴力破解
// 限制：每 IP 每分钟 10 次请求
func AuthRateLimit() app.HandlerFunc {
	limiter := getAuthRateLimiter()

	return func(ctx context.Context, c *app.RequestContext) {
		ip := GetRealClientIP(c)

		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    4029,
				"message": "too many authentication attempts, please try again later",
			})
			return
		}

		c.Next(ctx)
	}
}
