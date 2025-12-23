package logger

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogIDKey 用于从 context 中获取 logid
type logIDKey struct{}

// ContextWithLogID 将 logid 注入到 context
func ContextWithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, logIDKey{}, logID)
}

// GetLogID 从 context 获取 logid
func GetLogID(ctx context.Context) string {
	if id, ok := ctx.Value(logIDKey{}).(string); ok {
		return id
	}
	return ""
}

var Log *zap.Logger
var sugar *zap.SugaredLogger

type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	Filename   string // 日志文件路径，为空则不写入文件
	MaxSize    int    // 单个日志文件最大大小（MB）
	MaxBackups int    // 保留的旧日志文件数量
	MaxAge     int    // 保留天数
	Compress   bool   // 是否压缩
	Color      bool   // 控制台是否彩色输出
}

func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "console",
		Filename:   "",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
		Color:      true,
	}
}

func Init(cfg *Config) error {
	level := getLogLevel(cfg.Level)

	var cores []zapcore.Core

	// 控制台输出 - 彩色格式
	consoleCore := createConsoleCore(level, cfg.Color)
	cores = append(cores, consoleCore)

	// 文件输出 - JSON 格式 + 轮转
	if cfg.Filename != "" {
		fileCore := createFileCore(level, cfg)
		cores = append(cores, fileCore)
	}

	// 合并多个 core
	core := zapcore.NewTee(cores...)
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	// Sugar logger 使用相同的 caller skip
	sugar = Log.Sugar()

	return nil
}

// 创建彩色控制台 Core
func createConsoleCore(level zapcore.Level, color bool) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if color {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

// 创建 JSON 文件 Core（带轮转）
func createFileCore(level zapcore.Level, cfg *Config) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if logDir != "" && logDir != "." {
		_ = os.MkdirAll(logDir, 0755)
	}

	// 日志轮转
	writer := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level)
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// ============ 字段构造函数（无需导入 zap）============

type Field = zap.Field

var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Uint64   = zap.Uint64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Err      = zap.Error
	Any      = zap.Any
	Duration = zap.Duration
	Time     = zap.Time
)

// ============ 日志方法 ============

func Debug(msg string, fields ...Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

// ============ 简化版日志（key-value 格式）============

// Debugf 使用 key-value 对打印日志：Debugf("msg", "key1", val1, "key2", val2)
func Debugf(msg string, keysAndValues ...interface{}) {
	if sugar != nil {
		sugar.Debugw(msg, keysAndValues...)
	}
}

// Infof 使用 key-value 对打印日志：Infof("msg", "key1", val1, "key2", val2)
func Infof(msg string, keysAndValues ...interface{}) {
	if sugar != nil {
		sugar.Infow(msg, keysAndValues...)
	}
}

// Warnf 使用 key-value 对打印日志：Warnf("msg", "key1", val1, "key2", val2)
func Warnf(msg string, keysAndValues ...interface{}) {
	if sugar != nil {
		sugar.Warnw(msg, keysAndValues...)
	}
}

// Errorf 使用 key-value 对打印日志：Errorf("msg", "key1", val1, "key2", val2)
func Errorf(msg string, keysAndValues ...interface{}) {
	if sugar != nil {
		sugar.Errorw(msg, keysAndValues...)
	}
}

// With 返回带有预设字段的 logger
func With(fields ...Field) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	return Log.With(fields...)
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
	if sugar != nil {
		_ = sugar.Sync()
	}
}

// ============ 带 Context 的日志方法（自动携带 logid）============

// Ctx 从 context 创建带 logid 的 logger
func Ctx(ctx context.Context) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	logID := getLogIDFromCtx(ctx)
	if logID != "" {
		return Log.With(zap.String("logid", logID))
	}
	return Log
}

// CtxSugar 从 context 创建带 logid 的 sugar logger
func CtxSugar(ctx context.Context) *zap.SugaredLogger {
	if sugar == nil {
		return zap.NewNop().Sugar()
	}
	logID := getLogIDFromCtx(ctx)
	if logID != "" {
		return sugar.With("logid", logID)
	}
	return sugar
}

// DebugCtx 带 context 的 Debug 日志
func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	Ctx(ctx).Debug(msg, fields...)
}

// InfoCtx 带 context 的 Info 日志
func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	Ctx(ctx).Info(msg, fields...)
}

// WarnCtx 带 context 的 Warn 日志
func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	Ctx(ctx).Warn(msg, fields...)
}

// ErrorCtx 带 context 的 Error 日志
func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	Ctx(ctx).Error(msg, fields...)
}

// DebugCtxf 带 context 的 key-value 格式 Debug 日志
func DebugCtxf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	CtxSugar(ctx).Debugw(msg, keysAndValues...)
}

// InfoCtxf 带 context 的 key-value 格式 Info 日志
func InfoCtxf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	CtxSugar(ctx).Infow(msg, keysAndValues...)
}

// WarnCtxf 带 context 的 key-value 格式 Warn 日志
func WarnCtxf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	CtxSugar(ctx).Warnw(msg, keysAndValues...)
}

// ErrorCtxf 带 context 的 key-value 格式 Error 日志
func ErrorCtxf(ctx context.Context, msg string, keysAndValues ...interface{}) {
	CtxSugar(ctx).Errorw(msg, keysAndValues...)
}

// getLogIDFromCtx 尝试从多种方式获取 logid
func getLogIDFromCtx(ctx context.Context) string {
	// 优先从 logger 包自己的 key 获取
	if id, ok := ctx.Value(logIDKey{}).(string); ok {
		return id
	}
	return ""
}
