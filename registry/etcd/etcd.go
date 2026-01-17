package etcd

import (
	"fmt"
	"time"

	"github.com/addls/go-boot/config"
	etcdRegistry "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	etcdClient "go.etcd.io/etcd/client/v3"
)

// NewRegistrar 创建 etcd 注册中心
func NewRegistrar(cfg *config.Discovery) (registry.Registrar, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	return etcdRegistry.New(client), nil
}

// NewDiscovery 创建 etcd 服务发现客户端
func NewDiscovery(cfg *config.Discovery) (registry.Discovery, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	// etcd Registry 同时实现了 Registrar 和 Discovery 接口
	return etcdRegistry.New(client), nil
}

// newClient 创建 etcd 客户端
func newClient(cfg *config.Discovery) (*etcdClient.Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	// 解析超时时间
	timeout := 5 * time.Second
	if cfg.Timeout != "" {
		duration, err := time.ParseDuration(cfg.Timeout)
		if err == nil {
			timeout = duration
		}
	}

	// 创建 etcd 客户端
	client, err := etcdClient.New(etcdClient.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return client, nil
}
