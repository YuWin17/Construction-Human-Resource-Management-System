package httpapi

import (
	"construction-hrms/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	deliveryOrderStatusPending = "pending_signature"
	deliveryOrderStatusSigned  = "signed"
	deliveryOrderStatusExpired = "expired"
)

type deliveryOrderInput struct {
	CompanyID            string                       `json:"company_id"`
	RegistrationUnitName string                       `json:"registration_unit_name"`
	UnitNature           string                       `json:"unit_nature"`
	ContractExpiresOn    string                       `json:"contract_expires_on"`
	Note                 string                       `json:"note"`
	Talents              []domain.DeliveryOrderTalent `json:"talents"`
}

func validateDeliveryOrderInput(in deliveryOrderInput) (string, bool) {
	in.CompanyID = strings.TrimSpace(in.CompanyID)
	in.RegistrationUnitName = strings.TrimSpace(in.RegistrationUnitName)
	in.UnitNature = strings.TrimSpace(in.UnitNature)
	in.ContractExpiresOn = strings.TrimSpace(in.ContractExpiresOn)
	if in.CompanyID == "" || len(in.Talents) == 0 {
		return "请选择企业并至少添加一位人才", false
	}
	if in.RegistrationUnitName == "" || in.UnitNature == "" {
		return "请完整填写注册单位和单位性质", false
	}
	if in.ContractExpiresOn != "" {
		if _, err := time.Parse("2006-01-02", in.ContractExpiresOn); err != nil {
			return "签署到期日格式不正确", false
		}
	}
	seen := map[string]bool{}
	for _, item := range in.Talents {
		if strings.TrimSpace(item.TalentID) == "" || strings.TrimSpace(item.CertificateID) == "" {
			return "请选择人才及其证书", false
		}
		key := item.TalentID + ":" + item.CertificateID
		if seen[key] {
			return "同一送证单不能重复添加同一本证书", false
		}
		seen[key] = true
	}
	return "", true
}

func (h *Handler) validateDeliveryOrderReferences(tx *gorm.DB, in deliveryOrderInput) error {
	// 在调用方事务内校验全部外键，防止送证单引用不属于所选人才的证书。
	var company domain.Company
	if err := tx.First(&company, "id = ?", in.CompanyID).Error; err != nil {
		return err
	}
	for _, item := range in.Talents {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", item.TalentID).Error; err != nil {
			return err
		}
		var certificate domain.Certificate
		if err := tx.First(&certificate, "id = ? AND talent_id = ?", item.CertificateID, item.TalentID).Error; err != nil {
			return err
		}
	}
	return nil
}

func setDeliveryOrderTotals(order *domain.DeliveryOrder, talents []domain.DeliveryOrderTalent) {
	// 汇总金额由明细行计算，不能信任请求体传入的汇总值。
	order.PerformanceTotal = 0
	order.ReceivedTotal = 0
	order.PaidTotal = 0
	order.DirectPaymentTotal = 0
	for _, item := range talents {
		order.PerformanceTotal += item.PerformanceAmount
		order.ReceivedTotal += item.ReceivedAmount
		order.PaidTotal += item.PaidAmount
		order.DirectPaymentTotal += item.DirectPayment
	}
}

func setDeliveryOrderStatus(order *domain.DeliveryOrder) {
	// 状态由签署到期日推导，不是可独立编辑的字段。
	if order.ContractExpiresOn == "" {
		order.Status = deliveryOrderStatusPending
		return
	}
	if order.ContractExpiresOn < time.Now().Format("2006-01-02") {
		order.Status = deliveryOrderStatusExpired
		return
	}
	order.Status = deliveryOrderStatusSigned
}

func (h *Handler) expireDeliveryOrders(c *gin.Context) {
	if h.reminders == nil {
		return
	}
	db := h.reminders.DB().WithContext(c)
	today := time.Now().Format("2006-01-02")
	// 每次读取前同步日期推导的状态，保证列表和筛选结果一致。
	db.Model(&domain.DeliveryOrder{}).Where("contract_expires_on = '' AND status <> ?", deliveryOrderStatusPending).Update("status", deliveryOrderStatusPending)
	db.Model(&domain.DeliveryOrder{}).Where("contract_expires_on <> '' AND contract_expires_on >= ? AND status <> ?", today, deliveryOrderStatusSigned).Update("status", deliveryOrderStatusSigned)
	db.Model(&domain.DeliveryOrder{}).Where("contract_expires_on <> '' AND contract_expires_on < ? AND status <> ?", today, deliveryOrderStatusExpired).Update("status", deliveryOrderStatusExpired)
}

