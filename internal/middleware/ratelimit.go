package middleware

import (
	"container/list"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/logger"
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

// ipEntry 存储 IP 限流器和 LRU 信息
type ipEntry struct {
	ip       string
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter 基于 IP 的限流器（带 LRU 驱逐）
type IPRateLimiter struct {
	ips      map[string]*list.Element // IP -> LRU 链表元素
	lru      *list.List               // LRU 链表，头部是最近使用的
	mu       sync.Mutex
	config   *RateLimiterConfig
	maxSize  int           // 最大 IP 数量
	ttl      time.Duration // IP 过期时间
	stopChan chan struct{} // 停止清理 goroutine
}

const (
	defaultMaxIPCount = 10000            // 默认最大 IP 数量
	defaultIPTTL      = 10 * time.Minute // 默认 IP 过期时间
	cleanupInterval   = 1 * time.Minute  // 清理间隔
)

// NewIPRateLimiter 创建 IP 限流器
func NewIPRateLimiter(config *RateLimiterConfig) *IPRateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}
	limiter := &IPRateLimiter{
		ips:      make(map[string]*list.Element),
		lru:      list.New(),
		config:   config,
		maxSize:  defaultMaxIPCount,
		ttl:      defaultIPTTL,
		stopChan: make(chan struct{}),
	}

	// 启动后台清理 goroutine
	go limiter.cleanup()

	return limiter
}

// cleanup 定期清理过期的 IP
func (i *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.evictExpired()
		case <-i.stopChan:
			return
		}
	}
}

// evictExpired 驱逐过期的 IP
func (i *IPRateLimiter) evictExpired() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	evicted := 0

	// 从尾部开始检查（最老的）
	for e := i.lru.Back(); e != nil; {
		entry := e.Value.(*ipEntry)
		if now.Sub(entry.lastSeen) > i.ttl {
			prev := e.Prev()
			i.lru.Remove(e)
			delete(i.ips, entry.ip)
			evicted++
			e = prev
		} else {
			// 遇到未过期的就停止（因为是按时间排序的）
			break
		}
	}

	if evicted > 0 {
		logger.Infof("rate limiter evicted expired IPs", "count", evicted, "remaining", len(i.ips))
	}
}

// GetLimiter 获取或创建 IP 对应的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()

	if elem, exists := i.ips[ip]; exists {
		// 更新访问时间并移到头部
		entry := elem.Value.(*ipEntry)
		entry.lastSeen = now
		i.lru.MoveToFront(elem)
		return entry.limiter
	}

	// 检查是否需要驱逐
	for len(i.ips) >= i.maxSize {
		// 驱逐最老的（尾部）
		oldest := i.lru.Back()
		if oldest != nil {
			entry := oldest.Value.(*ipEntry)
			i.lru.Remove(oldest)
			delete(i.ips, entry.ip)
		}
	}

	// 创建新的限流器
	entry := &ipEntry{
		ip:       ip,
		limiter:  rate.NewLimiter(i.config.Rate, i.config.Burst),
		lastSeen: now,
	}
	elem := i.lru.PushFront(entry)
	i.ips[ip] = elem

	return entry.limiter
}

// Stop 停止后台清理
func (i *IPRateLimiter) Stop() {
	close(i.stopChan)
}

// Size 返回当前 IP 数量
func (i *IPRateLimiter) Size() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return len(i.ips)
}

// 全局限流器注册表，用于优雅关闭时清理
var (
	globalLimiters   []*IPRateLimiter
	globalLimitersMu sync.Mutex
)

// registerLimiter 注册限流器到全局注册表
func registerLimiter(l *IPRateLimiter) {
	globalLimitersMu.Lock()
	defer globalLimitersMu.Unlock()
	globalLimiters = append(globalLimiters, l)
}

// StopAllRateLimiters 停止所有限流器的后台清理 goroutine
// 应在服务关闭时调用
func StopAllRateLimiters() {
	globalLimitersMu.Lock()
	defer globalLimitersMu.Unlock()
	for _, l := range globalLimiters {
		l.Stop()
	}
	globalLimiters = nil
}

// RateLimit 限流中间件
func RateLimit(config *RateLimiterConfig) app.HandlerFunc {
	limiter := NewIPRateLimiter(config)
	registerLimiter(limiter)

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
