package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Charset         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogLevel        logger.LogLevel
}

// DefaultConfig 返回优化后的默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxIdleConns:    50,               // 空闲连接数（建议: CPU核心数 * 2）
		MaxOpenConns:    100,              // 最大连接数（建议: 不超过 MySQL max_connections / 应用实例数）
		ConnMaxLifetime: 30 * time.Minute, // 连接最大生存时间（建议: 小于 MySQL wait_timeout）
		ConnMaxIdleTime: 10 * time.Minute, // 空闲连接最大生存时间
		Charset:         "utf8mb4",
		LogLevel:        logger.Warn,
	}
}

func Init(cfg *Config) error {
	// 合并默认配置
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 50
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 100
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 30 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 10 * time.Minute
	}

	// DSN 优化参数:
	// - interpolateParams=true: 客户端插值，减少一次网络往返
	// - timeout=5s: 连接超时
	// - readTimeout=30s: 读超时
	// - writeTimeout=30s: 写超时
	// - maxAllowedPacket=0: 使用服务器默认值
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&interpolateParams=true&timeout=5s&readTimeout=30s&writeTimeout=30s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,   // string 类型默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度（MySQL 5.6 之前不支持）
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式
		DontSupportRenameColumn:   true,  // 用 change 重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger:                                   logger.Default.LogMode(cfg.LogLevel),
		SkipDefaultTransaction:                   true, // 跳过默认事务，提升性能
		PrepareStmt:                              true, // 预编译语句缓存
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
		QueryFields:                              true, // 使用字段名查询，避免 SELECT *
	})
	if err != nil {
		return fmt.Errorf("failed to connect mysql: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 连接池优化
	// MaxIdleConns: 保持足够的空闲连接避免频繁创建
	// MaxOpenConns: 限制最大连接数避免耗尽 MySQL 连接
	// ConnMaxLifetime: 定期回收连接避免使用过期连接
	// ConnMaxIdleTime: 回收长时间空闲的连接
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		// 连接验证失败时，关闭连接避免资源泄漏
		_ = sqlDB.Close()
		DB = nil
		return fmt.Errorf("failed to ping mysql: %w", err)
	}

	return nil
}

// Stats 获取连接池统计信息
func Stats() map[string]interface{} {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return nil
	}
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
