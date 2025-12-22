package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/pkg/cache"
	"golang.org/x/time/rate"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	Rate  rate.Limit // 每秒允许的请求数
	Burst int        // 突发请求数
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		Rate:  100, // 每秒100个请求
		Burst: 200, // 最大突发200个
	}
}

// IPRateLimiter 基于 IP 的限流器
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     sync.RWMutex
	config *RateLimiterConfig
}

// NewIPRateLimiter 创建 IP 限流器
func NewIPRateLimiter(config *RateLimiterConfig) *IPRateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}
	return &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		config: config,
	}
}

// GetLimiter 获取或创建 IP 对应的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.config.Rate, i.config.Burst)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimit 限流中间件
func RateLimit(config *RateLimiterConfig) app.HandlerFunc {
	limiter := NewIPRateLimiter(config)

	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    4029,
				"message": "too many requests",
			})
			return
		}
		c.Next(ctx)
	}
}

// GlobalRateLimit 全局限流中间件（不区分 IP）
func GlobalRateLimit(r rate.Limit, burst int) app.HandlerFunc {
	limiter := rate.NewLimiter(r, burst)

	return func(ctx context.Context, c *app.RequestContext) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    4029,
				"message": "too many requests",
			})
			return
		}
		c.Next(ctx)
	}
}

// DistributedRateLimit 分布式限流中间件（无状态服务使用）
// 使用 Redis 存储限流状态，支持多实例部署
func DistributedRateLimit(limiter *cache.DistributedRateLimiter) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		allowed, err := limiter.Allow(ctx, ip)
		if err != nil {
			// Redis 出错时放行，避免服务不可用
			c.Next(ctx)
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    4029,
				"message": "too many requests",
			})
			return
		}
		c.Next(ctx)
	}
}

// DistributedTokenBucketLimit 分布式令牌桶限流中间件
func DistributedTokenBucketLimit(limiter *cache.TokenBucketLimiter) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		allowed, err := limiter.Allow(ctx, ip)
		if err != nil {
			// Redis 出错时放行
			c.Next(ctx)
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    4029,
				"message": "too many requests",
			})
			return
		}
		c.Next(ctx)
	}
}
