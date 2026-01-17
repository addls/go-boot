package middleware

import (
	"github.com/addls/go-boot/config"
	"github.com/go-kratos/kratos/v2/middleware"
)

// DefaultWithConfig 返回默认中间件集合（使用指定配置）
// 按照执行顺序：Recovery -> Metadata -> Tracing(可选) -> Logging -> Metrics(可选)
func DefaultWithConfig(cfg config.Middleware) []middleware.Middleware {
	middlewares := []middleware.Middleware{
		Recovery(), // 最外层：panic 恢复（必须）
		Metadata(), // 元数据传递（必须，用于服务间通信）
	}

	// Tracing 在 Logging 之前，确保日志中包含 trace 信息
	if cfg.EnableTracing {
		middlewares = append(middlewares, Tracing())
	}

	// Logging 必须启用
	middlewares = append(middlewares, Logging())

	// Metrics 在最内层，记录最终的处理结果
	if cfg.EnableMetrics {
		middlewares = append(middlewares, Metrics())
	}

	return middlewares
}
