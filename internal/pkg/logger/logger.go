package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger 初始化日志
func InitLogger(env string) {
	// 设置日志级别
	logLevel := zapcore.InfoLevel
	if env == "development" {
		logLevel = zapcore.DebugLevel
	}

	// 配置控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 配置文件输出
	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// 创建日志文件
	logFile, _ := os.OpenFile(
		"logs/blog-"+time.Now().Format("2006-01-02")+".log",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)

	// 配置日志核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), logLevel),
	)

	// 创建日志记录器
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
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