type companyInput struct {
	Name              string                      `json:"name"`
	ContactName       string                      `json:"contact_name"`
	ContactPhone      string                      `json:"contact_phone"`
	ClientType        string                      `json:"client_type"`
	Note              string                      `json:"note"`
	ContractExpiresOn string                      `json:"contract_expires_on"`
	Requirements      []domain.CompanyRequirement `json:"requirements"`
}

func validateCompanyInput(in *companyInput) string {
	in.Name = strings.TrimSpace(in.Name)
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.ContactPhone = strings.TrimSpace(in.ContactPhone)
	in.ContractExpiresOn = strings.TrimSpace(in.ContractExpiresOn)
	if in.Name == "" {
		return "请输入企业名称"
	}
	if len(in.Requirements) != 1 {
		return "请填写一条需求证书"
	}
	in.Requirements[0].Specialty = strings.TrimSpace(in.Requirements[0].Specialty)
	in.Requirements[0].Conditions = strings.TrimSpace(in.Requirements[0].Conditions)
	if in.Requirements[0].Specialty == "" {
		return "请填写需求证书"
	}
	if in.Requirements[0].Quantity < 1 {
		in.Requirements[0].Quantity = 1
	}
	if in.ContractExpiresOn != "" {
		if _, err := time.Parse("2006-01-02", in.ContractExpiresOn); err != nil {
			return "合同到期日格式不正确"
		}
	}
	return ""
}

func (h *Handler) ListCompanies(c *gin.Context) {
	var rows []domain.Company
	q := h.reminders.DB().WithContext(c).Model(&domain.Company{}).Preload("Requirements").Order("companies.updated_at DESC")
	if k := strings.TrimSpace(c.Query("keyword")); k != "" {
		pattern := "%" + k + "%"
		q = q.Where("companies.name LIKE ? OR EXISTS (SELECT 1 FROM company_requirements WHERE company_requirements.company_id = companies.id AND company_requirements.specialty LIKE ?)", pattern, pattern)
	}
	if err := q.Find(&rows).Error; err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询企业列表失败")
		return
	}
	// 匹配状态由是否存在送证单推导，不持久化在企业记录中。
	companyIDs := make([]string, 0, len(rows))
	for _, company := range rows {
		companyIDs = append(companyIDs, company.ID)
	}
	matched := make(map[string]bool, len(companyIDs))
	if len(companyIDs) > 0 {
		var matches []struct{ CompanyID string }
		h.reminders.DB().WithContext(c).Table("delivery_orders").Select("company_id").Where("company_id IN ?", companyIDs).Group("company_id").Scan(&matches)
		for _, match := range matches {
			matched[match.CompanyID] = true
		}
	}
	for index := range rows {
		rows[index].MatchStatus = "unmatched"
		if matched[rows[index].ID] {
			rows[index].MatchStatus = "matched"
		}
	}
	RespondData(c, 200, rows)
}
func (h *Handler) CreateCompany(c *gin.Context) {
	var in companyInput
	if c.ShouldBindJSON(&in) != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "请检查企业资料格式")
		return
	}
	if reason := validateCompanyInput(&in); reason != "" {
		RespondError(c, 400, "VALIDATION_ERROR", reason)
		return
	}
	company := domain.Company{Code: "QY" + time.Now().Format("20060102150405"), Name: in.Name, ContactName: in.ContactName, ContactPhone: in.ContactPhone, ClientType: in.ClientType, Note: in.Note, ContractExpiresOn: in.ContractExpiresOn}
	tx := h.reminders.DB().WithContext(c).Begin()
	if tx.Create(&company).Error != nil {
		tx.Rollback()
		RespondError(c, 409, "DUPLICATE_COMPANY_CODE", "企业编号已存在")
		return
	}
	for _, r := range in.Requirements {
		r.CompanyID = company.ID
		tx.Create(&r)
	}
	a, _ := CurrentAdmin(c)
	tx.Create(&domain.AuditLog{AdminID: a.ID, Action: "company.created", ResourceType: "company", ResourceID: company.ID, DisplayName: company.Name, Summary: "创建企业档案"})
	tx.Commit()
	RespondData(c, http.StatusCreated, company)
}

