package registry

import (
	"fmt"

	"github.com/addls/go-boot/config"
	"github.com/addls/go-boot/registry/consul"
	"github.com/addls/go-boot/registry/etcd"
	"github.com/go-kratos/kratos/v2/registry"
)

// NewRegistrar 根据配置创建服务注册中心
// 支持 etcd、consul 等常见注册中心
func NewRegistrar(cfg *config.Discovery) (registry.Registrar, error) {
	if cfg == nil || !cfg.Register {
		return nil, nil
	}

	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("endpoints cannot be empty when register is enabled")
	}

	switch cfg.Type {
	case "etcd":
		return etcd.NewRegistrar(cfg)
	case "consul":
		return consul.NewRegistrar(cfg)
	case "nacos":
		// TODO: 实现 nacos 支持
		return nil, fmt.Errorf("nacos registry not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", cfg.Type)
	}
}

// NewDiscovery 根据配置创建服务发现客户端
// 支持 etcd、consul 等常见注册中心
// 如果 cfg 为 nil，则使用全局配置（通过 config.GetGlobalConfig() 获取）
func NewDiscovery(cfg *config.Discovery) (registry.Discovery, error) {
	// 如果未传入配置，使用全局配置
	if cfg == nil {
		if config.GlobalConfig != nil && config.GlobalConfig.App.Discovery != nil {
			cfg = config.GlobalConfig.App.Discovery
		}
	}

	if cfg == nil || len(cfg.Endpoints) == 0 {
		return nil, nil
	}

	switch cfg.Type {
	case "etcd":
		return etcd.NewDiscovery(cfg)
	case "consul":
		return consul.NewDiscovery(cfg)
	case "nacos":
		// TODO: 实现 nacos 支持
		return nil, fmt.Errorf("nacos discovery not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", cfg.Type)
	}
}
