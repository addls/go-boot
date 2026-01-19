package config

import (
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/addls/go-boot/common"
	"gopkg.in/yaml.v3"
)

// GlobalConfig 全局配置
// 由 bootstrap 在启动时设置，业务代码可以直接访问
var GlobalConfig *Config

// Config 应用配置
type Config struct {
	Server     Server     `json:"server" yaml:"server"`
	Middleware Middleware `json:"middleware" yaml:"middleware"`
	App        App        `json:"app" yaml:"app"`
	Log        Log        `json:"log" yaml:"log"`
}

// App 应用配置
type App struct {
	Version     string            `json:"version" yaml:"version"`         // 应用版本（可选，默认 v1.0.0）
	StopTimeout string            `json:"stopTimeout" yaml:"stopTimeout"` // 优雅关闭超时（如 "10s", "30s"）
	Discovery   *Discovery        `json:"discovery" yaml:"discovery"`     // 服务注册与发现配置（可选）
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`       // 服务元数据（可选）
}

// Discovery 服务注册与发现配置
type Discovery struct {
	Type      string   `json:"type" yaml:"type"`           // 注册中心类型：etcd, consul, nacos 等
	Register  bool     `json:"register" yaml:"register"`   // 是否开启服务注册（默认 false）
	Endpoints []string `json:"endpoints" yaml:"endpoints"` // 注册中心地址列表（配置后自动连接 Discovery）
	Timeout   string   `json:"timeout" yaml:"timeout"`     // 连接超时（如 "5s"）
}

// Server 服务器配置
type Server struct {
	GRPC ServerConfig `json:"grpc" yaml:"grpc"`
	HTTP ServerConfig `json:"http" yaml:"http"`
}

// ServerConfig 服务器配置项
type ServerConfig struct {
	Addr    string `json:"addr" yaml:"addr"`
	Timeout string `json:"timeout" yaml:"timeout"`
}

// Middleware 中间件配置
type Middleware struct {
	EnableMetrics bool `json:"enableMetrics" yaml:"enableMetrics"`
	EnableTracing bool `json:"enableTracing" yaml:"enableTracing"`
}

// Log 日志配置
type Log struct {
	Output string `json:"output" yaml:"output"` // 日志输出位置：stdout, file, 或文件路径（默认 "logs/app.log"）
	Level  string `json:"level" yaml:"level"`   // 日志级别：debug, info, warn, error（默认 "info"）
}

// Load 加载配置文件
func LoadFile(path string) (*Config, error) {
	// 如果路径为空，使用默认配置
	if path == "" {
		return DefaultConfig(), nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: Server{
			GRPC: ServerConfig{Addr: common.DefaultGRPCAddr},
			HTTP: ServerConfig{Addr: common.DefaultHTTPAddr},
		},
		Middleware: Middleware{
			EnableMetrics: false,
			EnableTracing: false,
		},
		App: App{
			StopTimeout: common.DefaultStopTimeout,
		},
		Log: Log{
			Output: "logs/app.log", // 默认输出到文件
			Level:  "info",
		},
	}
}

// FindConfigFile 查找配置文件
// 按优先级查找：./config.yaml -> ./configs/config.yaml
func FindConfigFile(service string) string {
	paths := []string{
		"config.yaml",
		"configs/config.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// SetGlobalConfig 设置全局配置
// 由 bootstrap 在启动时调用，业务代码无需关心配置
// 注意：配置只在启动时设置一次，之后都是只读操作，无需加锁
func SetGlobalConfig(cfg *Config) {
	GlobalConfig = cfg
}

// LoadConfig 加载配置
// 优先级：先加载文件配置，然后用直接传入的配置覆写
// - service: 服务名称（用于自动查找配置文件）
// - configFile: 指定的配置文件路径（如果为空，则自动查找）
// - directConfig: 直接传入的配置（会覆写文件配置中的对应字段）
func LoadConfig(service, configFile string, directConfig *Config) (*Config, error) {
	var fileConfig *Config
	var err error

	// 先加载文件配置（如果有）
	if configFile != "" {
		// 使用指定的配置文件
		fileConfig, err = LoadFile(configFile)
		if err != nil {
			return nil, err
		}
	} else if path := FindConfigFile(service); path != "" {
		// 自动查找配置文件
		fileConfig, err = LoadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		// 使用默认配置
		fileConfig = DefaultConfig()
	}

	// 如果有直接传入的配置，用它覆写文件配置
	if directConfig != nil {
		if err := mergo.Merge(fileConfig, directConfig, mergo.WithOverride); err != nil {
			fmt.Printf("WARN: failed to merge config: %v\n", err)
		}
	}

	return fileConfig, nil
}
