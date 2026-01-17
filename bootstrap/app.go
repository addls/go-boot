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
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// Run 启动应用
func Run(service string, opts ...Option) error {
	// 解析选项
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	// 加载配置（优先级：先加载文件配置，然后用 WithConfig 覆写）
	bootstrapConfig, err := config.LoadConfig(service, cfg.configFile, cfg.config)
	if err != nil {
		return err
	}

	// 设置全局配置，业务代码可以随时获取配置，无需重复读取
	config.SetGlobalConfig(bootstrapConfig)

	// 创建日志
	logger := log.NewKratosLogger(service)
	
	// 设置中间件
	middlewares := middleware.DefaultWithConfig(bootstrapConfig.Middleware)
	if len(cfg.customMiddleware) > 0 {
		middlewares = append(middlewares, cfg.customMiddleware...)
	}

	// 创建 Kratos App 选项
	appOpts := []kratos.Option{
		kratos.Name(service),
		kratos.Logger(logger),
	}

	// 配置版本（优先使用配置文件，否则使用默认版本）
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

	// 配置服务注册与发现
	if bootstrapConfig.App.Discovery != nil {
		// 如果 register 为 true，创建并注册服务
		if bootstrapConfig.App.Discovery.Register {
			registrar, err := registry.NewRegistrar(bootstrapConfig.App.Discovery)
			if err != nil {
				return fmt.Errorf("failed to create registrar: %w", err)
			}
			if registrar != nil {
				appOpts = append(appOpts, kratos.Registrar(registrar))
			}
		}
	}

	// 创建 gRPC 服务器
	if bootstrapConfig.Server.GRPC.Addr != "" {
		grpcOpts := []grpc.ServerOption{
			grpc.Address(bootstrapConfig.Server.GRPC.Addr),
			grpc.Middleware(middlewares...),
		}
		// 配置超时（如果指定）
		if timeout := common.ParseTimeout(bootstrapConfig.Server.GRPC.Timeout); timeout > 0 {
			grpcOpts = append(grpcOpts, grpc.Timeout(timeout))
		}
		grpcOpts = append(grpcOpts, cfg.grpcOpts...)
		appOpts = append(appOpts, kratos.Server(grpc.NewServer(grpcOpts...)))
	}

	// 创建 HTTP 服务器
	if bootstrapConfig.Server.HTTP.Addr != "" {
		httpOpts := []http.ServerOption{
			http.Address(bootstrapConfig.Server.HTTP.Addr),
			http.Middleware(middlewares...),
			http.ResponseEncoder(response.ResponseEncoder()), // 统一响应格式
			http.ErrorEncoder(response.ErrorEncoder()),       // 统一错误格式
		}
		// 配置超时（如果指定）
		if timeout := common.ParseTimeout(bootstrapConfig.Server.HTTP.Timeout); timeout > 0 {
			httpOpts = append(httpOpts, http.Timeout(timeout))
		}
		httpOpts = append(httpOpts, cfg.httpOpts...)
		appOpts = append(appOpts, kratos.Server(http.NewServer(httpOpts...)))
	}

	// 添加业务代码传入的额外 App 选项
	appOpts = append(appOpts, cfg.appOpts...)

	return kratos.New(appOpts...).Run()
}