func (h *Handler) UpdateCompany(c *gin.Context) {
	var in companyInput
	if c.ShouldBindJSON(&in) != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "请检查企业资料格式")
		return
	}
	if reason := validateCompanyInput(&in); reason != "" {
		RespondError(c, 400, "VALIDATION_ERROR", reason)
		return
	}
	db := h.reminders.DB().WithContext(c)
	var company domain.Company
	if err := db.First(&company, "id = ?", c.Param("id")).Error; err != nil {
		RespondError(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "未找到企业")
		return
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&company).Updates(map[string]any{"name": in.Name, "contact_name": in.ContactName, "contact_phone": in.ContactPhone, "client_type": in.ClientType, "note": in.Note, "contract_expires_on": in.ContractExpiresOn}).Error; err != nil {
			return err
		}
		if err := tx.Where("company_id = ?", company.ID).Delete(&domain.CompanyRequirement{}).Error; err != nil {
			return err
		}
		// 企业仅保留一条当前需求，因此在事务内原子替换旧记录。
		requirement := in.Requirements[0]
		requirement.ID = ""
		requirement.CompanyID = company.ID
		if err := tx.Create(&requirement).Error; err != nil {
			return err
		}
		admin, _ := CurrentAdmin(c)
		return tx.Create(&domain.AuditLog{AdminID: admin.ID, Action: "company.updated", ResourceType: "company", ResourceID: company.ID, DisplayName: in.Name, Summary: "更新企业档案"}).Error
	}); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "更新企业失败")
		return
	}
	RespondData(c, http.StatusOK, company)
}

func (h *Handler) DeleteCompany(c *gin.Context) {
	db := h.reminders.DB().WithContext(c)
	var company domain.Company
	if err := db.First(&company, "id = ?", c.Param("id")).Error; err != nil {
		RespondError(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "未找到企业")
		return
	}
	var deliveryOrderCount int64
	if err := db.Model(&domain.DeliveryOrder{}).Where("company_id = ?", company.ID).Count(&deliveryOrderCount).Error; err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "检查关联送证单失败")
		return
	}
	if deliveryOrderCount > 0 {
		RespondError(c, http.StatusConflict, "COMPANY_HAS_DELIVERY_ORDERS", "该企业已有送证单，不能删除")
		return
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ?", company.ID).Delete(&domain.CompanyRequirement{}).Error; err != nil {
			return err
		}
		admin, _ := CurrentAdmin(c)
		if err := tx.Create(&domain.AuditLog{AdminID: admin.ID, Action: "company.deleted", ResourceType: "company", ResourceID: company.ID, DisplayName: company.Name, Summary: "删除企业档案"}).Error; err != nil {
			return err
		}
		return tx.Delete(&company).Error
	}); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "删除企业失败")
		return
	}
	if company.ContractAttachmentPath != "" {
		_ = os.Remove(company.ContractAttachmentPath)
	}
	RespondData(c, http.StatusOK, gin.H{"message": "企业已删除"})
}

const maxCompanyAttachmentSize = 20 << 20

func validCompanyAttachment(fileName string, size int64) bool {
	if size <= 0 || size > maxCompanyAttachmentSize {
		return false
	}
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx", ".xls", ".xlsx":
		return true
	default:
		return false
	}
}

func (h *Handler) UploadCompanyContractAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil || !validCompanyAttachment(file.Filename, file.Size) {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请选择 20 MB 以内的 PDF、图片或 Office 合同附件")
		return
	}
	db := h.reminders.DB().WithContext(c)
	var company domain.Company
	if err := db.First(&company, "id = ?", c.Param("id")).Error; err != nil {
		RespondError(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "未找到企业")
		return
	}
	directory := filepath.Join("uploads", "company-contracts")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建附件目录失败")
		return
	}
	// 使用生成的文件名保存附件，避免同名原文件互相覆盖。
	relativePath := filepath.Join(directory, uuid.NewString()+strings.ToLower(filepath.Ext(file.Filename)))
	if err := c.SaveUploadedFile(file, relativePath); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "上传合同附件失败")
		return
	}
	if company.ContractAttachmentPath != "" {
		_ = os.Remove(company.ContractAttachmentPath)
	}
	if err := db.Model(&company).Updates(map[string]any{"contract_attachment_name": file.Filename, "contract_attachment_path": relativePath}).Error; err != nil {
		_ = os.Remove(relativePath)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "保存合同附件失败")
		return
	}
	admin, _ := CurrentAdmin(c)
	_ = db.Create(&domain.AuditLog{AdminID: admin.ID, Action: "company.attachment_uploaded", ResourceType: "company", ResourceID: company.ID, DisplayName: company.Name, Summary: "上传合同附件：" + file.Filename}).Error
	company.ContractAttachmentName = file.Filename
	company.ContractAttachmentPath = relativePath
	RespondData(c, http.StatusOK, company)
}

