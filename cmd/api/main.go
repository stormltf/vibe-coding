package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/test-tt/config"
	"github.com/test-tt/internal/middleware"
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

	// 验证配置
	if err := config.Validate(cfg); err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
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

	logger.Infof("starting server", "config", configPath, "env", cfg.Env)

	// 资源清理函数列表（按逆序执行）
	var cleanups []func()
	defer func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}()

	// 初始化 MySQL
	mysqlRequired := cfg.IsProd() // 生产环境必须连接 MySQL
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
		if mysqlRequired {
			panic(fmt.Sprintf("init mysql failed (required in production): %v", err))
		}
		logger.Warnf("init mysql failed (service will run without database)", "error", err)
	} else {
		logger.Info("MySQL connected")
		cleanups = append(cleanups, func() {
			logger.Info("closing MySQL connection...")
			database.Close()
		})
	}

	// 初始化 Redis
	redisRequired := cfg.IsProd() // 生产环境必须连接 Redis
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
		if redisRequired {
			panic(fmt.Sprintf("init redis failed (required in production): %v", err))
		}
		logger.Warnf("init redis failed (service will run without cache)", "error", err)
	} else {
		logger.Info("Redis connected")
		cleanups = append(cleanups, func() {
			logger.Info("closing Redis connection...")
			cache.Close()
		})
	}

	// 初始化本地缓存（L1 缓存）
	if err := cache.InitLocalCache(nil); err != nil {
		logger.Warnf("init local cache failed", "error", err)
	} else {
		logger.Info("LocalCache initialized (64MB)")
		cleanups = append(cleanups, func() {
			logger.Info("closing local cache...")
			if lc := cache.GetLocalCache(); lc != nil {
				lc.Close()
			}
		})
	}

	// 启动连接池指标收集器
	stopMetricsCollector := middleware.StartPoolMetricsCollector(15 * time.Second)
	cleanups = append(cleanups, func() {
		logger.Info("stopping pool metrics collector...")
		stopMetricsCollector()
	})

	// 初始化 HTTP 服务器
	h := server.Default(
		server.WithHostPorts(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)),
		server.WithExitWaitTime(5*time.Second),
		server.WithMaxRequestBodySize(4*1024*1024), // 4MB 请求体限制
		server.WithReadTimeout(30*time.Second),     // 读超时
		server.WithWriteTimeout(30*time.Second),    // 写超时
		server.WithIdleTimeout(120*time.Second),    // 空闲连接超时
	)

	// 注册路由
	router.Register(h)

	// 优雅关闭
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 使用缓冲区 2 以捕获多个信号
		quit := make(chan os.Signal, 2)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		logger.Infof("received shutdown signal", "signal", sig.String())
		logger.Info("shutting down server (waiting for active requests)...")

		// 增加超时时间到 30 秒，给活跃请求更多时间完成
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 1. 先关闭服务器（停止接收新请求，等待活跃请求完成）
		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("server shutdown error", "error", err)
		} else {
			logger.Info("server shutdown completed gracefully")
		}

		// 2. 服务器关闭后停止限流器（此时不再有新请求）
		logger.Info("stopping rate limiters...")
		middleware.StopAllRateLimiters()
	}()

	logger.Infof("server started", "host", cfg.Server.Host, "port", cfg.Server.Port)
	h.Spin()

	// 等待优雅关闭完成
	wg.Wait()
	logger.Info("all resources cleaned up, server stopped")
}
