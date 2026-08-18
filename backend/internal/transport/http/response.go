// Package httpapi 实现 JSON HTTP 传输层。
package httpapi

import "github.com/gin-gonic/gin"

// ErrorDetail 标识校验字段及其可读错误信息。
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// RespondData 保持所有成功 API 响应的结构一致。
func RespondData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// RespondError 通过稳定的 API 错误码隐藏内部实现细节。
func RespondError(c *gin.Context, status int, code, message string, details ...ErrorDetail) {
	c.JSON(status, gin.H{
		"error": errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
