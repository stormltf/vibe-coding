package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/config"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/database"
	"github.com/test-tt/pkg/logger"
	"github.com/test-tt/pkg/response"
	"go.uber.org/zap"
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

	mysqlOK := false
	redisOK := false

	// 检查 MySQL 连接
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil {
			// 使用带超时的 Ping
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if sqlDB.PingContext(pingCtx) == nil {
				status.MySQL = "connected"
				mysqlOK = true
			}
		}
	}

	// 检查 Redis 连接
	if cache.RDB != nil {
		// 使用带超时的 Ping
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if _, err := cache.RDB.Ping(pingCtx).Result(); err == nil {
			status.Redis = "connected"
			redisOK = true
		}
	}

	// 确定整体状态
	cfg := config.Cfg
	if cfg != nil && cfg.IsProd() {
		// 生产环境：MySQL 和 Redis 都是必需的
		if !mysqlOK || !redisOK {
			status.Status = "unhealthy"
		}
	} else {
		// 开发环境：只要有一个可用就是降级状态，都不可用才是不健康
		if !mysqlOK && !redisOK {
			// 如果 DB 和 Redis 都未配置（nil），则认为是健康的（无状态模式）
			if database.DB == nil && cache.RDB == nil {
				status.Status = "healthy"
			} else {
				status.Status = "unhealthy"
			}
		} else if !mysqlOK || !redisOK {
			status.Status = "degraded"
		}
	}

	logger.Info("health check completed",
		zap.String("status", status.Status),
		zap.String("mysql", status.MySQL),
		zap.String("redis", status.Redis))

	return status
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
