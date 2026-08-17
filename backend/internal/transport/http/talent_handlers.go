package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type talentListItem struct {
	ID                     string              `json:"id"`
	Code                   string              `json:"code"`
	Name                   string              `json:"name"`
	IDCardNumber           string              `json:"id_card_number"`
	Phone                  string              `json:"phone"`
	SocialInsuranceStatus  string              `json:"social_insurance_status"`
	PrimaryCertificate     string              `json:"primary_certificate"`
	Major                  string              `json:"major"`
	Compensation           string              `json:"compensation"`
	BIExpiresOn            string              `json:"bi_expires_on"`
	CertificateExpiresOn   string              `json:"certificate_expires_on"`
	CertificateRenewalNote string              `json:"certificate_renewal_note"`
	CertificateNames       []string            `json:"certificate_names"`
	CertificateOptions     []certificateOption `json:"certificate_options"`
	SigningStatus          string              `json:"signing_status"`
	MatchStatus            string              `json:"match_status"`
	Status                 string              `json:"status"`
	CurrentLocation        string              `json:"current_location"`
	UpdatedAt              string              `json:"updated_at"`
}

type certificateOption struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

type certificateResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Specialty          string `json:"specialty"`
	CertificateNumber  string `json:"certificate_number"`
	Issuer             string `json:"issuer"`
	IssuedDate         string `json:"issued_date"`
	ExpiresOn          string `json:"expires_on"`
	RegistrationStatus string `json:"registration_status"`
	RegisteredCompany  string `json:"registered_company"`
	IsAvailable        bool   `json:"is_available"`
	SigningStatus      string `json:"signing_status"`
	IsCooperating      bool   `json:"is_cooperating"`
	Note               string `json:"note"`
}

type talentDetailResponse struct {
	ID                     string                `json:"id"`
	Code                   string                `json:"code"`
	Name                   string                `json:"name"`
	IDCardNumber           string                `json:"id_card_number"`
	Gender                 string                `json:"gender"`
	BirthDate              string                `json:"birth_date"`
	Phone                  string                `json:"phone"`
	SocialInsuranceStatus  string                `json:"social_insurance_status"`
	NativePlace            string                `json:"native_place"`
	CurrentLocation        string                `json:"current_location"`
	Education              string                `json:"education"`
	Major                  string                `json:"major"`
	YearsOfExperience      *int                  `json:"years_of_experience"`
	PrimaryCertificate     string                `json:"primary_certificate"`
	Compensation           string                `json:"compensation"`
	BIExpiresOn            string                `json:"bi_expires_on"`
	CertificateRenewalNote string                `json:"certificate_renewal_note"`
	CooperationIntentions  []string              `json:"cooperation_intentions"`
	ExpectedLocations      []string              `json:"expected_locations"`
	Note                   string                `json:"note"`
	Status                 string                `json:"status"`
	SigningStatus          string                `json:"signing_status"`
	Certificates           []certificateResponse `json:"certificates"`
	CreatedAt              string                `json:"created_at"`
	UpdatedAt              string                `json:"updated_at"`
}

