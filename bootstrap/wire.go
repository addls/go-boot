//go:build wireinject
// +build wireinject

package bootstrap

import (
	"github.com/google/wire"
)

// InitializeApp 初始化应用的所有依赖
// Wire 会根据 ProviderSet 自动生成依赖注入代码
func InitializeApp(service string, opts ...Option) (*App, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
