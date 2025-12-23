package breaker

import (
	"errors"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

var (
	ErrCircuitOpen    = errors.New("circuit breaker is open")
	ErrTooManyRequest = errors.New("too many requests")
)

// Config 熔断器配置
type Config struct {
	Name         string        // 熔断器名称
	MaxRequests  uint32        // 半开状态下允许的最大请求数
	Interval     time.Duration // 统计周期
	Timeout      time.Duration // 熔断超时时间（从开启到半开）
	FailureRatio float64       // 触发熔断的失败率阈值
	MinRequests  uint32        // 触发熔断的最小请求数
}

// DefaultConfig 默认配置
func DefaultConfig(name string) *Config {
	return &Config{
		Name:         name,
		MaxRequests:  5,
		Interval:     10 * time.Second,
		Timeout:      30 * time.Second,
		FailureRatio: 0.5,
		MinRequests:  10,
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// New 创建熔断器
func New(cfg *Config) *CircuitBreaker {
	if cfg == nil {
		cfg = DefaultConfig("default")
	}

	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 请求数达到最小阈值且失败率超过设定值时触发熔断
			if counts.Requests < cfg.MinRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// 状态变化时可以记录日志或发送告警
		},
	}

	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute 执行带熔断保护的函数
func (c *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return c.cb.Execute(fn)
}

// State 获取当前状态
func (c *CircuitBreaker) State() gobreaker.State {
	return c.cb.State()
}

// IsOpen 是否处于开启状态（熔断中）
func (c *CircuitBreaker) IsOpen() bool {
	return c.cb.State() == gobreaker.StateOpen
}

// Manager 熔断器管理器
type Manager struct {
	breakers sync.Map
	config   *Config
}

// NewManager 创建熔断器管理器
func NewManager(defaultCfg *Config) *Manager {
	if defaultCfg == nil {
		defaultCfg = DefaultConfig("default")
	}
	return &Manager{
		config: defaultCfg,
	}
}

// Get 获取或创建熔断器
func (m *Manager) Get(name string) *CircuitBreaker {
	if cb, ok := m.breakers.Load(name); ok {
		return cb.(*CircuitBreaker)
	}

	cfg := *m.config
	cfg.Name = name
	cb := New(&cfg)
	m.breakers.Store(name, cb)
	return cb
}

// Execute 通过名称执行熔断保护
func (m *Manager) Execute(name string, fn func() (interface{}, error)) (interface{}, error) {
	return m.Get(name).Execute(fn)
}

// 全局默认管理器
var defaultManager = NewManager(nil)

// GetBreaker 获取全局熔断器
func GetBreaker(name string) *CircuitBreaker {
	return defaultManager.Get(name)
}

// Execute 使用全局熔断器执行
func Execute(name string, fn func() (interface{}, error)) (interface{}, error) {
	return defaultManager.Execute(name, fn)
}