func (h *Handler) ListTalents(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	available, hasAvailable := parseBool(c.Query("certificate_available"))
	var availableFilter *bool
	if hasAvailable {
		availableFilter = &available
	}
	talents, total, err := h.talents.List(c.Request.Context(), c.Query("keyword"), c.Query("status"), c.Query("current_location"), c.Query("certificate_name"), c.Query("certificate_category"), availableFilter, page, pageSize)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询人才列表失败")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	h.expireDeliveryOrders(c)
	deliveryStatuses := h.talentDeliveryStatuses(c, talents)
	items := make([]talentListItem, 0, len(talents))
	for _, talent := range talents {
		items = append(items, talentListDTO(talent, deliveryStatuses[talent.ID]))
	}
	RespondData(c, http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) Dashboard(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	overview, err := h.talents.Dashboard(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询仪表盘失败")
		return
	}
	recent := make([]talentListItem, 0, len(overview.RecentTalents))
	for _, talent := range overview.RecentTalents {
		recent = append(recent, talentListDTO(talent, talentDeliveryStatus{}))
	}
	pendingReminderTotal := 0
	if h.reminders != nil {
		if reminders, reminderErr := h.reminders.List(c, "pending"); reminderErr == nil {
			pendingReminderTotal = len(reminders)
		}
	}
	RespondData(c, http.StatusOK, gin.H{"talent_total": overview.TalentTotal, "active_talent_total": overview.ActiveTalentTotal, "signed_talent_total": 0, "unsigned_talent_total": overview.TalentTotal, "certificate_total": overview.CertificateTotal, "pending_reminder_total": pendingReminderTotal, "recent_talents": recent})
}

func (h *Handler) GetTalent(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	talent, err := h.talents.Get(c.Request.Context(), c.Param("id"))
	if !h.respondTalentError(c, err) {
		return
	}
	if h.reminders != nil {
		h.expireDeliveryOrders(c)
	}
	certificateSigningStatuses := h.certificateSigningStatuses(c, talent.ID)
	deliveryStatus := h.talentDeliveryStatuses(c, []domain.Talent{talent})[talent.ID]
	signingStatus := deliveryStatus.SigningStatus
	if signingStatus == "" || signingStatus == deliveryOrderStatusPending {
		signingStatus = signingStatusFromCertificateStatuses(certificateSigningStatuses)
	}
	RespondData(c, http.StatusOK, talentDetailDTO(talent, certificateSigningStatuses, signingStatus))
}

func (h *Handler) CreateTalent(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	var input service.TalentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查人才资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	talent, err := h.talents.Create(c.Request.Context(), input, admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusCreated, talentDetailDTO(talent, nil, "unsigned"))
}

func (h *Handler) UpdateTalent(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	var input service.TalentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查人才资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	talent, err := h.talents.Update(c.Request.Context(), c.Param("id"), input, admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, talentDetailDTO(talent, nil, "unsigned"))
}

func (h *Handler) ArchiveTalent(c *gin.Context) { h.setTalentStatus(c, domain.TalentStatusArchived) }
func (h *Handler) RestoreTalent(c *gin.Context) { h.setTalentStatus(c, domain.TalentStatusActive) }

func (h *Handler) setTalentStatus(c *gin.Context, status string) {
	if !h.ensureTalentService(c) {
		return
	}
	admin, _ := CurrentAdmin(c)
	err := h.talents.SetStatus(c.Request.Context(), c.Param("id"), status, admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, gin.H{"status": status})
}

func (h *Handler) DeleteTalent(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	admin, _ := CurrentAdmin(c)
	err := h.talents.Delete(c.Request.Context(), c.Param("id"), admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, gin.H{"message": "人才档案已删除"})
}

func (h *Handler) AddCertificate(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	var input service.CertificateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查证书资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	certificate, err := h.talents.AddCertificate(c.Request.Context(), c.Param("id"), input, admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusCreated, certificateDTO(certificate, ""))
}

func (h *Handler) UpdateCertificate(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	var input service.CertificateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查证书资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	certificate, err := h.talents.UpdateCertificate(c.Request.Context(), c.Param("id"), c.Param("certificateId"), input, admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, certificateDTO(certificate, ""))
}

func (h *Handler) DeleteCertificate(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	admin, _ := CurrentAdmin(c)
	err := h.talents.DeleteCertificate(c.Request.Context(), c.Param("id"), c.Param("certificateId"), admin.ID)
	if !h.respondTalentError(c, err) {
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListCertificateCatalogs(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	catalogs, err := h.talents.Catalogs(c.Request.Context(), c.Query("enabled") != "false")
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询证书名称失败")
		return
	}
	RespondData(c, http.StatusOK, catalogs)
}

func (h *Handler) ListTalentAuditLogs(c *gin.Context) {
	if !h.ensureTalentService(c) {
		return
	}
	logs, err := h.talents.AuditLogs(c.Request.Context(), c.Param("id"))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询操作记录失败")
		return
	}
	RespondData(c, http.StatusOK, logs)
}

func (h *Handler) ensureTalentService(c *gin.Context) bool {
	if h.talents != nil {
		return true
	}
	RespondError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "人才服务尚未配置")
	return false
}

