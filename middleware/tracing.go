package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
)

// Tracing 返回一个链路追踪中间件
// 使用 Kratos 内置的 tracing 中间件，统一追踪格式
func Tracing() middleware.Middleware {
	return tracing.Server()
}
