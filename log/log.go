package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

var logger *zap.Logger
var once sync.Once

func init() {
	l, err := zap.NewProduction(
		zap.AddStacktrace(zap.FatalLevel),
		zap.WithCaller(false),
	)
	if err != nil {
		panic(err)
	}
	logger = l
}

func SetLogLevel(level string) {
	switch level {
	case "debug":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	case "info":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	case "warn":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case "error":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	case "fatal":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.FatalLevel))
	default:
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	}
}

func GetLogger() *zap.Logger {
	return logger
}

var CloseFunc []func() error

func AddCloseFunc(f func() error) {
	CloseFunc = append(CloseFunc, f)
}

func Close() {
	once.Do(func() {
		for _, f := range CloseFunc {
			err := f()
			if err != nil {
				logger.Error("close error", zap.Error(err))
			}
		}
		logger.Sync()
	})
}

func SubLogger(name string) *zap.Logger {
	return logger.Named(name)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func ErrorE(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
	Close()
	os.Exit(1)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
