package consul

import (
	"fmt"
	"time"

	"github.com/addls/go-boot/config"
	consulRegistry "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	consulAPI "github.com/hashicorp/consul/api"
)

// NewRegistrar 创建 consul 注册中心
func NewRegistrar(cfg *config.Discovery) (registry.Registrar, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return consulRegistry.New(client), nil
}

// NewDiscovery 创建 consul 服务发现客户端
func NewDiscovery(cfg *config.Discovery) (registry.Discovery, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	// consul Registry 同时实现了 Registrar 和 Discovery 接口
	return consulRegistry.New(client), nil
}

// newClient 创建 consul 客户端
func newClient(cfg *config.Discovery) (*consulAPI.Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("consul endpoints cannot be empty")
	}

	// 解析超时时间
	timeout := 5 * time.Second
	if cfg.Timeout != "" {
		duration, err := time.ParseDuration(cfg.Timeout)
		if err == nil {
			timeout = duration
		}
	}

	// 创建 consul 客户端配置
	consulConfig := consulAPI.DefaultConfig()
	consulConfig.Address = cfg.Endpoints[0] // consul 通常只需要一个地址
	consulConfig.WaitTime = timeout

	// 创建 consul 客户端
	client, err := consulAPI.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return client, nil
}
