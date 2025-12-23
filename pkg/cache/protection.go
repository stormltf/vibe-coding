package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"golang.org/x/sync/singleflight"
)

var (
	ErrNotFound    = errors.New("key not found")
	ErrBloomFilter = errors.New("key does not exist (bloom filter)")
)

// ProtectedCache 带保护的缓存
// 1. 布隆过滤器：防止缓存穿透
// 2. Singleflight：防止缓存击穿
// 3. 空值缓存：防止穿透攻击
type ProtectedCache struct {
	bloom       *bloom.BloomFilter
	sf          singleflight.Group
	nullCache   sync.Map // 空值缓存
	nullTTL     time.Duration
	bloomMu     sync.RWMutex
	enableBloom bool
}

// ProtectionConfig 保护配置
type ProtectionConfig struct {
	EnableBloom    bool          // 启用布隆过滤器
	BloomSize      uint          // 布隆过滤器预期元素数量
	BloomFalseRate float64       // 布隆过滤器误判率
	NullCacheTTL   time.Duration // 空值缓存过期时间
}

// DefaultProtectionConfig 默认配置
func DefaultProtectionConfig() *ProtectionConfig {
	return &ProtectionConfig{
		EnableBloom:    true,
		BloomSize:      100000,
		BloomFalseRate: 0.01,
		NullCacheTTL:   1 * time.Minute,
	}
}

// NewProtectedCache 创建带保护的缓存
func NewProtectedCache(cfg *ProtectionConfig) *ProtectedCache {
	if cfg == nil {
		cfg = DefaultProtectionConfig()
	}

	pc := &ProtectedCache{
		nullTTL:     cfg.NullCacheTTL,
		enableBloom: cfg.EnableBloom,
	}

	if cfg.EnableBloom {
		pc.bloom = bloom.NewWithEstimates(cfg.BloomSize, cfg.BloomFalseRate)
	}

	return pc
}

// nullCacheItem 空值缓存项
type nullCacheItem struct {
	expireAt time.Time
}

// AddToBloom 将 key 添加到布隆过滤器
func (p *ProtectedCache) AddToBloom(key string) {
	if p.bloom == nil {
		return
	}
	p.bloomMu.Lock()
	p.bloom.AddString(key)
	p.bloomMu.Unlock()
}

// MightExist 检查 key 是否可能存在（布隆过滤器）
func (p *ProtectedCache) MightExist(key string) bool {
	if p.bloom == nil {
		return true
	}
	p.bloomMu.RLock()
	defer p.bloomMu.RUnlock()
	return p.bloom.TestString(key)
}

// IsNullCached 检查是否是空值缓存
func (p *ProtectedCache) IsNullCached(key string) bool {
	if item, ok := p.nullCache.Load(key); ok {
		ni := item.(*nullCacheItem)
		if time.Now().Before(ni.expireAt) {
			return true
		}
		// 过期了，删除
		p.nullCache.Delete(key)
	}
	return false
}

// SetNullCache 设置空值缓存
func (p *ProtectedCache) SetNullCache(key string) {
	p.nullCache.Store(key, &nullCacheItem{
		expireAt: time.Now().Add(p.nullTTL),
	})
}

// Get 带保护的缓存获取
// loader 是当缓存未命中时的加载函数
func (p *ProtectedCache) Get(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
	// 1. 检查布隆过滤器（如果启用）
	if p.enableBloom && !p.MightExist(key) {
		return nil, ErrBloomFilter
	}

	// 2. 检查空值缓存
	if p.IsNullCached(key) {
		return nil, ErrNotFound
	}

	// 3. 使用 singleflight 防止击穿
	result, err, _ := p.sf.Do(key, func() (interface{}, error) {
		// 再次检查空值缓存（双重检查）
		if p.IsNullCached(key) {
			return nil, ErrNotFound
		}

		data, err := loader()
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// 设置空值缓存，防止穿透
				p.SetNullCache(key)
			}
			return nil, err
		}

		// 添加到布隆过滤器
		p.AddToBloom(key)

		return data, nil
	})

	return result, err
}

// 全局保护缓存实例
var (
	defaultProtectedCache *ProtectedCache
	protectedCacheOnce    sync.Once
)

// GetProtectedCache 获取全局保护缓存
func GetProtectedCache() *ProtectedCache {
	protectedCacheOnce.Do(func() {
		defaultProtectedCache = NewProtectedCache(nil)
	})
	return defaultProtectedCache
}

// InitProtectedCache 初始化全局保护缓存
func InitProtectedCache(cfg *ProtectionConfig) {
	defaultProtectedCache = NewProtectedCache(cfg)
}
