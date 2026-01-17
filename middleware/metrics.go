package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
)

// Metrics 返回一个监控指标中间件
// 使用 Kratos 内置的 metrics 中间件，统一监控指标格式
func Metrics() middleware.Middleware {
	return metrics.Server()
}
