package httpapi

import (
	"net/http"
	"strings"
	"time"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
)

const adminContextKey = "current_admin"

// RequestLogger adds a request ID and records concise request completion data.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		started := time.Now()
		c.Next()

		logger.Info("request completed",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
		)
	}
}

// CORS permits configured browser origins and handles preflight requests.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAuth verifies a bearer token and attaches the administrator to context.
func RequireAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
			c.Abort()
			return
		}

		admin, err := auth.CurrentAdmin(c.Request.Context(), parts[1])
		if err != nil {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "登录已失效，请重新登录")
			c.Abort()
			return
		}
		c.Set(adminContextKey, admin)
		c.Next()
	}
}

// CurrentAdmin returns the authenticated administrator stored by RequireAuth.
func CurrentAdmin(c *gin.Context) (domain.Admin, bool) {
	value, ok := c.Get(adminContextKey)
	if !ok {
		return domain.Admin{}, false
	}
	admin, ok := value.(domain.Admin)
	return admin, ok
}
