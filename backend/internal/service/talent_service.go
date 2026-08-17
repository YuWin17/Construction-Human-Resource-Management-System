package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrTalentNotFound         = errors.New("talent not found")
	ErrValidation             = errors.New("validation error")
	ErrTalentCertificateLimit = errors.New("talent certificate limit reached")
	validIDCard               = regexp.MustCompile(`^\d{17}[\dXx]$`)
	validPhone                = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

type TalentInput struct {
	Name                   string             `json:"name"`
	IDCardNumber           string             `json:"id_card_number"`
	Gender                 string             `json:"gender"`
	BirthDate              string             `json:"birth_date"`
	Phone                  string             `json:"phone"`
	SocialInsuranceStatus  string             `json:"social_insurance_status"`
	NativePlace            string             `json:"native_place"`
	CurrentLocation        string             `json:"current_location"`
	Education              string             `json:"education"`
	Major                  string             `json:"major"`
	YearsOfExperience      *int               `json:"years_of_experience"`
	PrimaryCertificate     string             `json:"primary_certificate"`
	Compensation           string             `json:"compensation"`
	BIExpiresOn            string             `json:"bi_expires_on"`
	CertificateRenewalNote string             `json:"certificate_renewal_note"`
	CooperationIntentions  []string           `json:"cooperation_intentions"`
	ExpectedLocations      []string           `json:"expected_locations"`
	Note                   string             `json:"note"`
	Status                 string             `json:"status"`
	Certificates           []CertificateInput `json:"certificates"`
}

type CertificateInput struct {
	Name               string `json:"name"`
	Category           string `json:"category"`
	Specialty          string `json:"specialty"`
	CertificateNumber  string `json:"certificate_number"`
	Issuer             string `json:"issuer"`
	IssuedDate         string `json:"issued_date"`
	ExpiresOn          string `json:"expires_on"`
	RegistrationStatus string `json:"registration_status"`
	RegisteredCompany  string `json:"registered_company"`
	IsAvailable        *bool  `json:"is_available"`
	Note               string `json:"note"`
}

type TalentService struct{ repo *repository.TalentRepository }

type DashboardOverview struct {
	TalentTotal       int64
	ActiveTalentTotal int64
	CertificateTotal  int64
	RecentTalents     []domain.Talent
}

func NewTalentService(repo *repository.TalentRepository) *TalentService {
	return &TalentService{repo: repo}
}

func (s *TalentService) List(ctx context.Context, keyword, status, currentLocation, certificateName, certificateCategory string, available *bool, page, pageSize int) ([]domain.Talent, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, keyword, status, currentLocation, certificateName, certificateCategory, available, page, pageSize)
}

func (s *TalentService) Dashboard(ctx context.Context) (DashboardOverview, error) {
	total, active, certificates, recent, err := s.repo.DashboardCounts(ctx)
	if err != nil {
		return DashboardOverview{}, err
	}
	return DashboardOverview{TalentTotal: total, ActiveTalentTotal: active, CertificateTotal: certificates, RecentTalents: recent}, nil
}

func (s *TalentService) Get(ctx context.Context, id string) (domain.Talent, error) {
	talent, err := s.repo.Find(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Talent{}, ErrTalentNotFound
	}
	return talent, err
}

func (s *TalentService) Create(ctx context.Context, input TalentInput, adminID string) (domain.Talent, error) {
	if err := validateTalentInput(input); err != nil {
		return domain.Talent{}, err
	}
	if len(input.Certificates) != 1 {
		return domain.Talent{}, ErrValidation
	}
	talent := talentFromInput(input)
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code, err := generateTalentCode(tx)
		if err != nil {
			return err
		}
		talent.Code = code
		if err := tx.Create(&talent).Error; err != nil {
			return err
		}
		if err := createCertificates(tx, talent.ID, input.Certificates); err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "talent.created", ResourceType: "talent", ResourceID: talent.ID, DisplayName: talent.Name, Summary: "创建人才档案"}).Error
	})
	if err != nil {
		return domain.Talent{}, err
	}
	return s.Get(ctx, talent.ID)
}

