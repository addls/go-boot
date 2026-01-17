package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
)

// ResponseEncoder 统一响应编码器
// 自动将业务返回的数据包装成统一的 Response 结构
func ResponseEncoder() func(http.ResponseWriter, *http.Request, interface{}) error {
	return func(w http.ResponseWriter, r *http.Request, v interface{}) error {
		// 如果已经是 Response 类型，直接返回
		if resp, ok := v.(*Response); ok {
			return encodeResponse(w, resp)
		}

		// 如果是错误，转换为错误响应
		if err, ok := v.(error); ok {
			return encodeError(w, err)
		}

		// 其他情况包装为成功响应
		return encodeResponse(w, Success(v))
	}
}

// ErrorEncoder 统一错误编码器
// 将 Kratos errors 转换为统一的错误响应格式
func ErrorEncoder() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		encodeError(w, err)
	}
}

// encodeResponse 编码响应
func encodeResponse(w http.ResponseWriter, resp *Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

// encodeError 编码错误
func encodeError(w http.ResponseWriter, err error) error {
	// 尝试从 Kratos errors 中提取信息
	se := errors.FromError(err)
	if se != nil {
		resp := Error(int(se.Code), se.Message)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 统一返回 200，错误信息在 body 中
		return json.NewEncoder(w).Encode(resp)
	}

	// 其他错误
	resp := Error(500, err.Error())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}
