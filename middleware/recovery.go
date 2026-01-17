package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Recovery 返回一个恢复中间件，捕获 panic 并记录
func Recovery() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					var (
						kind      string
						operation string
					)
					if info, ok := transport.FromServerContext(ctx); ok {
						kind = info.Kind().String()
						operation = info.Operation()
					}
					logger := log.NewHelper(log.DefaultLogger)
					logger.Log(log.LevelError,
						"kind", kind,
						"operation", operation,
						"panic", r,
						"stack", string(debug.Stack()),
					)
					err = errors.InternalServer("INTERNAL_ERROR", fmt.Sprintf("panic recovered: %v", r))
				}
			}()
			return handler(ctx, req)
		}
	}
}
