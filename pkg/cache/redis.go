package cache

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	MaxRetries   int
}

// DefaultConfig 返回优化后的默认配置
func DefaultConfig() *Config {
	poolSize := runtime.GOMAXPROCS(0) * 10 // 每个 CPU 10 个连接
	if poolSize < 100 {
		poolSize = 100
	}
	return &Config{
		PoolSize:     poolSize,
		MinIdleConns: 20,              // 最小空闲连接
		MaxIdleConns: poolSize / 2,    // 最大空闲连接
		DialTimeout:  3 * time.Second, // 连接超时
		ReadTimeout:  1 * time.Second, // 读超时
		WriteTimeout: 1 * time.Second, // 写超时
		PoolTimeout:  2 * time.Second, // 获取连接超时
		MaxRetries:   3,               // 最大重试次数
	}
}

func Init(cfg *Config) error {
	// 合并默认配置
	defaults := DefaultConfig()
	if cfg.PoolSize == 0 {
		cfg.PoolSize = defaults.PoolSize
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = defaults.MinIdleConns
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = defaults.MaxIdleConns
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = defaults.DialTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaults.ReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaults.WriteTimeout
	}
	if cfg.PoolTimeout == 0 {
		cfg.PoolTimeout = defaults.PoolTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaults.MaxRetries
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池配置
		PoolSize:     cfg.PoolSize,     // 连接池大小
		MinIdleConns: cfg.MinIdleConns, // 最小空闲连接（保持热连接）
		MaxIdleConns: cfg.MaxIdleConns, // 最大空闲连接

		// 超时配置
		DialTimeout:  cfg.DialTimeout,  // 建立连接超时
		ReadTimeout:  cfg.ReadTimeout,  // 读超时
		WriteTimeout: cfg.WriteTimeout, // 写超时
		PoolTimeout:  cfg.PoolTimeout,  // 从连接池获取连接超时

		// 重试配置
		MaxRetries:      cfg.MaxRetries,         // 命令最大重试次数
		MinRetryBackoff: 8 * time.Millisecond,   // 最小重试间隔
		MaxRetryBackoff: 512 * time.Millisecond, // 最大重试间隔

		// 连接健康检查
		ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接最大存活时间
		ConnMaxLifetime: 30 * time.Minute, // 连接最大存活时间
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	return nil
}

// Stats 获取连接池统计信息
func Stats() *redis.PoolStats {
	if RDB == nil {
		return nil
	}
	return RDB.PoolStats()
}

func Close() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}

// 常用操作封装
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

func Del(ctx context.Context, keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, keys ...string) (int64, error) {
	return RDB.Exists(ctx, keys...).Result()
}

func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return RDB.Expire(ctx, key, expiration).Err()
}

// MGet 批量获取（减少网络往返）
func MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return RDB.MGet(ctx, keys...).Result()
}

// MSet 批量设置
func MSet(ctx context.Context, values ...interface{}) error {
	return RDB.MSet(ctx, values...).Err()
}

// Pipeline 管道操作（批量命令）
func Pipeline() redis.Pipeliner {
	return RDB.Pipeline()
}