func (s *TalentService) Update(ctx context.Context, id string, input TalentInput, adminID string) (domain.Talent, error) {
	if err := validateTalentInput(input); err != nil {
		return domain.Talent{}, err
	}
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		updated := talentFromInput(input)
		updated.ID = id
		updated.Code = talent.Code
		if err := tx.Model(&talent).Updates(&updated).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "talent.updated", ResourceType: "talent", ResourceID: id, DisplayName: updated.Name, Summary: "更新人才档案"}).Error
	})
	if err != nil {
		return domain.Talent{}, err
	}
	return s.Get(ctx, id)
}

func (s *TalentService) SetStatus(ctx context.Context, id, status, adminID string) error {
	if status != domain.TalentStatusActive && status != domain.TalentStatusArchived {
		return ErrValidation
	}
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		if err := tx.Model(&talent).Update("status", status).Error; err != nil {
			return err
		}
		action := "talent.archived"
		summary := "归档人才档案"
		if status == domain.TalentStatusActive {
			action = "talent.restored"
			summary = "恢复人才档案"
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: action, ResourceType: "talent", ResourceID: id, DisplayName: talent.Name, Summary: summary}).Error
	})
}

func (s *TalentService) Delete(ctx context.Context, id, adminID string) error {
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		if err := tx.Create(&domain.AuditLog{AdminID: adminID, Action: "talent.deleted", ResourceType: "talent", ResourceID: id, DisplayName: talent.Name, Summary: "已删除人才档案"}).Error; err != nil {
			return err
		}
		if err := tx.Where("talent_id = ?", id).Delete(&domain.Certificate{}).Error; err != nil {
			return err
		}
		return tx.Delete(&talent).Error
	})
}

func (s *TalentService) AddCertificate(ctx context.Context, talentID string, input CertificateInput, adminID string) (domain.Certificate, error) {
	if err := validateCertificateInput(input); err != nil {
		return domain.Certificate{}, err
	}
	var created domain.Certificate
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		var certificateCount int64
		if err := tx.Model(&domain.Certificate{}).Where("talent_id = ?", talentID).Count(&certificateCount).Error; err != nil {
			return err
		}
		if certificateCount >= 1 {
			return ErrTalentCertificateLimit
		}
		certificate, err := buildCertificate(tx, talentID, input)
		if err != nil {
			return err
		}
		created = certificate
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "certificate.created", ResourceType: "certificate", ResourceID: created.ID, DisplayName: created.CertificateNameSnapshot, Summary: "talent:" + talentID + " 新增证书"}).Error
	})
	return created, err
}

func (s *TalentService) UpdateCertificate(ctx context.Context, talentID, certificateID string, input CertificateInput, adminID string) (domain.Certificate, error) {
	if err := validateCertificateInput(input); err != nil {
		return domain.Certificate{}, err
	}
	var certificate domain.Certificate
	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&certificate, "id = ? AND talent_id = ?", certificateID, talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		updated, err := buildCertificate(tx, talentID, input)
		if err != nil {
			return err
		}
		updated.ID = certificate.ID
		if err := tx.Model(&certificate).Updates(&updated).Error; err != nil {
			return err
		}
		certificate = updated
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "certificate.updated", ResourceType: "certificate", ResourceID: certificateID, DisplayName: updated.CertificateNameSnapshot, Summary: "talent:" + talentID + " 更新证书"}).Error
	})
	return certificate, err
}

func (s *TalentService) DeleteCertificate(ctx context.Context, talentID, certificateID, adminID string) error {
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var certificate domain.Certificate
		if err := tx.First(&certificate, "id = ? AND talent_id = ?", certificateID, talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		if err := tx.Delete(&certificate).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "certificate.deleted", ResourceType: "certificate", ResourceID: certificateID, DisplayName: certificate.CertificateNameSnapshot, Summary: "talent:" + talentID + " 删除证书"}).Error
	})
}

func (s *TalentService) Catalogs(ctx context.Context, enabledOnly bool) ([]domain.CertificateCatalog, error) {
	return s.repo.ListCatalogs(ctx, enabledOnly)
}
func (s *TalentService) AuditLogs(ctx context.Context, talentID string) ([]domain.AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, talentID)
}

