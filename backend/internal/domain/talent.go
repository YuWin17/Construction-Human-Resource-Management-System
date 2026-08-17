package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TalentStatusActive    = "active"
	TalentStatusSuspended = "suspended"
	TalentStatusArchived  = "archived"
)

// Talent is the core personnel profile. Sensitive fields are only masked by
// list DTOs; detail responses intentionally contain the complete values.
type Talent struct {
	ID                     string `gorm:"primaryKey;size:36"`
	Code                   string `gorm:"size:64"`
	Name                   string `gorm:"not null;size:50"`
	IDCardNumber           string `gorm:"index;not null;size:18"`
	Gender                 string `gorm:"size:16"`
	BirthDate              string `gorm:"size:10"`
	Phone                  string `gorm:"index;not null;size:32"`
	SocialInsuranceStatus  string `gorm:"size:64"`
	NativePlace            string `gorm:"size:255"`
	CurrentLocation        string `gorm:"index;size:255"`
	Education              string `gorm:"size:32"`
	Major                  string `gorm:"size:255"`
	YearsOfExperience      *int
	PrimaryCertificate     string   `gorm:"size:100"`
	Compensation           string   `gorm:"size:64"`
	BIExpiresOn            string   `gorm:"index;size:10"`
	CertificateRenewalNote string   `gorm:"type:text"`
	CooperationIntentions  []string `gorm:"serializer:json"`
	ExpectedLocations      []string `gorm:"serializer:json"`
	Note                   string   `gorm:"type:text"`
	Status                 string   `gorm:"index;not null;size:32"`
	Certificates           []Certificate
	CreatedAt              time.Time `gorm:"not null"`
	UpdatedAt              time.Time `gorm:"index;not null"`
}

func (t *Talent) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}

type CertificateCatalog struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	IsEnabled bool      `gorm:"not null;default:true" json:"is_enabled"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (c *CertificateCatalog) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

type Certificate struct {
	ID                      string    `gorm:"primaryKey;size:36"`
	TalentID                string    `gorm:"uniqueIndex:idx_certificate_talent_number;index;not null;size:36"`
	CatalogID               string    `gorm:"index;size:36"`
	CertificateNameSnapshot string    `gorm:"not null;size:100"`
	Category                string    `gorm:"index;not null;size:32"`
	Specialty               string    `gorm:"size:100"`
	CertificateNumber       string    `gorm:"index;size:100"`
	Issuer                  string    `gorm:"size:255"`
	IssuedDate              string    `gorm:"size:10"`
	ExpiresOn               string    `gorm:"index;size:10"`
	RegistrationStatus      string    `gorm:"not null;size:32"`
	RegisteredCompany       string    `gorm:"size:255"`
	IsAvailable             bool      `gorm:"not null;default:true"`
	Note                    string    `gorm:"type:text"`
	CreatedAt               time.Time `gorm:"not null"`
	UpdatedAt               time.Time `gorm:"not null"`
}

func (c *Certificate) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

type AuditLog struct {
	ID           string    `gorm:"primaryKey;size:36"`
	AdminID      string    `gorm:"index;size:36"`
	Action       string    `gorm:"index;not null;size:64"`
	ResourceType string    `gorm:"index;not null;size:64"`
	ResourceID   string    `gorm:"index;size:36"`
	DisplayName  string    `gorm:"size:255"`
	Summary      string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index;not null"`
}

func (a *AuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}
