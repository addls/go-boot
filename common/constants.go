package common

// 应用默认配置
const (
	DefaultVersion     = "v1.0.0" // 默认应用版本
	DefaultStopTimeout = "10s"    // 默认优雅关闭超时
)

// 服务器默认配置
const (
	DefaultGRPCAddr = ":9000" // 默认 gRPC 服务地址
	DefaultHTTPAddr = ":8000" // 默认 HTTP 服务地址
)

// 响应消息
const (
	SuccessMessage = "success" // 成功响应消息
)

// HTTP 状态码
const (
	HTTPStatusOK = 200 // 成功
)
