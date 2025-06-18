package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger 初始化日志
func InitLogger(env string) {
	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	// 设置日志级别
	logLevel := zapcore.InfoLevel
	if env == "development" {
		logLevel = zapcore.DebugLevel
	}

	// 配置文件输出
	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 创建日志文件
	logFile, _ := os.OpenFile(
		filepath.Join(logDir, "blog-"+time.Now().Format("2006-01-02")+".log"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	// 配置日志核心
	core := zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(logFile),
		logLevel,
	)

	// 创建日志记录器
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// GinLogger Gin 框架的日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := time.Since(start)
		Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("cost", cost),
			zap.Int("body_size", c.Writer.Size()),
		)
	}
}

// Debug level logging
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info level logging
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn level logging
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error level logging
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal level logging
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// WithContext 添加上下文信息到日志
func WithContext(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}
