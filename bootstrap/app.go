package bootstrap

import (
	"fmt"
	"os"

	"github.com/addls/go-boot/common"
	"github.com/addls/go-boot/config"
	"github.com/addls/go-boot/log"
	"github.com/addls/go-boot/middleware"
	"github.com/addls/go-boot/registry"
	"github.com/addls/go-boot/response"
	"github.com/go-kratos/kratos/v2"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Run 启动应用
func Run(service string, opts ...Option) error {
	// 解析选项并加载配置
	cfg, bootstrapConfig, err := parseOptionsAndLoadConfig(service, opts...)
	if err != nil {
		return err
	}

	// 设置全局配置
	config.SetGlobalConfig(bootstrapConfig)

	// 创建中间件
	middlewares := buildMiddlewares(bootstrapConfig.Middleware, cfg.customMiddleware)

	// 构建基础 App 选项
	appOpts := buildBaseAppOptions(service, bootstrapConfig)

	// 配置服务注册与发现
	appOpts, err = configureDiscovery(appOpts, bootstrapConfig.App.Discovery)
	if err != nil {
		return err
	}

	// 创建并注册服务器
	servers := buildServers(bootstrapConfig, middlewares, cfg)
	if len(servers) > 0 {
		appOpts = append(appOpts, kratos.Server(servers...))
	}

	// 添加业务代码传入的额外 App 选项
	appOpts = append(appOpts, cfg.appOpts...)

	return kratos.New(appOpts...).Run()
}

// parseOptionsAndLoadConfig 解析选项并加载配置
func parseOptionsAndLoadConfig(service string, opts ...Option) (*options, *config.Config, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	bootstrapConfig, err := config.LoadConfig(service, cfg.configFile, cfg.config)
	if err != nil {
		return nil, nil, err
	}

	return cfg, bootstrapConfig, nil
}

// buildMiddlewares 构建中间件列表
// 按照执行顺序：Recovery -> Metadata -> Tracing(可选) -> Logging -> Metrics(可选) -> 自定义中间件
func buildMiddlewares(middlewareConfig config.Middleware, customMiddleware []kratosMiddleware.Middleware) []kratosMiddleware.Middleware {
	middlewares := []kratosMiddleware.Middleware{
		middleware.Recovery(), // 最外层：panic 恢复（必须）
		middleware.Metadata(), // 元数据传递（必须，用于服务间通信）
	}

	// Tracing 在 Logging 之前，确保日志中包含 trace 信息
	if middlewareConfig.EnableTracing {
		middlewares = append(middlewares, middleware.Tracing())
	}

	// Logging 必须启用
	middlewares = append(middlewares, middleware.Logging())

	// Metrics 在最内层，记录最终的处理结果
	if middlewareConfig.EnableMetrics {
		middlewares = append(middlewares, middleware.Metrics())
	}

	// 添加自定义中间件
	if len(customMiddleware) > 0 {
		middlewares = append(middlewares, customMiddleware...)
	}

	return middlewares
}

// buildBaseAppOptions 构建基础 App 选项
func buildBaseAppOptions(service string, bootstrapConfig *config.Config) []kratos.Option {
	logger := log.NewKratosLogger(service, bootstrapConfig.Log)

	appOpts := []kratos.Option{
		kratos.Name(service),
		kratos.Logger(logger),
	}

	// 配置版本
	version := bootstrapConfig.App.Version
	if version == "" {
		version = common.DefaultVersion
	}
	appOpts = append(appOpts, kratos.Version(version))

	// 配置优雅关闭超时
	if stopTimeout := common.ParseTimeout(bootstrapConfig.App.StopTimeout); stopTimeout > 0 {
		appOpts = append(appOpts, kratos.StopTimeout(stopTimeout))
	}

	// 统一信号处理（优雅关闭）
	appOpts = append(appOpts, kratos.Signal(os.Interrupt, os.Kill))

	// 配置服务元数据（如果指定）
	if len(bootstrapConfig.App.Metadata) > 0 {
		appOpts = append(appOpts, kratos.Metadata(bootstrapConfig.App.Metadata))
	}

	return appOpts
}

// configureDiscovery 配置服务注册与发现
func configureDiscovery(appOpts []kratos.Option, discovery *config.Discovery) ([]kratos.Option, error) {
	if discovery == nil || !discovery.Register {
		return appOpts, nil
	}

	registrar, err := registry.NewRegistrar(discovery)
	if err != nil {
		return nil, fmt.Errorf("failed to create registrar: %w", err)
	}
	if registrar != nil {
		appOpts = append(appOpts, kratos.Registrar(registrar))
	}

	return appOpts, nil
}

// buildServers 创建并注册所有服务器
func buildServers(bootstrapConfig *config.Config, middlewares []kratosMiddleware.Middleware, cfg *options) []transport.Server {
	var servers []transport.Server

	// 创建 gRPC 服务器
	if grpcSrv := buildGRPCServer(bootstrapConfig.Server.GRPC, middlewares, cfg); grpcSrv != nil {
		servers = append(servers, grpcSrv)
	}

	// 创建 HTTP 服务器
	if httpSrv := buildHTTPServer(bootstrapConfig.Server.HTTP, middlewares, cfg); httpSrv != nil {
		servers = append(servers, httpSrv)
	}

	return servers
}

// buildGRPCServer 创建 gRPC 服务器
func buildGRPCServer(serverConfig config.ServerConfig, middlewares []kratosMiddleware.Middleware, cfg *options) transport.Server {
	if serverConfig.Addr == "" {
		return nil
	}

	grpcOpts := []grpc.ServerOption{
		grpc.Address(serverConfig.Addr),
		grpc.Middleware(middlewares...),
	}

	// 配置超时（如果指定）
	if timeout := common.ParseTimeout(serverConfig.Timeout); timeout > 0 {
		grpcOpts = append(grpcOpts, grpc.Timeout(timeout))
	}

	grpcOpts = append(grpcOpts, cfg.grpcOpts...)
	grpcSrv := grpc.NewServer(grpcOpts...)

	// 在服务器创建后注册服务
	for _, register := range cfg.grpcRegisters {
		register(grpcSrv)
	}

	return grpcSrv
}

// buildHTTPServer 创建 HTTP 服务器
func buildHTTPServer(serverConfig config.ServerConfig, middlewares []kratosMiddleware.Middleware, cfg *options) transport.Server {
	if serverConfig.Addr == "" {
		return nil
	}

	httpOpts := []http.ServerOption{
		http.Address(serverConfig.Addr),
		http.Middleware(middlewares...),
		http.ResponseEncoder(response.ResponseEncoder()), // 统一响应格式
		http.ErrorEncoder(response.ErrorEncoder()),       // 统一错误格式
	}

	// 配置超时（如果指定）
	if timeout := common.ParseTimeout(serverConfig.Timeout); timeout > 0 {
		httpOpts = append(httpOpts, http.Timeout(timeout))
	}

	httpOpts = append(httpOpts, cfg.httpOpts...)
	httpSrv := http.NewServer(httpOpts...)

	// 在服务器创建后注册路由
	for _, register := range cfg.httpRegisters {
		register(httpSrv)
	}

	return httpSrv
}
