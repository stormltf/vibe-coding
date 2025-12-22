package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
)

// LocalCache 本地缓存（基于 ristretto）
// 特点:
// - 高性能: 比 sync.Map 快 10x
// - 内存控制: 自动 LRU 淘汰
// - 并发安全: 无锁设计
// - TTL 支持: 自动过期
type LocalCache struct {
	cache *ristretto.Cache
}

// LocalCacheConfig 本地缓存配置
type LocalCacheConfig struct {
	MaxCost     int64 // 最大内存成本（字节）
	NumCounters int64 // 计数器数量（建议: MaxCost * 10）
	BufferItems int64 // 缓冲区大小
}

// DefaultLocalCacheConfig 默认配置（64MB 缓存）
func DefaultLocalCacheConfig() *LocalCacheConfig {
	return &LocalCacheConfig{
		MaxCost:     64 << 20, // 64MB
		NumCounters: 1e7,      // 1000 万个计数器
		BufferItems: 64,       // 缓冲区大小
	}
}

var localCache *LocalCache

// InitLocalCache 初始化本地缓存
func InitLocalCache(cfg *LocalCacheConfig) error {
	if cfg == nil {
		cfg = DefaultLocalCacheConfig()
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
		Metrics:     true, // 启用指标收集
	})
	if err != nil {
		return err
	}

	localCache = &LocalCache{cache: cache}
	return nil
}

// GetLocalCache 获取本地缓存实例
func GetLocalCache() *LocalCache {
	return localCache
}

// Get 获取缓存值
func (c *LocalCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Set 设置缓存值（不过期）
func (c *LocalCache) Set(key string, value interface{}, cost int64) bool {
	return c.cache.Set(key, value, cost)
}

// SetWithTTL 设置缓存值（带过期时间）
func (c *LocalCache) SetWithTTL(key string, value interface{}, cost int64, ttl time.Duration) bool {
	return c.cache.SetWithTTL(key, value, cost, ttl)
}

// Del 删除缓存
func (c *LocalCache) Del(key string) {
	c.cache.Del(key)
}

// Clear 清空缓存
func (c *LocalCache) Clear() {
	c.cache.Clear()
}

// Wait 等待所有 Set 操作完成（用于测试）
func (c *LocalCache) Wait() {
	c.cache.Wait()
}

// Metrics 获取缓存指标
func (c *LocalCache) Metrics() *ristretto.Metrics {
	return c.cache.Metrics
}

// Close 关闭缓存
func (c *LocalCache) Close() {
	c.cache.Close()
}

// MultiLevelCache 多级缓存
// L1: LocalCache (本地内存，最快)
// L2: Redis (分布式，次快)
// L3: Database (持久化，最慢)
type MultiLevelCache struct {
	local *LocalCache
	// Redis 使用全局 RDB
}

// NewMultiLevelCache 创建多级缓存
func NewMultiLevelCache() *MultiLevelCache {
	return &MultiLevelCache{
		local: localCache,
	}
}

// Get 多级缓存获取
// 优先级: LocalCache -> Redis
func (m *MultiLevelCache) Get(ctx context.Context, key string) (string, bool) {
	// L1: 本地缓存
	if m.local != nil {
		if val, ok := m.local.Get(key); ok {
			if s, ok := val.(string); ok {
				return s, true
			}
		}
	}

	// L2: Redis
	if RDB != nil {
		val, err := Get(ctx, key)
		if err == nil && val != "" {
			// 回填本地缓存
			if m.local != nil {
				m.local.SetWithTTL(key, val, int64(len(val)), time.Minute)
			}
			return val, true
		}
	}

	return "", false
}

// Set 多级缓存设置
func (m *MultiLevelCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	// L1: 本地缓存（TTL 短一些，避免数据不一致）
	localTTL := ttl / 2
	if localTTL < time.Second {
		localTTL = time.Second
	}
	if m.local != nil {
		m.local.SetWithTTL(key, value, int64(len(value)), localTTL)
	}

	// L2: Redis
	if RDB != nil {
		return Set(ctx, key, value, ttl)
	}

	return nil
}

// Del 多级缓存删除
func (m *MultiLevelCache) Del(ctx context.Context, keys ...string) error {
	// L1: 本地缓存
	if m.local != nil {
		for _, key := range keys {
			m.local.Del(key)
		}
	}

	// L2: Redis
	if RDB != nil {
		return Del(ctx, keys...)
	}

	return nil
}

// 全局多级缓存实例
var multiCache *MultiLevelCache

// InitMultiLevelCache 初始化多级缓存
func InitMultiLevelCache() {
	multiCache = NewMultiLevelCache()
}

// GetMultiCache 获取多级缓存
func GetMultiCache() *MultiLevelCache {
	return multiCache
}
