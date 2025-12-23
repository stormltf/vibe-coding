package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/test-tt/config"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/database"
	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/response"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	MySQL     string            `json:"mysql"`
	Redis     string            `json:"redis"`
	Timestamp int64             `json:"timestamp"`
	Details   map[string]string `json:"details,omitempty"`
}

// Ping 快速健康检查（用于负载均衡器）
func (h *PingHandler) Ping(ctx context.Context, c *app.RequestContext) {
	status := h.checkHealth(ctx)

	// 根据健康状态返回对应的 HTTP 状态码
	if status.Status == "healthy" {
		response.SuccessWithMessage(c, "pong", status)
	} else if status.Status == "degraded" {
		// 降级状态：服务可用但部分依赖不可用
		c.JSON(http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "pong (degraded)",
			"data":    status,
		})
	} else {
		// 不健康：关键依赖不可用
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"code":    5003,
			"message": "service unhealthy",
			"data":    status,
		})
	}
}

// Health 完整健康检查（包含详细信息）
func (h *PingHandler) Health(ctx context.Context, c *app.RequestContext) {
	status := h.checkHealthDetailed(ctx)

	if status.Status == "healthy" {
		response.SuccessWithMessage(c, "healthy", status)
	} else if status.Status == "degraded" {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": "degraded",
			"data":    status,
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"code":    5003,
			"message": "unhealthy",
			"data":    status,
		})
	}
}

// checkHealth 检查健康状态
func (h *PingHandler) checkHealth(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Status:    "healthy",
		MySQL:     "disconnected",
		Redis:     "disconnected",
		Timestamp: time.Now().Unix(),
	}

	// 检查各组件
	mysqlOK := h.checkMySQL(ctx, &status)
	redisOK := h.checkRedis(ctx, &status)

	// 确定整体状态
	status.Status = h.determineOverallStatus(mysqlOK, redisOK)

	logger.Info("health check completed",
		zap.String("status", status.Status),
		zap.String("mysql", status.MySQL),
		zap.String("redis", status.Redis))

	return status
}

// checkMySQL 检查 MySQL 连接状态
func (h *PingHandler) checkMySQL(ctx context.Context, status *HealthStatus) bool {
	if database.DB == nil {
		return false
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		return false
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if sqlDB.PingContext(pingCtx) == nil {
		status.MySQL = "connected"
		return true
	}
	return false
}

// checkRedis 检查 Redis 连接状态
func (h *PingHandler) checkRedis(ctx context.Context, status *HealthStatus) bool {
	if cache.RDB == nil {
		return false
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := cache.RDB.Ping(pingCtx).Result(); err == nil {
		status.Redis = "connected"
		return true
	}
	return false
}

// determineOverallStatus 根据组件状态确定整体健康状态
func (h *PingHandler) determineOverallStatus(mysqlOK, redisOK bool) string {
	cfg := config.Cfg

	// 生产环境：所有依赖都是必需的
	if cfg != nil && cfg.IsProd() {
		if mysqlOK && redisOK {
			return "healthy"
		}
		return "unhealthy"
	}

	// 开发环境：更宽松的判断
	bothConfigured := database.DB != nil || cache.RDB != nil
	if !bothConfigured {
		return "healthy" // 无状态模式
	}

	if mysqlOK && redisOK {
		return "healthy"
	}
	if mysqlOK || redisOK {
		return "degraded"
	}
	return "unhealthy"
}

// checkHealthDetailed 详细健康检查
func (h *PingHandler) checkHealthDetailed(ctx context.Context) HealthStatus {
	status := h.checkHealth(ctx)
	status.Details = make(map[string]string)

	// 添加 MySQL 连接池状态
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil {
			stats := sqlDB.Stats()
			status.Details["mysql_open_connections"] = fmt.Sprintf("%d", stats.OpenConnections)
			status.Details["mysql_in_use"] = fmt.Sprintf("%d", stats.InUse)
			status.Details["mysql_idle"] = fmt.Sprintf("%d", stats.Idle)
		}
	}

	// 添加 Redis 连接池状态
	if cache.RDB != nil {
		stats := cache.RDB.PoolStats()
		status.Details["redis_total_conns"] = fmt.Sprintf("%d", stats.TotalConns)
		status.Details["redis_idle_conns"] = fmt.Sprintf("%d", stats.IdleConns)
	}

	return status
}
