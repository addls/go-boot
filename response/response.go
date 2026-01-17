package response

import "github.com/addls/go-boot/common"

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`            // 状态码：200 表示成功，其他表示失败
	Message string      `json:"msg"`             // 消息
	Data    interface{} `json:"data,omitempty"`  // 数据（成功时返回）
	Error   string      `json:"error,omitempty"` // 错误信息（失败时返回）
}

// Success 创建成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    common.HTTPStatusOK,
		Message: common.SuccessMessage,
		Data:    data,
	}
}

// Error 创建错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Error:   message,
	}
}
