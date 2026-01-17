package common

import "time"

// ParseTimeout 解析超时时间字符串（如 "30s", "1m"）
// 如果解析失败或为空，返回 0（使用 Kratos 默认值）
func ParseTimeout(timeoutStr string) time.Duration {
	if timeoutStr == "" {
		return 0
	}
	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0
	}
	return duration
}
