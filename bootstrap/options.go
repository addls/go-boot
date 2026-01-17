package bootstrap

import (
	"github.com/addls/go-boot/config"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Option 启动选项
type Option func(*options)

type options struct {
	configFile       string
	config           *config.Config
	grpcOpts         []grpc.ServerOption
	httpOpts         []http.ServerOption
	customMiddleware []middleware.Middleware
	appOpts          []kratos.Option
}

// WithConfigFile 指定配置文件路径
func WithConfigFile(path string) Option {
	return func(o *options) {
		o.configFile = path
	}
}

// WithConfig 直接传入配置（优先级高于配置文件）
func WithConfig(cfg *config.Config) Option {
	return func(o *options) {
		o.config = cfg
	}
}

// WithGRPCOptions 配置额外的 gRPC 服务器选项
func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(o *options) {
		o.grpcOpts = append(o.grpcOpts, opts...)
	}
}

// WithHTTPOptions 配置额外的 HTTP 服务器选项
func WithHTTPOptions(opts ...http.ServerOption) Option {
	return func(o *options) {
		o.httpOpts = append(o.httpOpts, opts...)
	}
}

// WithHTTPRouter 注册 HTTP 路由（推荐方式）
// 这是业务代码注册路由的便捷方法，内部通过 WithHTTPOptions 实现
func WithHTTPRouter(fn func(*http.Server)) Option {
	return WithHTTPOptions(fn)
}

// WithMiddleware 添加自定义中间件
// 自定义中间件会在默认中间件之后执行
func WithMiddleware(middlewares ...middleware.Middleware) Option {
	return func(o *options) {
		o.customMiddleware = append(o.customMiddleware, middlewares...)
	}
}

// WithAppOptions 添加额外的 Kratos App 选项
// 用于配置生命周期钩子等业务特定选项
// 注意：服务注册通过配置文件自动处理，无需手动配置
func WithAppOptions(opts ...kratos.Option) Option {
	return func(o *options) {
		o.appOpts = append(o.appOpts, opts...)
	}
}