func validateTalentInput(input TalentInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.IDCardNumber = strings.ToUpper(strings.TrimSpace(input.IDCardNumber))
	input.Phone = strings.TrimSpace(input.Phone)
	if len([]rune(input.Name)) < 2 || len([]rune(input.Name)) > 50 || !validIDCard.MatchString(input.IDCardNumber) || !validPhone.MatchString(input.Phone) {
		return ErrValidation
	}
	if input.Status == "" {
		input.Status = domain.TalentStatusActive
	}
	if input.Status != domain.TalentStatusActive && input.Status != domain.TalentStatusSuspended && input.Status != domain.TalentStatusArchived {
		return ErrValidation
	}
	if input.YearsOfExperience != nil && *input.YearsOfExperience < 0 {
		return ErrValidation
	}
	if len([]rune(input.Note)) > 1000 {
		return ErrValidation
	}
	if len(input.Certificates) > 1 {
		return ErrValidation
	}
	for _, certificate := range input.Certificates {
		if err := validateCertificateInput(certificate); err != nil {
			return err
		}
	}
	return nil
}

func validateCertificateInput(input CertificateInput) error {
	if strings.TrimSpace(input.Name) == "" || len([]rune(input.Note)) > 500 {
		return ErrValidation
	}
	return nil
}

func talentFromInput(input TalentInput) domain.Talent {
	status := input.Status
	if status == "" {
		status = domain.TalentStatusActive
	}
	return domain.Talent{Name: strings.TrimSpace(input.Name), IDCardNumber: strings.ToUpper(strings.TrimSpace(input.IDCardNumber)), Gender: input.Gender, BirthDate: input.BirthDate, Phone: strings.TrimSpace(input.Phone), SocialInsuranceStatus: input.SocialInsuranceStatus, NativePlace: input.NativePlace, CurrentLocation: input.CurrentLocation, Education: input.Education, Major: strings.TrimSpace(input.Major), YearsOfExperience: input.YearsOfExperience, PrimaryCertificate: strings.TrimSpace(input.PrimaryCertificate), Compensation: strings.TrimSpace(input.Compensation), BIExpiresOn: strings.TrimSpace(input.BIExpiresOn), CertificateRenewalNote: strings.TrimSpace(input.CertificateRenewalNote), CooperationIntentions: input.CooperationIntentions, ExpectedLocations: input.ExpectedLocations, Note: input.Note, Status: status}
}

func generateTalentCode(tx *gorm.DB) (string, error) {
	// Match the enterprise-number format: prefix plus yyyyMMddHHmmss.
	for attempt := 0; attempt < 2; attempt++ {
		code := "RC" + time.Now().Format("20060102150405")
		var count int64
		if err := tx.Model(&domain.Talent{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
		// The number format has second-level precision, just like enterprise
		// numbers. Wait for the next second instead of appending a suffix.
		time.Sleep(time.Until(time.Now().Truncate(time.Second).Add(time.Second)))
	}
	return "", ErrValidation
}

func createCertificates(tx *gorm.DB, talentID string, inputs []CertificateInput) error {
	for _, input := range inputs {
		certificate, err := buildCertificate(tx, talentID, input)
		if err != nil {
			return err
		}
		if err := tx.Create(&certificate).Error; err != nil {
			return err
		}
	}
	return nil
}

func buildCertificate(tx *gorm.DB, talentID string, input CertificateInput) (domain.Certificate, error) {
	name := strings.TrimSpace(input.Name)
	var catalog domain.CertificateCatalog
	err := tx.Where("name = ?", name).First(&catalog).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		catalog = domain.CertificateCatalog{Name: name, IsEnabled: true}
		if err := tx.Create(&catalog).Error; err != nil {
			return domain.Certificate{}, err
		}
	} else if err != nil {
		return domain.Certificate{}, err
	}
	available := true
	if input.IsAvailable != nil {
		available = *input.IsAvailable
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "未分类"
	}
	return domain.Certificate{TalentID: talentID, CatalogID: catalog.ID, CertificateNameSnapshot: catalog.Name, Category: category, Specialty: input.Specialty, CertificateNumber: strings.TrimSpace(input.CertificateNumber), Issuer: input.Issuer, IssuedDate: input.IssuedDate, ExpiresOn: input.ExpiresOn, RegistrationStatus: input.RegistrationStatus, RegisteredCompany: input.RegisteredCompany, IsAvailable: available, Note: input.Note}, nil
}