func (h *Handler) DownloadCompanyContractAttachment(c *gin.Context) {
	var company domain.Company
	if err := h.reminders.DB().WithContext(c).First(&company, "id = ?", c.Param("id")).Error; err != nil {
		RespondError(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "未找到企业")
		return
	}
	if company.ContractAttachmentPath == "" {
		RespondError(c, http.StatusNotFound, "ATTACHMENT_NOT_FOUND", "该企业未上传合同附件")
		return
	}
	if _, err := os.Stat(company.ContractAttachmentPath); err != nil {
		RespondError(c, http.StatusNotFound, "ATTACHMENT_NOT_FOUND", "合同附件不存在")
		return
	}
	c.FileAttachment(company.ContractAttachmentPath, company.ContractAttachmentName)
}
func (h *Handler) ListDeliveryOrders(c *gin.Context) {
	var rows []domain.DeliveryOrder
	h.expireDeliveryOrders(c)
	q := h.reminders.DB().WithContext(c).Preload("Talents").Order("delivery_orders.created_at DESC")
	if talentID := strings.TrimSpace(c.Query("talent_id")); talentID != "" {
		q = q.Joins("JOIN delivery_order_talents ON delivery_order_talents.delivery_order_id = delivery_orders.id").Where("delivery_order_talents.talent_id = ?", talentID)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("delivery_orders.code LIKE ? OR delivery_orders.registration_unit_name LIKE ? OR EXISTS (SELECT 1 FROM delivery_order_talents JOIN talents ON talents.id = delivery_order_talents.talent_id WHERE delivery_order_talents.delivery_order_id = delivery_orders.id AND talents.name LIKE ?)", pattern, pattern, pattern)
	}
	if companyID := strings.TrimSpace(c.Query("company_id")); companyID != "" {
		q = q.Where("delivery_orders.company_id = ?", companyID)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		q = q.Where("delivery_orders.status = ?", status)
	}
	if expiresFrom := strings.TrimSpace(c.Query("contract_expires_from")); expiresFrom != "" {
		q = q.Where("delivery_orders.contract_expires_on >= ?", expiresFrom)
	}
	if expiresTo := strings.TrimSpace(c.Query("contract_expires_to")); expiresTo != "" {
		q = q.Where("delivery_orders.contract_expires_on <= ?", expiresTo)
	}
	if err := q.Find(&rows).Error; err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询送证单失败")
		return
	}
	RespondData(c, 200, rows)
}
func (h *Handler) CreateDeliveryOrder(c *gin.Context) {
	var in deliveryOrderInput
	if c.ShouldBindJSON(&in) != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "请检查送证单资料格式")
		return
	}
	if reason, ok := validateDeliveryOrderInput(in); !ok {
		RespondError(c, 400, "VALIDATION_ERROR", reason)
		return
	}
	order := domain.DeliveryOrder{Code: "SZ" + time.Now().Format("20060102150405"), CompanyID: in.CompanyID, RegistrationUnitName: in.RegistrationUnitName, UnitNature: in.UnitNature, ContractExpiresOn: in.ContractExpiresOn, Note: in.Note}
	setDeliveryOrderStatus(&order)
	setDeliveryOrderTotals(&order, in.Talents)
	// 单据头、人才明细和操作日志构成一次完整业务操作。
	tx := h.reminders.DB().WithContext(c).Begin()
	if err := h.validateDeliveryOrderReferences(tx, in); err != nil {
		tx.Rollback()
		RespondError(c, http.StatusBadRequest, "REFERENCE_NOT_FOUND", "关联企业或人才不存在")
		return
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		RespondError(c, 500, "INTERNAL_ERROR", "创建送证单失败")
		return
	}
	for _, v := range in.Talents {
		v.ID = ""
		v.DeliveryOrderID = order.ID
		if err := tx.Create(&v).Error; err != nil {
			tx.Rollback()
			RespondError(c, 500, "INTERNAL_ERROR", "保存送证单人才明细失败")
			return
		}
	}
	a, _ := CurrentAdmin(c)
	tx.Create(&domain.AuditLog{AdminID: a.ID, Action: "delivery_order.created", ResourceType: "delivery_order", ResourceID: order.ID, DisplayName: order.Code, Summary: "创建送证单"})
	if err := tx.Commit().Error; err != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "创建送证单失败")
		return
	}
	RespondData(c, http.StatusCreated, order)
}

