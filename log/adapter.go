package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger 将 zap.Logger 适配为 kratos log.Logger
type ZapLogger struct {
	logger *zap.Logger
}

// NewKratosLogger 创建适配 kratos 的 logger
func NewKratosLogger(service string) log.Logger {
	zapLogger := New(service)
	return &ZapLogger{logger: zapLogger}
}

// Log 实现 kratos log.Logger 接口
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.logger.Info("Invalid keyvals")
		return nil
	}

	var zapLevel zapcore.Level
	switch level {
	case log.LevelDebug:
		zapLevel = zapcore.DebugLevel
	case log.LevelInfo:
		zapLevel = zapcore.InfoLevel
	case log.LevelWarn:
		zapLevel = zapcore.WarnLevel
	case log.LevelError:
		zapLevel = zapcore.ErrorLevel
	case log.LevelFatal:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	if !l.logger.Core().Enabled(zapLevel) {
		return nil
	}

	entry := l.logger.With(fields...)
	switch zapLevel {
	case zapcore.DebugLevel:
		entry.Debug("")
	case zapcore.InfoLevel:
		entry.Info("")
	case zapcore.WarnLevel:
		entry.Warn("")
	case zapcore.ErrorLevel:
		entry.Error("")
	case zapcore.FatalLevel:
		entry.Fatal("")
	}

	return nil
}