func (h *Handler) respondTalentError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, service.ErrTalentNotFound):
		RespondError(c, http.StatusNotFound, "TALENT_NOT_FOUND", "未找到人才档案")
	case errors.Is(err, service.ErrTalentCertificateLimit):
		RespondError(c, http.StatusConflict, "TALENT_CERTIFICATE_LIMIT", "一条人才档案只能关联一个证书，请另建人才记录")
	case errors.Is(err, service.ErrValidation):
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查必填项和格式")
	default:
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "操作失败，请稍后重试")
	}
	return false
}

type talentDeliveryStatus struct {
	SigningStatus string
	Matched       bool
}

func (h *Handler) talentDeliveryStatuses(c *gin.Context, talents []domain.Talent) map[string]talentDeliveryStatus {
	result := make(map[string]talentDeliveryStatus, len(talents))
	if len(talents) == 0 || h.reminders == nil {
		return result
	}
	ids := make([]string, 0, len(talents))
	for _, talent := range talents {
		ids = append(ids, talent.ID)
	}
	var rows []struct {
		TalentID string
		Status   string
	}
	h.reminders.DB().WithContext(c).
		Table("delivery_order_talents").
		Select("delivery_order_talents.talent_id, delivery_orders.status").
		Joins("JOIN delivery_orders ON delivery_orders.id = delivery_order_talents.delivery_order_id").
		Where("delivery_order_talents.talent_id IN ?", ids).
		Scan(&rows)
	for _, row := range rows {
		status := result[row.TalentID]
		status.Matched = true
		if deliveryOrderSigningPriority(row.Status) > deliveryOrderSigningPriority(status.SigningStatus) {
			status.SigningStatus = row.Status
		}
		result[row.TalentID] = status
	}
	return result
}

func deliveryOrderSigningPriority(status string) int {
	switch status {
	case deliveryOrderStatusSigned:
		return 3
	case deliveryOrderStatusExpired:
		return 2
	case deliveryOrderStatusPending:
		return 1
	default:
		return 0
	}
}

func talentListDTO(t domain.Talent, deliveryStatus talentDeliveryStatus) talentListItem {
	names := make([]string, 0, len(t.Certificates))
	options := make([]certificateOption, 0, len(t.Certificates))
	certificateExpiresOn := ""
	for _, certificate := range t.Certificates {
		names = append(names, certificate.CertificateNameSnapshot)
		options = append(options, certificateOption{ID: certificate.ID, Name: certificate.CertificateNameSnapshot, Specialty: certificate.Specialty})
		if certificate.ExpiresOn != "" && (certificateExpiresOn == "" || certificate.ExpiresOn < certificateExpiresOn) {
			certificateExpiresOn = certificate.ExpiresOn
		}
	}
	signingStatus := deliveryStatus.SigningStatus
	if signingStatus == "" || signingStatus == deliveryOrderStatusPending {
		signingStatus = "unsigned"
	}
	matchStatus := "unmatched"
	if deliveryStatus.Matched {
		matchStatus = "matched"
	}
	return talentListItem{ID: t.ID, Code: t.Code, Name: t.Name, IDCardNumber: maskIDCard(t.IDCardNumber), Phone: maskPhone(t.Phone), SocialInsuranceStatus: t.SocialInsuranceStatus, PrimaryCertificate: t.PrimaryCertificate, Major: t.Major, Compensation: t.Compensation, BIExpiresOn: t.BIExpiresOn, CertificateExpiresOn: certificateExpiresOn, CertificateRenewalNote: t.CertificateRenewalNote, CertificateNames: names, CertificateOptions: options, SigningStatus: signingStatus, MatchStatus: matchStatus, Status: t.Status, CurrentLocation: t.CurrentLocation, UpdatedAt: t.UpdatedAt.Format(timeLayout)}
}

