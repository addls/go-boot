package middleware

// Config 中间件配置
type Config struct {
	EnableMetrics bool // 是否启用监控指标
	EnableTracing bool // 是否启用链路追踪
}
