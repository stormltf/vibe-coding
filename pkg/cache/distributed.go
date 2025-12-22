package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedCache 分布式缓存（无状态服务使用）
// 所有状态存储在 Redis，支持多实例部署
type DistributedCache struct {
	rdb        *redis.Client
	nullPrefix string        // 空值缓存前缀
	nullTTL    time.Duration // 空值过期时间
	lockTTL    time.Duration // 分布式锁过期时间
}

// DistributedConfig 分布式缓存配置
type DistributedConfig struct {
	NullPrefix string
	NullTTL    time.Duration
	LockTTL    time.Duration
}

// DefaultDistributedConfig 默认配置
func DefaultDistributedConfig() *DistributedConfig {
	return &DistributedConfig{
		NullPrefix: "null:",
		NullTTL:    1 * time.Minute,
		LockTTL:    10 * time.Second,
	}
}

// NewDistributedCache 创建分布式缓存
func NewDistributedCache(rdb *redis.Client, cfg *DistributedConfig) *DistributedCache {
	if cfg == nil {
		cfg = DefaultDistributedConfig()
	}
	return &DistributedCache{
		rdb:        rdb,
		nullPrefix: cfg.NullPrefix,
		nullTTL:    cfg.NullTTL,
		lockTTL:    cfg.LockTTL,
	}
}

// IsNullCached 检查是否是空值缓存
func (d *DistributedCache) IsNullCached(ctx context.Context, key string) bool {
	exists, _ := d.rdb.Exists(ctx, d.nullPrefix+key).Result()
	return exists > 0
}

// SetNullCache 设置空值缓存
func (d *DistributedCache) SetNullCache(ctx context.Context, key string) error {
	return d.rdb.Set(ctx, d.nullPrefix+key, "1", d.nullTTL).Err()
}

// DeleteNullCache 删除空值缓存
func (d *DistributedCache) DeleteNullCache(ctx context.Context, key string) error {
	return d.rdb.Del(ctx, d.nullPrefix+key).Err()
}

// TryLock 尝试获取分布式锁（用于替代 singleflight）
func (d *DistributedCache) TryLock(ctx context.Context, key string) (bool, error) {
	lockKey := "lock:" + key
	return d.rdb.SetNX(ctx, lockKey, "1", d.lockTTL).Result()
}

// Unlock 释放分布式锁
func (d *DistributedCache) Unlock(ctx context.Context, key string) error {
	lockKey := "lock:" + key
	return d.rdb.Del(ctx, lockKey).Err()
}

// Get 带保护的缓存获取（分布式版本）
func (d *DistributedCache) Get(ctx context.Context, key string, loader func() (string, error)) (string, error) {
	// 1. 检查空值缓存
	if d.IsNullCached(ctx, key) {
		return "", ErrNotFound
	}

	// 2. 尝试从缓存获取
	val, err := d.rdb.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, redis.Nil) {
		return "", err
	}

	// 3. 尝试获取锁（防止缓存击穿）
	locked, err := d.TryLock(ctx, key)
	if err != nil {
		return "", err
	}

	if locked {
		defer d.Unlock(ctx, key)

		// 再次检查缓存（双重检查）
		val, err = d.rdb.Get(ctx, key).Result()
		if err == nil {
			return val, nil
		}

		// 加载数据
		data, err := loader()
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// 设置空值缓存
				d.SetNullCache(ctx, key)
			}
			return "", err
		}

		return data, nil
	}

	// 未获取到锁，等待一下重试
	time.Sleep(100 * time.Millisecond)
	return d.rdb.Get(ctx, key).Result()
}

// DistributedRateLimiter 分布式限流器
type DistributedRateLimiter struct {
	rdb    *redis.Client
	prefix string
	rate   int           // 每秒允许的请求数
	window time.Duration // 时间窗口
}

// NewDistributedRateLimiter 创建分布式限流器
func NewDistributedRateLimiter(rdb *redis.Client, rate int) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		rdb:    rdb,
		prefix: "ratelimit:",
		rate:   rate,
		window: time.Second,
	}
}

// Allow 检查是否允许请求（滑动窗口算法）
func (r *DistributedRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - int64(r.window)

	redisKey := r.prefix + key

	// 使用 Lua 脚本保证原子性
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])

		-- 移除窗口外的请求
		redis.call('ZREMRANGEBYSCORE', key, 0, window)

		-- 获取当前窗口内的请求数
		local count = redis.call('ZCARD', key)

		if count < limit then
			-- 添加当前请求
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, 2)
			return 1
		end

		return 0
	`)

	result, err := script.Run(ctx, r.rdb, []string{redisKey}, now, windowStart, r.rate).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// AllowN 检查是否允许 n 个请求
func (r *DistributedRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	for i := 0; i < n; i++ {
		allowed, err := r.Allow(ctx, key)
		if err != nil || !allowed {
			return false, err
		}
	}
	return true, nil
}

// TokenBucketLimiter 令牌桶限流器（分布式版本）
type TokenBucketLimiter struct {
	rdb      *redis.Client
	prefix   string
	rate     float64       // 每秒生成的令牌数
	capacity int64         // 桶容量
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rdb *redis.Client, rate float64, capacity int64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rdb:      rdb,
		prefix:   "tokenbucket:",
		rate:     rate,
		capacity: capacity,
	}
}

// Allow 检查是否允许请求
func (t *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := t.prefix + key
	now := float64(time.Now().UnixNano()) / 1e9

	// 使用 Lua 脚本实现令牌桶
	script := redis.NewScript(`
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local data = redis.call('HMGET', key, 'tokens', 'last')
		local tokens = tonumber(data[1]) or capacity
		local last = tonumber(data[2]) or now

		-- 计算新增的令牌数
		local delta = math.max(0, now - last)
		tokens = math.min(capacity, tokens + delta * rate)

		local allowed = 0
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
		end

		redis.call('HMSET', key, 'tokens', tokens, 'last', now)
		redis.call('EXPIRE', key, 60)

		return allowed
	`)

	result, err := script.Run(ctx, t.rdb, []string{redisKey}, t.rate, t.capacity, now).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// 使用分布式限流的中间件
func DistributedRateLimitMiddleware(limiter *DistributedRateLimiter) func(ctx context.Context, key string) bool {
	return func(ctx context.Context, key string) bool {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			// 限流器出错时放行，避免服务不可用
			return true
		}
		return allowed
	}
}

// 全局分布式缓存实例
var globalDistributedCache *DistributedCache

// InitDistributedCache 初始化全局分布式缓存
func InitDistributedCache(rdb *redis.Client, cfg *DistributedConfig) {
	globalDistributedCache = NewDistributedCache(rdb, cfg)
}

// GetDistributedCache 获取全局分布式缓存
func GetDistributedCache() *DistributedCache {
	return globalDistributedCache
}
