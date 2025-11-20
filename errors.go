package goauth

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误代码
type ErrorCode string

const (
	// 认证相关错误
	ErrCodeInvalidParams     ErrorCode = "INVALID_PARAMS"      // 参数无效
	ErrCodeAppNotFound       ErrorCode = "APP_NOT_FOUND"       // 应用不存在
	ErrCodeAppDisabled       ErrorCode = "APP_DISABLED"        // 应用已禁用
	ErrCodeInvalidTimestamp  ErrorCode = "INVALID_TIMESTAMP"   // 时间戳无效
	ErrCodeInvalidSign       ErrorCode = "INVALID_SIGN"        // 签名无效
	ErrCodeIPNotAllowed      ErrorCode = "IP_NOT_ALLOWED"      // IP不在白名单
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED" // 超过速率限制
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"        // 未授权
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"           // 禁止访问
	
	// 系统错误
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR" // 内部错误
	ErrCodeConfigError   ErrorCode = "CONFIG_ERROR"   // 配置错误
)

// AuthError 认证错误
type AuthError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Detail     string    `json:"detail,omitempty"`
	HTTPStatus int       `json:"-"`
}

// Error 实现error接口
func (e *AuthError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAuthError 创建认证错误
func NewAuthError(code ErrorCode, message string, detail string, httpStatus int) *AuthError {
	return &AuthError{
		Code:       code,
		Message:    message,
		Detail:     detail,
		HTTPStatus: httpStatus,
	}
}

// 预定义的错误
var (
	ErrInvalidParams = NewAuthError(
		ErrCodeInvalidParams,
		"请求参数不完整或格式错误",
		"",
		http.StatusBadRequest,
	)
	
	ErrAppNotFound = NewAuthError(
		ErrCodeAppNotFound,
		"应用不存在",
		"",
		http.StatusUnauthorized,
	)
	
	ErrAppDisabled = NewAuthError(
		ErrCodeAppDisabled,
		"应用已被禁用",
		"",
		http.StatusUnauthorized,
	)
	
	ErrInvalidTimestamp = NewAuthError(
		ErrCodeInvalidTimestamp,
		"时间戳无效",
		"时间戳格式错误或超出有效期",
		http.StatusBadRequest,
	)
	
	ErrInvalidSign = NewAuthError(
		ErrCodeInvalidSign,
		"签名验证失败",
		"",
		http.StatusUnauthorized,
	)
	
	ErrIPNotAllowed = NewAuthError(
		ErrCodeIPNotAllowed,
		"IP地址不在白名单中",
		"",
		http.StatusForbidden,
	)
	
	ErrRateLimitExceeded = NewAuthError(
		ErrCodeRateLimitExceeded,
		"请求过于频繁",
		"",
		http.StatusTooManyRequests,
	)
)

// ErrorResponse 标准错误响应格式
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     *AuthError `json:"error"`
	Timestamp int64     `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(err *AuthError, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     err,
		Timestamp: currentTimestamp(),
		RequestID: requestID,
	}
}

// SuccessResponse 标准成功响应格式
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, requestID string) *SuccessResponse {
	return &SuccessResponse{
		Success:   true,
		Data:      data,
		Timestamp: currentTimestamp(),
		RequestID: requestID,
	}
}

func currentTimestamp() int64 {
	return timeNow().Unix()
}
