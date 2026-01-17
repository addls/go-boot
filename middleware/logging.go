package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Logging 返回一个日志中间件，记录请求和响应信息
func Logging() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var (
				code      int
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			reply, err := handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = int(se.Code)
				reason = se.Reason
			}
			logger := log.NewHelper(log.DefaultLogger)
			if err != nil {
				logger.Log(log.LevelError,
					"kind", kind,
					"operation", operation,
					"code", code,
					"reason", reason,
					"latency", time.Since(startTime).Seconds(),
					"error", err,
				)
			} else {
				logger.Log(log.LevelInfo,
					"kind", kind,
					"operation", operation,
					"code", code,
					"reason", reason,
					"latency", time.Since(startTime).Seconds(),
				)
			}
			return reply, err
		}
	}
}
