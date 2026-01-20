package bootstrap

import (
	"os"

	"github.com/addls/go-boot/common"
	"github.com/addls/go-boot/config"
	"github.com/addls/go-boot/log"
	"github.com/addls/go-boot/middleware"
	"github.com/addls/go-boot/registry"
	"github.com/addls/go-boot/response"
	"github.com/go-kratos/kratos/v2"
	kratosLog "github.com/go-kratos/kratos/v2/log"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	kratosRegistry "github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// App 应用结构体，包含所有依赖
type App struct {
	*kratos.App
	Config    *config.Config
	Logger    kratosLog.Logger
	Servers   []transport.Server
	Registrar kratosRegistry.Registrar
	Discovery kratosRegistry.Discovery
}

// ProviderSet 是 Wire 的 Provider 集合
var ProviderSet = wire.NewSet(
	// Options
	NewOptions,

	// 配置相关
	NewConfig,

	// 日志相关
	NewLogger,

	// 中间件相关
	NewMiddlewares,

	// 服务器相关
	NewGRPCServer,
	NewHTTPServer,
	NewServers,

	// 注册中心相关
	NewRegistrar,
	NewDiscovery,

	// Kratos App
	NewKratosApp,

	// 最终 App
	NewApp,
)

// NewOptions 创建 options Provider
func NewOptions(opts ...Option) *options {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// NewConfig 创建配置 Provider
func NewConfig(service string, opts *options) (*config.Config, error) {
	bootstrapConfig, err := config.LoadConfig(service, opts.configFile, opts.config)
	if err != nil {
		return nil, err
	}

	// 设置全局配置
	config.SetGlobalConfig(bootstrapConfig)

	return bootstrapConfig, nil
}

// NewLogger 创建日志 Provider
func NewLogger(service string, cfg *config.Config) (kratosLog.Logger, error) {
	logger := log.NewKratosLogger(service, cfg.Log)
	return logger, nil
}

// NewMiddlewares 创建中间件列表 Provider
func NewMiddlewares(cfg *config.Config, logger kratosLog.Logger, opts *options) []kratosMiddleware.Middleware {
	middlewares := []kratosMiddleware.Middleware{
		middleware.Recovery(logger), // 最外层：panic 恢复（必须）
		middleware.Metadata(),       // 元数据传递（必须，用于服务间通信）
	}

	// Tracing 在 Logging 之前，确保日志中包含 trace 信息
	if cfg.Middleware.EnableTracing {
		middlewares = append(middlewares, middleware.Tracing())
	}

	// Logging 必须启用
	middlewares = append(middlewares, middleware.Logging(logger))

	// Metrics 在最内层，记录最终的处理结果
	if cfg.Middleware.EnableMetrics {
		middlewares = append(middlewares, middleware.Metrics())
	}

	// 添加自定义中间件
	if len(opts.customMiddleware) > 0 {
		middlewares = append(middlewares, opts.customMiddleware...)
	}

	return middlewares
}

// NewGRPCServer 创建 gRPC 服务器 Provider
func NewGRPCServer(cfg *config.Config, middlewares []kratosMiddleware.Middleware, opts *options) GRPCServer {
	if cfg.Server.GRPC.Addr == "" {
		return nil
	}

	grpcOpts := []grpc.ServerOption{
		grpc.Address(cfg.Server.GRPC.Addr),
		grpc.Middleware(middlewares...),
	}

	// 配置超时（如果指定）
	if timeout := common.ParseTimeout(cfg.Server.GRPC.Timeout); timeout > 0 {
		grpcOpts = append(grpcOpts, grpc.Timeout(timeout))
	}

	grpcOpts = append(grpcOpts, opts.grpcOpts...)
	grpcSrv := grpc.NewServer(grpcOpts...)

	// 在服务器创建后注册服务
	for _, register := range opts.grpcRegisters {
		register(grpcSrv)
	}

	return GRPCServer(grpcSrv)
}

// NewHTTPServer 创建 HTTP 服务器 Provider
func NewHTTPServer(cfg *config.Config, middlewares []kratosMiddleware.Middleware, opts *options) HTTPServer {
	if cfg.Server.HTTP.Addr == "" {
		return nil
	}

	httpOpts := []http.ServerOption{
		http.Address(cfg.Server.HTTP.Addr),
		http.Middleware(middlewares...),
		http.ResponseEncoder(response.ResponseEncoder()), // 统一响应格式
		http.ErrorEncoder(response.ErrorEncoder()),       // 统一错误格式
	}

	// 配置超时（如果指定）
	if timeout := common.ParseTimeout(cfg.Server.HTTP.Timeout); timeout > 0 {
		httpOpts = append(httpOpts, http.Timeout(timeout))
	}

	httpOpts = append(httpOpts, opts.httpOpts...)
	httpSrv := http.NewServer(httpOpts...)

	// 在服务器创建后注册路由
	for _, register := range opts.httpRegisters {
		register(httpSrv)
	}

	return HTTPServer(httpSrv)
}

// GRPCServer gRPC 服务器类型别名，用于 Wire 依赖注入
type GRPCServer transport.Server

// HTTPServer HTTP 服务器类型别名，用于 Wire 依赖注入
type HTTPServer transport.Server

// NewServers 创建服务器列表 Provider
func NewServers(grpcSrv GRPCServer, httpSrv HTTPServer) []transport.Server {
	var servers []transport.Server
	if grpcSrv != nil {
		servers = append(servers, transport.Server(grpcSrv))
	}
	if httpSrv != nil {
		servers = append(servers, transport.Server(httpSrv))
	}
	return servers
}

// NewRegistrar 创建服务注册中心 Provider
func NewRegistrar(cfg *config.Config) (kratosRegistry.Registrar, error) {
	return registry.NewRegistrar(cfg.App.Discovery)
}

// NewDiscovery 创建服务发现客户端 Provider
func NewDiscovery(cfg *config.Config) (kratosRegistry.Discovery, error) {
	return registry.NewDiscovery(cfg.App.Discovery)
}

// NewKratosApp 创建 Kratos App Provider
func NewKratosApp(service string, cfg *config.Config, logger kratosLog.Logger, servers []transport.Server, registrar kratosRegistry.Registrar, opts *options) (*kratos.App, error) {
	appOpts := []kratos.Option{
		kratos.Name(service),
		kratos.Logger(logger),
	}

	// 配置版本
	version := cfg.App.Version
	if version == "" {
		version = common.DefaultVersion
	}
	appOpts = append(appOpts, kratos.Version(version))

	// 配置优雅关闭超时
	if stopTimeout := common.ParseTimeout(cfg.App.StopTimeout); stopTimeout > 0 {
		appOpts = append(appOpts, kratos.StopTimeout(stopTimeout))
	}

	// 统一信号处理（优雅关闭）
	appOpts = append(appOpts, kratos.Signal(os.Interrupt, os.Kill))

	// 配置服务元数据（如果指定）
	if len(cfg.App.Metadata) > 0 {
		appOpts = append(appOpts, kratos.Metadata(cfg.App.Metadata))
	}

	// 配置服务注册
	if registrar != nil {
		appOpts = append(appOpts, kratos.Registrar(registrar))
	}

	// 添加服务器
	if len(servers) > 0 {
		appOpts = append(appOpts, kratos.Server(servers...))
	}

	// 添加业务代码传入的额外 App 选项
	appOpts = append(appOpts, opts.appOpts...)

	return kratos.New(appOpts...), nil
}

// NewApp 创建最终 App Provider
func NewApp(app *kratos.App, cfg *config.Config, logger kratosLog.Logger, servers []transport.Server, registrar kratosRegistry.Registrar, discovery kratosRegistry.Discovery) *App {
	return &App{
		App:       app,
		Config:    cfg,
		Logger:    logger,
		Servers:   servers,
		Registrar: registrar,
		Discovery: discovery,
	}
}
