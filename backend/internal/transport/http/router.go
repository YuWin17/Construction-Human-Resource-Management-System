package httpapi

import (
	"net/http"

	"construction-hrms/backend/internal/config"
	"construction-hrms/backend/internal/service"
	"github.com/gin-gonic/gin"
	"log/slog"
)

// NewRouter creates the stage-A route tree. Business endpoints remain explicit
// placeholders so the frontend can establish protected navigation now.
func NewRouter(cfg config.Config, logger *slog.Logger, auth *service.AuthService, services ...any) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), RequestLogger(logger), CORS(cfg.CORSAllowedOrigins))

	router.GET("/healthz", func(c *gin.Context) {
		RespondData(c, http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	var talentService *service.TalentService
	var contractService *service.ContractService
	var reminderService *service.ReminderService
	for _, candidate := range services {
		switch value := candidate.(type) {
		case *service.TalentService:
			talentService = value
		case *service.ContractService:
			contractService = value
		case *service.ReminderService:
			reminderService = value
		}
	}
	handler := NewHandler(auth, talentService, contractService, reminderService)
	api.POST("/auth/login", handler.Login)

	protected := api.Group("")
	protected.Use(RequireAuth(auth))
	protected.POST("/auth/logout", handler.Logout)
	protected.GET("/auth/me", handler.Me)
	protected.GET("/talents", handler.ListTalents)
	protected.POST("/talents", handler.CreateTalent)
	protected.GET("/talents/:id", handler.GetTalent)
	protected.PUT("/talents/:id", handler.UpdateTalent)
	protected.POST("/talents/:id/archive", handler.ArchiveTalent)
	protected.POST("/talents/:id/restore", handler.RestoreTalent)
	protected.DELETE("/talents/:id", handler.DeleteTalent)
	protected.POST("/talents/:id/certificates", handler.AddCertificate)
	protected.PUT("/talents/:id/certificates/:certificateId", handler.UpdateCertificate)
	protected.DELETE("/talents/:id/certificates/:certificateId", handler.DeleteCertificate)
	protected.GET("/talents/:id/audit-logs", handler.ListTalentAuditLogs)
	protected.GET("/talents/:id/contracts", handler.ListContracts)
	protected.POST("/talents/:id/contracts", handler.CreateContract)
	protected.PUT("/talents/:id/contracts/:contractId", handler.UpdateContract)
	protected.POST("/talents/:id/contracts/:contractId/renew", handler.RenewContract)
	protected.POST("/talents/:id/contracts/:contractId/terminate", handler.TerminateContract)
	protected.GET("/certificate-catalogs", handler.ListCertificateCatalogs)

	protected.GET("/dashboard", handler.Dashboard)
	protected.GET("/reminders", handler.ListReminders)
	protected.POST("/reminders/:id/handle", func(c *gin.Context) { handler.HandleReminder(c, "handled") })
	protected.POST("/reminders/:id/ignore", func(c *gin.Context) { handler.HandleReminder(c, "ignored") })
	protected.GET("/audit-logs", handler.ListAuditLogs)
	protected.GET("/settings", handler.GetSettings)
	protected.PUT("/settings/reminders", handler.UpdateSettings)
	protected.GET("/companies", handler.ListCompanies)
	protected.POST("/companies", handler.CreateCompany)
	protected.PUT("/companies/:id", handler.UpdateCompany)
	protected.DELETE("/companies/:id", handler.DeleteCompany)
	protected.POST("/companies/:id/contract-attachment", handler.UploadCompanyContractAttachment)
	protected.GET("/companies/:id/contract-attachment", handler.DownloadCompanyContractAttachment)
	protected.GET("/delivery-orders", handler.ListDeliveryOrders)
	protected.POST("/delivery-orders", handler.CreateDeliveryOrder)
	protected.PUT("/delivery-orders/:id", handler.UpdateDeliveryOrder)
	protected.DELETE("/delivery-orders/:id", handler.DeleteDeliveryOrder)

	return router
}
