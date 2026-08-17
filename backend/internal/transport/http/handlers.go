package httpapi

import (
	"net/http"
	"strings"

	"construction-hrms/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler gathers dependencies used by HTTP endpoint methods.
type Handler struct {
	auth      *service.AuthService
	talents   *service.TalentService
	contracts *service.ContractService
	reminders *service.ReminderService
}

func NewHandler(auth *service.AuthService, talents *service.TalentService, contracts *service.ContractService, reminders *service.ReminderService) *Handler {
	return &Handler{auth: auth, talents: talents, contracts: contracts, reminders: reminders}
}

type loginRequest struct {
	Username string `json:"username" binding:"required,max=64"`
	Password string `json:"password" binding:"required,max=256"`
}

func (h *Handler) Health(c *gin.Context) {
	RespondData(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请输入管理员账号和密码")
		return
	}

	token, admin, err := h.auth.Login(c.Request.Context(), strings.TrimSpace(request.Username), request.Password)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "账号或密码错误")
		return
	}

	RespondData(c, http.StatusOK, gin.H{
		"access_token": token,
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
		},
	})
}

func (h *Handler) Logout(c *gin.Context) {
	// JWT is stateless in this first framework stage. The client clears its copy.
	RespondData(c, http.StatusOK, gin.H{"message": "已退出登录"})
}

func (h *Handler) Me(c *gin.Context) {
	admin, ok := CurrentAdmin(c)
	if !ok {
		RespondError(c, http.StatusInternalServerError, "AUTH_CONTEXT_MISSING", "认证上下文不可用")
		return
	}
	RespondData(c, http.StatusOK, gin.H{
		"id":       admin.ID,
		"username": admin.Username,
	})
}

func (h *Handler) PlaceholderPage(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		RespondData(c, http.StatusOK, gin.H{
			"module":  name,
			"status":  "not_implemented",
			"message": "该业务模块将在下一阶段实现",
		})
	}
}
