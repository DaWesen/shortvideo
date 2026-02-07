package logger

import (
	"os"
	"shortvideo/pkg/config"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerInstance *zap.Logger
	loggerOnce     sync.Once
)

// 日志接口
type Logger interface {
	//不同级别的日志记录
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)

	//结构化日志
	With(fields ...zapcore.Field) *zap.Logger

	//关闭日志
	Sync() error
}

// 获取日志实例
func GetLogger() *zap.Logger {
	loggerOnce.Do(func() {
		loggerInstance = initZapLogger()
	})
	return loggerInstance
}

// 初始化Zap日志记录器
func initZapLogger() *zap.Logger {
	logConfig := config.Get().Log

	//配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	//创建编码器
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	//配置日志级别
	var level zapcore.Level
	switch logConfig.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	//确保日志目录存在
	if logConfig.FilePath != "" {
		logDir := getLogDir(logConfig.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logConfig.FilePath = ""
		}
	}

	//创建Core
	var core zapcore.Core
	if logConfig.FilePath != "" {
		fileWriter, err := os.OpenFile(logConfig.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			core = zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stdout),
				level,
			)
		} else {
			core = zapcore.NewTee(
				zapcore.NewCore(
					encoder,
					zapcore.AddSync(fileWriter),
					level,
				),
				zapcore.NewCore(
					encoder,
					zapcore.AddSync(os.Stdout),
					level,
				),
			)
		}
	} else {
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
	}

	//创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger
}

// 获取日志目录
func getLogDir(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '\\' || filePath[i] == '/' {
			return filePath[:i]
		}
	}
	return ""
}

// 记录调试级别日志
func Debug(msg string, fields ...zapcore.Field) {
	GetLogger().Debug(msg, fields...)
}

// 记录信息级别日志
func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

// 记录警告级别日志
func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

// 记录错误级别日志
func Error(msg string, fields ...zapcore.Field) {
	GetLogger().Error(msg, fields...)
}

// 记录致命级别日志
func Fatal(msg string, fields ...zapcore.Field) {
	GetLogger().Fatal(msg, fields...)
}

// 创建带有字段的
func With(fields ...zapcore.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// 刷新缓冲区
func Sync() error {
	return GetLogger().Sync()
}

// 创建错误字段
func ErrorField(err error) zapcore.Field {
	return zap.Error(err)
}

// 创建字符串字段
func StringField(key, value string) zapcore.Field {
	return zap.String(key, value)
}

// 创建整数字段
func IntField(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

// 创建int64字段
func Int64Field(key string, value int64) zapcore.Field {
	return zap.Int64(key, value)
}

// 创建布尔字段
func BoolField(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

// 创建持续时间字段
func DurationField(key string, value time.Duration) zapcore.Field {
	return zap.Duration(key, value)
}

// 创建时间字段
func TimeField(key string, value time.Time) zapcore.Field {
	return zap.Time(key, value)
}

// 创建任意类型字段
func AnyField(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}