func (h *Handler) certificateSigningStatuses(c *gin.Context, talentID string) map[string]string {
	if h.reminders == nil {
		return map[string]string{}
	}
	var rows []struct {
		CertificateID string
		Status        string
	}
	h.reminders.DB().WithContext(c).
		Table("delivery_order_talents").
		Select("delivery_order_talents.certificate_id, delivery_orders.status").
		Joins("JOIN delivery_orders ON delivery_orders.id = delivery_order_talents.delivery_order_id").
		Where("delivery_order_talents.talent_id = ? AND delivery_order_talents.certificate_id <> '' AND delivery_orders.status IN (?, ?)", talentID, deliveryOrderStatusSigned, deliveryOrderStatusExpired).
		Scan(&rows)
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		if row.Status == deliveryOrderStatusSigned || result[row.CertificateID] == "" {
			result[row.CertificateID] = row.Status
		}
	}
	return result
}

func signingStatusFromCertificateStatuses(statuses map[string]string) string {
	for _, status := range statuses {
		if status == deliveryOrderStatusSigned {
			return deliveryOrderStatusSigned
		}
	}
	for _, status := range statuses {
		if status == deliveryOrderStatusExpired {
			return deliveryOrderStatusExpired
		}
	}
	return "unsigned"
}

func talentDetailDTO(t domain.Talent, certificateSigningStatuses map[string]string, signingStatus string) talentDetailResponse {
	certificates := make([]certificateResponse, 0, len(t.Certificates))
	for _, certificate := range t.Certificates {
		certificates = append(certificates, certificateDTO(certificate, certificateSigningStatuses[certificate.ID]))
	}
	return talentDetailResponse{ID: t.ID, Code: t.Code, Name: t.Name, IDCardNumber: t.IDCardNumber, Gender: t.Gender, BirthDate: t.BirthDate, Phone: t.Phone, SocialInsuranceStatus: t.SocialInsuranceStatus, NativePlace: t.NativePlace, CurrentLocation: t.CurrentLocation, Education: t.Education, Major: t.Major, YearsOfExperience: t.YearsOfExperience, PrimaryCertificate: t.PrimaryCertificate, Compensation: t.Compensation, BIExpiresOn: t.BIExpiresOn, CertificateRenewalNote: t.CertificateRenewalNote, CooperationIntentions: t.CooperationIntentions, ExpectedLocations: t.ExpectedLocations, Note: t.Note, Status: t.Status, SigningStatus: signingStatus, Certificates: certificates, CreatedAt: t.CreatedAt.Format(timeLayout), UpdatedAt: t.UpdatedAt.Format(timeLayout)}
}

const timeLayout = "2006-01-02 15:04"

func certificateDTO(certificate domain.Certificate, signingStatus string) certificateResponse {
	if signingStatus == "" {
		signingStatus = "unsigned"
	}
	isCooperating := signingStatus == deliveryOrderStatusSigned
	return certificateResponse{ID: certificate.ID, Name: certificate.CertificateNameSnapshot, Specialty: certificate.Specialty, CertificateNumber: certificate.CertificateNumber, Issuer: certificate.Issuer, IssuedDate: certificate.IssuedDate, ExpiresOn: certificate.ExpiresOn, RegistrationStatus: certificate.RegistrationStatus, RegisteredCompany: certificate.RegisteredCompany, IsAvailable: certificate.IsAvailable, SigningStatus: signingStatus, IsCooperating: isCooperating, Note: certificate.Note}
}

func maskIDCard(value string) string {
	if len(value) < 8 {
		return value
	}
	return value[:3] + strings.Repeat("*", len(value)-7) + value[len(value)-4:]
}
func maskPhone(value string) string {
	if len(value) != 11 {
		return value
	}
	return value[:3] + "****" + value[7:]
}
func parseBool(value string) (bool, bool) {
	if value == "true" {
		return true, true
	}
	if value == "false" {
		return false, true
	}
	return false, false
}