func (h *Handler) UpdateDeliveryOrder(c *gin.Context) {
	var in deliveryOrderInput
	if c.ShouldBindJSON(&in) != nil {
		RespondError(c, 400, "VALIDATION_ERROR", "请检查送证单资料格式")
		return
	}
	if reason, ok := validateDeliveryOrderInput(in); !ok {
		RespondError(c, 400, "VALIDATION_ERROR", reason)
		return
	}
	tx := h.reminders.DB().WithContext(c).Begin()
	var order domain.DeliveryOrder
	if err := tx.First(&order, "id = ?", c.Param("id")).Error; err != nil {
		tx.Rollback()
		RespondError(c, 404, "DELIVERY_ORDER_NOT_FOUND", "送证单不存在")
		return
	}
	if err := h.validateDeliveryOrderReferences(tx, in); err != nil {
		tx.Rollback()
		RespondError(c, http.StatusBadRequest, "REFERENCE_NOT_FOUND", "关联企业或人才不存在")
		return
	}
	order.CompanyID = in.CompanyID
	order.RegistrationUnitName = in.RegistrationUnitName
	order.UnitNature = in.UnitNature
	order.ContractExpiresOn = in.ContractExpiresOn
	order.Note = in.Note
	setDeliveryOrderTotals(&order, in.Talents)
	setDeliveryOrderStatus(&order)
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		RespondError(c, 500, "INTERNAL_ERROR", "更新送证单失败")
		return
	}
	// 提交的明细集合即为最终状态，因此在同一事务内整体替换。
	if err := tx.Where("delivery_order_id = ?", order.ID).Delete(&domain.DeliveryOrderTalent{}).Error; err != nil {
		tx.Rollback()
		RespondError(c, 500, "INTERNAL_ERROR", "更新送证单人才明细失败")
		return
	}
	for _, item := range in.Talents {
		item.ID = ""
		item.DeliveryOrderID = order.ID
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			RespondError(c, 500, "INTERNAL_ERROR", "更新送证单人才明细失败")
			return
		}
	}
	if order.Status != deliveryOrderStatusSigned && order.Status != deliveryOrderStatusExpired {
		tx.Where("reminder_type = ? AND source_id = ?", "delivery_order_expiry", order.ID).Delete(&domain.Reminder{})
	}
	a, _ := CurrentAdmin(c)
	tx.Create(&domain.AuditLog{AdminID: a.ID, Action: "delivery_order.updated", ResourceType: "delivery_order", ResourceID: order.ID, DisplayName: order.Code, Summary: "编辑送证单"})
	if err := tx.Commit().Error; err != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "更新送证单失败")
		return
	}
	RespondData(c, http.StatusOK, order)
}

func (h *Handler) DeleteDeliveryOrder(c *gin.Context) {
	tx := h.reminders.DB().WithContext(c).Begin()
	var order domain.DeliveryOrder
	if err := tx.First(&order, "id = ?", c.Param("id")).Error; err != nil {
		tx.Rollback()
		RespondError(c, 404, "DELIVERY_ORDER_NOT_FOUND", "送证单不存在")
		return
	}
	if err := tx.Where("delivery_order_id = ?", order.ID).Delete(&domain.DeliveryOrderTalent{}).Error; err != nil {
		tx.Rollback()
		RespondError(c, 500, "INTERNAL_ERROR", "删除送证单失败")
		return
	}
	tx.Where("reminder_type = ? AND source_id = ?", "delivery_order_expiry", order.ID).Delete(&domain.Reminder{})
	if err := tx.Delete(&order).Error; err != nil {
		tx.Rollback()
		RespondError(c, 500, "INTERNAL_ERROR", "删除送证单失败")
		return
	}
	a, _ := CurrentAdmin(c)
	tx.Create(&domain.AuditLog{AdminID: a.ID, Action: "delivery_order.deleted", ResourceType: "delivery_order", ResourceID: order.ID, DisplayName: order.Code, Summary: "删除送证单"})
	if err := tx.Commit().Error; err != nil {
		RespondError(c, 500, "INTERNAL_ERROR", "删除送证单失败")
		return
	}
	RespondData(c, http.StatusOK, gin.H{"message": "送证单已删除"})
}
