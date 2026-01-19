package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/addls/go-boot/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(service string, logConfig config.Log) *zap.Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// 确定日志输出位置
	var writeSyncer zapcore.WriteSyncer
	output := logConfig.Output
	if output == "" {
		output = "logs/app.log" // 默认输出到文件
	}

	switch output {
	case "stdout", "STDOUT":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "stderr", "STDERR":
		writeSyncer = zapcore.AddSync(os.Stderr)
	default:
		// 输出到文件，在文件名中添加日期
		logPath := addDateToLogPath(output)

		// 确保日志目录存在
		logDir := filepath.Dir(logPath)
		if logDir != "" && logDir != "." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				// 如果创建目录失败，回退到 stdout
				writeSyncer = zapcore.AddSync(os.Stdout)
			} else {
				// 打开或创建日志文件（追加模式）
				file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					// 如果打开文件失败，回退到 stdout
					writeSyncer = zapcore.AddSync(os.Stdout)
				} else {
					writeSyncer = zapcore.AddSync(file)
				}
			}
		} else {
			// 如果路径是当前目录，直接创建文件
			file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				writeSyncer = zapcore.AddSync(os.Stdout)
			} else {
				writeSyncer = zapcore.AddSync(file)
			}
		}
	}

	// 确定日志级别
	level := zap.InfoLevel
	switch logConfig.Level {
	case "debug", "DEBUG":
		level = zap.DebugLevel
	case "info", "INFO":
		level = zap.InfoLevel
	case "warn", "WARN", "warning", "WARNING":
		level = zap.WarnLevel
	case "error", "ERROR":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		writeSyncer,
		level,
	)

	return zap.New(core).With(zap.String("service", service))
}

// addDateToLogPath 在日志文件路径中添加日期
// 例如：logs/app.log -> logs/app-2024-01-01.log
func addDateToLogPath(path string) string {
	// 如果路径是 stdout 或 stderr，直接返回
	if path == "stdout" || path == "STDOUT" || path == "stderr" || path == "STDERR" {
		return path
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// 添加日期：YYYY-MM-DD
	date := time.Now().Format("2006-01-02")
	newName := fmt.Sprintf("%s-%s%s", name, date, ext)

	if dir == "." || dir == "" {
		return newName
	}

	return filepath.Join(dir, newName)
}
