// Package httpapi implements the JSON HTTP transport layer.
package httpapi

import "github.com/gin-gonic/gin"

// ErrorDetail identifies a validation field and its human-readable message.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type errorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// RespondData keeps every successful API response structurally consistent.
func RespondData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// RespondError hides implementation details behind stable API error codes.
func RespondError(c *gin.Context, status int, code, message string, details ...ErrorDetail) {
	c.JSON(status, gin.H{
		"error": errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
