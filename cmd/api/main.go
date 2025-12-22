package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/test-tt/config"
	"github.com/test-tt/internal/router"
	"github.com/test-tt/pkg/cache"
	"github.com/test-tt/pkg/database"
	"github.com/test-tt/pkg/logger"

	_ "github.com/test-tt/docs" // swagger docs
)

// @title           Test-TT API
// @version         1.0
// @description     企业级 Go Web API 模板项目
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/test-tt
// @contact.email  support@test-tt.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8888
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT Token, 格式: Bearer {token}

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "config file path")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		Color:      cfg.Log.Color,
	}); err != nil {
		panic(fmt.Sprintf("init logger failed: %v", err))
	}
	defer logger.Sync()

	logger.Infof("starting server", "config", configPath)

	// 初始化 MySQL
	if err := database.Init(&database.Config{
		Host:            cfg.MySQL.Host,
		Port:            cfg.MySQL.Port,
		Username:        cfg.MySQL.Username,
		Password:        cfg.MySQL.Password,
		Database:        cfg.MySQL.Database,
		Charset:         cfg.MySQL.Charset,
		MaxIdleConns:    cfg.MySQL.MaxIdleConns,
		MaxOpenConns:    cfg.MySQL.MaxOpenConns,
		ConnMaxLifetime: cfg.MySQL.ConnMaxLifetime,
		LogLevel:        cfg.MySQL.LogLevel,
	}); err != nil {
		logger.Warnf("init mysql failed", "error", err)
	} else {
		logger.Info("MySQL connected")
		defer database.Close()
	}

	// 初始化 Redis
	if err := cache.Init(&cache.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	}); err != nil {
		logger.Warnf("init redis failed", "error", err)
	} else {
		logger.Info("Redis connected")
		defer cache.Close()
	}

	// 初始化本地缓存（L1 缓存）
	if err := cache.InitLocalCache(nil); err != nil {
		logger.Warnf("init local cache failed", "error", err)
	} else {
		logger.Info("LocalCache initialized (64MB)")
	}

	// 初始化 HTTP 服务器
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithExitWaitTime(5*time.Second),
		server.WithMaxRequestBodySize(4*1024*1024),          // 4MB 请求体限制
		server.WithReadTimeout(30*time.Second),               // 读超时
		server.WithWriteTimeout(30*time.Second),              // 写超时
		server.WithIdleTimeout(120*time.Second),              // 空闲连接超时
	)

	// 注册路由
	router.Register(h)

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("server shutdown error", "error", err)
		}

		logger.Info("server stopped")
	}()

	logger.Infof("server started", "host", cfg.Server.Host, "port", cfg.Server.Port)
	h.Spin()
}
