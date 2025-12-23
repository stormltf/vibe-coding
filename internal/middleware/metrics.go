package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/database"
)

var (
	// httpRequestsTotal HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration HTTP 请求延迟
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// httpRequestsInFlight 当前处理中的请求数
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// MySQL 连接池指标
	mysqlOpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mysql_pool_open_connections",
		Help: "Number of open connections to MySQL",
	})
	mysqlInUseConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mysql_pool_in_use_connections",
		Help: "Number of connections currently in use",
	})
	mysqlIdleConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mysql_pool_idle_connections",
		Help: "Number of idle connections",
	})
	mysqlWaitCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "mysql_pool_wait_count_total",
		Help: "Total number of connections waited for",
	})
	mysqlWaitDuration = promauto.NewCounter(prometheus.CounterOpts{
		Name: "mysql_pool_wait_duration_seconds_total",
		Help: "Total time blocked waiting for a new connection",
	})

	// Redis 连接池指标
	redisPoolHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_pool_hits_total",
		Help: "Number of times a free connection was found in the pool",
	})
	redisPoolMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_pool_misses_total",
		Help: "Number of times a free connection was NOT found in the pool",
	})
	redisPoolTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_pool_timeouts_total",
		Help: "Number of times a wait timeout occurred",
	})
	redisTotalConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_total_connections",
		Help: "Number of total connections in the pool",
	})
	redisIdleConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "redis_pool_idle_connections",
		Help: "Number of idle connections in the pool",
	})
	redisStaleConnections = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_pool_stale_connections_total",
		Help: "Number of stale connections removed from the pool",
	})
)

// 用于跟踪增量指标的前值
var (
	lastMySQLWaitCount    int64
	lastMySQLWaitDuration float64
	lastRedisHits         uint32
	lastRedisMisses       uint32
	lastRedisTimeouts     uint32
	lastRedisStale        uint32
)

// Metrics Prometheus 指标中间件
func Metrics() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		// 增加处理中请求计数
		httpRequestsInFlight.Inc()

		c.Next(ctx)

		// 减少处理中请求计数
		httpRequestsInFlight.Dec()

		// 使用路由模板避免高基数问题
		// FullPath() 返回 "/users/:id" 而不是 "/users/123"
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}

		// 获取请求方法和状态码
		method := string(c.Method())
		status := strconv.Itoa(c.Response.StatusCode())

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	}
}

// UpdatePoolMetrics 更新连接池指标
// 应该定期调用（如每 15 秒）或在 /metrics 端点时调用
func UpdatePoolMetrics() {
	// 更新 MySQL 指标
	if stats := database.Stats(); stats != nil {
		mysqlOpenConnections.Set(float64(stats["open_connections"].(int)))
		mysqlInUseConnections.Set(float64(stats["in_use"].(int)))
		mysqlIdleConnections.Set(float64(stats["idle"].(int)))

		// 增量更新 counter
		waitCount := stats["wait_count"].(int64)
		if waitCount > lastMySQLWaitCount {
			mysqlWaitCount.Add(float64(waitCount - lastMySQLWaitCount))
			lastMySQLWaitCount = waitCount
		}

		// 解析 wait_duration
		if durationStr, ok := stats["wait_duration"].(string); ok {
			if d, err := time.ParseDuration(durationStr); err == nil {
				seconds := d.Seconds()
				if seconds > lastMySQLWaitDuration {
					mysqlWaitDuration.Add(seconds - lastMySQLWaitDuration)
					lastMySQLWaitDuration = seconds
				}
			}
		}
	}

	// 更新 Redis 指标
	if stats := cache.Stats(); stats != nil {
		redisTotalConnections.Set(float64(stats.TotalConns))
		redisIdleConnections.Set(float64(stats.IdleConns))

		// 增量更新 counter
		if stats.Hits > lastRedisHits {
			redisPoolHits.Add(float64(stats.Hits - lastRedisHits))
			lastRedisHits = stats.Hits
		}
		if stats.Misses > lastRedisMisses {
			redisPoolMisses.Add(float64(stats.Misses - lastRedisMisses))
			lastRedisMisses = stats.Misses
		}
		if stats.Timeouts > lastRedisTimeouts {
			redisPoolTimeouts.Add(float64(stats.Timeouts - lastRedisTimeouts))
			lastRedisTimeouts = stats.Timeouts
		}
		if stats.StaleConns > lastRedisStale {
			redisStaleConnections.Add(float64(stats.StaleConns - lastRedisStale))
			lastRedisStale = stats.StaleConns
		}
	}
}

// StartPoolMetricsCollector 启动连接池指标收集器
// 返回停止函数
func StartPoolMetricsCollector(interval time.Duration) func() {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				UpdatePoolMetrics()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(stopChan)
	}
}
