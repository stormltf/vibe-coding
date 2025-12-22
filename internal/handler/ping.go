package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
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

type HealthStatus struct {
	Status string `json:"status"`
	MySQL  string `json:"mysql"`
	Redis  string `json:"redis"`
}

func (h *PingHandler) Ping(ctx context.Context, c *app.RequestContext) {
	logger.Info("health check requested")

	status := HealthStatus{
		Status: "ok",
		MySQL:  "disconnected",
		Redis:  "disconnected",
	}

	// 检查 MySQL 连接
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil && sqlDB.Ping() == nil {
			status.MySQL = "connected"
		}
	}

	// 检查 Redis 连接
	if cache.RDB != nil {
		if _, err := cache.RDB.Ping(ctx).Result(); err == nil {
			status.Redis = "connected"
		}
	}

	logger.Info("health check completed",
		zap.String("mysql", status.MySQL),
		zap.String("redis", status.Redis))

	response.SuccessWithMessage(c, "pong", status)
}
