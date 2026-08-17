package httpapi

import (
	"construction-hrms/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) ListReminders(c *gin.Context) {
	if h.reminders == nil {
		RespondError(c, 503, "SERVICE_UNAVAILABLE", "提醒服务尚未配置")
		return
	}
	items, err := h.reminders.List(c, c.Query("status"))
	if err != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "查询提醒失败")
		return
	}
	RespondData(c, 200, items)
}
func (h *Handler) HandleReminder(c *gin.Context, status string) {
	if h.reminders == nil {
		RespondError(c, 503, "SERVICE_UNAVAILABLE", "提醒服务尚未配置")
		return
	}
	a, _ := CurrentAdmin(c)
	if err := h.reminders.Handle(c, c.Param("id"), status, a.ID); err != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "处理提醒失败")
		return
	}
	RespondData(c, 200, gin.H{"status": status})
}
func (h *Handler) GetSettings(c *gin.Context) {
	if h.reminders == nil {
		return
	}
	a, b, e := h.reminders.Settings(c)
	if e != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "读取设置失败")
		return
	}
	RespondData(c, 200, gin.H{"contract_reminder_days": a, "certificate_reminder_days": b})
}
func (h *Handler) UpdateSettings(c *gin.Context) {
	var in struct {
		ContractReminderDays    int `json:"contract_reminder_days"`
		CertificateReminderDays int `json:"certificate_reminder_days"`
	}
	if c.ShouldBindJSON(&in) != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "请输入提醒天数")
		return
	}
	a, _ := CurrentAdmin(c)
	if e := h.reminders.UpdateSettings(c, in.ContractReminderDays, in.CertificateReminderDays, a.ID); e != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "提醒天数范围为 1 至 365")
		return
	}
	RespondData(c, http.StatusOK, gin.H{"message": "设置已保存"})
}
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var logs []domain.AuditLog
	if err := h.reminders.DB().WithContext(c).Order("created_at DESC").Limit(200).Find(&logs).Error; err != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "查询操作日志失败")
		return
	}
	items := make([]gin.H, 0, len(logs))
	for _, log := range logs {
		items = append(items, gin.H{"id": log.ID, "action": log.Action, "resource_type": log.ResourceType, "display_name": log.DisplayName, "summary": log.Summary, "created_at": log.CreatedAt.Format(timeLayout)})
	}
	RespondData(c, 200, items)
}
