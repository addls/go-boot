package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
)

// Metadata 返回一个元数据中间件
// 用于在服务间传递请求元数据（如 trace-id、request-id 等）
// 这是微服务间通信的基础能力，应该默认启用
func Metadata() middleware.Middleware {
	return metadata.Server()
}
