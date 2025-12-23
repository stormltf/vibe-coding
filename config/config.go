package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

var Cfg *Config

type Config struct {
	Env       string           `mapstructure:"env"`
	Server    *ServerConfig    `mapstructure:"server"`
	MySQL     *MySQLConfig     `mapstructure:"mysql"`
	Redis     *RedisConfig     `mapstructure:"redis"`
	Log       *LogConfig       `mapstructure:"log"`
	JWT       *JWTConfig       `mapstructure:"jwt"`
	RateLimit *RateLimitConfig `mapstructure:"ratelimit"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MySQLConfig struct {
	Host            string          `mapstructure:"host"`
	Port            int             `mapstructure:"port"`
	Username        string          `mapstructure:"username"`
	Password        string          `mapstructure:"password"`
	Database        string          `mapstructure:"database"`
	Charset         string          `mapstructure:"charset"`
	MaxIdleConns    int             `mapstructure:"max_idle_conns"`
	MaxOpenConns    int             `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration   `mapstructure:"conn_max_lifetime"`
	LogLevel        logger.LogLevel `mapstructure:"log_level"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Color      bool   `mapstructure:"color"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Issuer     string        `mapstructure:"issuer"`
	ExpireTime time.Duration `mapstructure:"expire_time"`
}

type RateLimitConfig struct {
	Rate  float64 `mapstructure:"rate"`
	Burst int     `mapstructure:"burst"`
}

// Load 从配置文件和环境变量加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// 环境变量支持（前缀 APP_，如 APP_SERVER_PORT）
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file error: %w", err)
		}
		// 配置文件不存在，使用默认值
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config error: %w", err)
	}

	Cfg = &cfg
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8888)

	// MySQL
	v.SetDefault("mysql.host", "127.0.0.1")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.username", "root")
	v.SetDefault("mysql.password", "")
	v.SetDefault("mysql.database", "test")
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.max_idle_conns", 50)
	v.SetDefault("mysql.max_open_conns", 200)
	v.SetDefault("mysql.conn_max_lifetime", "30m")
	v.SetDefault("mysql.log_level", 4)

	// Redis
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 20)
	v.SetDefault("redis.dial_timeout", "5s")  // 连接超时
	v.SetDefault("redis.read_timeout", "3s")  // 读超时（适当放宽，避免网络抖动）
	v.SetDefault("redis.write_timeout", "3s") // 写超时

	// Log
	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 3)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.compress", true)
	v.SetDefault("log.color", true)

	// JWT
	v.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	v.SetDefault("jwt.issuer", "test-tt")
	v.SetDefault("jwt.expire_time", "24h")

	// RateLimit
	v.SetDefault("ratelimit.rate", 100)
	v.SetDefault("ratelimit.burst", 200)

	// Env
	v.SetDefault("env", "dev")
}

// IsDev 是否开发环境
func (c *Config) IsDev() bool {
	return c.Env == "dev" || c.Env == ""
}

// IsProd 是否生产环境
func (c *Config) IsProd() bool {
	return c.Env == "prod"
}

// 默认不安全的 JWT Secret
const defaultInsecureSecret = "your-secret-key-change-in-production"

// Validate 验证配置的有效性
func Validate(cfg *Config) error {
	var errs []string

	errs = append(errs, validateJWT(cfg)...)
	errs = append(errs, validateMySQL(cfg.MySQL)...)
	errs = append(errs, validateRedis(cfg.Redis)...)
	errs = append(errs, validateServer(cfg.Server)...)
	errs = append(errs, validateRateLimit(cfg.RateLimit)...)

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", errs)
	}
	return nil
}

// validateJWT 验证 JWT 配置
func validateJWT(cfg *Config) []string {
	if cfg.JWT == nil {
		return nil
	}
	var errs []string
	if cfg.JWT.Secret == "" {
		errs = append(errs, "jwt.secret is required")
	} else if cfg.JWT.Secret == defaultInsecureSecret && cfg.IsProd() {
		errs = append(errs, "jwt.secret must be changed in production (use APP_JWT_SECRET env var)")
	} else if len(cfg.JWT.Secret) < 32 && cfg.IsProd() {
		errs = append(errs, "jwt.secret must be at least 32 characters in production")
	}
	return errs
}

// validateMySQL 验证 MySQL 配置
func validateMySQL(cfg *MySQLConfig) []string {
	if cfg == nil {
		return nil
	}
	var errs []string
	if cfg.MaxOpenConns > 500 {
		errs = append(errs, "mysql.max_open_conns should not exceed 500")
	}
	if cfg.MaxOpenConns > 0 && cfg.MaxIdleConns > cfg.MaxOpenConns {
		errs = append(errs, "mysql.max_idle_conns should not exceed max_open_conns")
	}
	if cfg.ConnMaxLifetime < 0 {
		errs = append(errs, "mysql.conn_max_lifetime must be positive")
	}
	if cfg.MaxOpenConns <= 0 {
		errs = append(errs, "mysql.max_open_conns must be positive")
	}
	if cfg.MaxIdleConns < 0 {
		errs = append(errs, "mysql.max_idle_conns must be non-negative")
	}
	return errs
}

// validateRedis 验证 Redis 配置
func validateRedis(cfg *RedisConfig) []string {
	if cfg == nil {
		return nil
	}
	var errs []string
	if cfg.PoolSize > 1000 {
		errs = append(errs, "redis.pool_size should not exceed 1000")
	}
	if cfg.DialTimeout <= 0 {
		errs = append(errs, "redis.dial_timeout must be positive")
	}
	if cfg.ReadTimeout <= 0 {
		errs = append(errs, "redis.read_timeout must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		errs = append(errs, "redis.write_timeout must be positive")
	}
	return errs
}

// validateServer 验证 Server 配置
func validateServer(cfg *ServerConfig) []string {
	if cfg == nil {
		return nil
	}
	var errs []string
	if cfg.Port <= 0 || cfg.Port > 65535 {
		errs = append(errs, "server.port must be between 1 and 65535")
	}
	return errs
}

// validateRateLimit 验证 RateLimit 配置
func validateRateLimit(cfg *RateLimitConfig) []string {
	if cfg == nil {
		return nil
	}
	var errs []string
	if cfg.Rate <= 0 {
		errs = append(errs, "ratelimit.rate must be positive")
	}
	if cfg.Burst <= 0 {
		errs = append(errs, "ratelimit.burst must be positive")
	}
	return errs
}

// MustValidate 验证配置，失败则 panic
func MustValidate(cfg *Config) {
	if err := Validate(cfg); err != nil {
		panic(err)
	}
}
