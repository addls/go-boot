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
	httpRegisters    []func(*http.Server) // HTTP 路由注册函数（在服务器创建后调用）
	grpcRegisters    []func(*grpc.Server) // gRPC 服务注册函数（在服务器创建后调用）
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

// WithHTTPRegister 注册 HTTP 路由（在服务器创建后调用）
// 推荐使用此方法注册路由，比 WithHTTPOptions 更明确
func WithHTTPRegister(fn func(*http.Server)) Option {
	return func(o *options) {
		o.httpRegisters = append(o.httpRegisters, fn)
	}
}

// WithGRPCRegister 注册 gRPC 服务（在服务器创建后调用）
// 推荐使用此方法注册服务，比 WithGRPCOptions 更明确
func WithGRPCRegister(fn func(*grpc.Server)) Option {
	return func(o *options) {
		o.grpcRegisters = append(o.grpcRegisters, fn)
	}
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
