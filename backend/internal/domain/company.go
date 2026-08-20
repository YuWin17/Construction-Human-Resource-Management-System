package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Company struct {
	ID                     string               `gorm:"primaryKey;size:36"`
	Code                   string               `gorm:"uniqueIndex;not null;size:64"`
	Name                   string               `gorm:"index;not null;size:255"`
	ContactName            string               `gorm:"size:64"`
	ContactPhone           string               `gorm:"size:32"`
	OwnerName              string               `gorm:"size:64"`
	ClientType             string               `gorm:"size:32"`
	Note                   string               `gorm:"type:text"`
	ContractAttachmentName string               `gorm:"size:255"`
	ContractAttachmentPath string               `gorm:"size:255"`
	ContractExpiresOn      string               `gorm:"index;size:10"`
	Requirements           []CompanyRequirement `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	MatchStatus            string               `gorm:"-"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (c *Company) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

type CompanyRequirement struct {
	ID         string `gorm:"primaryKey;size:36"`
	CompanyID  string `gorm:"index;not null"`
	Specialty  string `gorm:"not null;size:100"`
	Conditions string `gorm:"size:255"`
	Quantity   int    `gorm:"not null;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (c *CompanyRequirement) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

type DeliveryOrder struct {
	ID                   string                `gorm:"primaryKey;size:36"`
	Code                 string                `gorm:"uniqueIndex;not null;size:64"`
	CompanyID            string                `gorm:"index;not null"`
	RegistrationUnitName string                `gorm:"size:255"`
	UnitNature           string                `gorm:"size:64"`
	Status               string                `gorm:"index;not null"`
	ApprovalStatus       string                `gorm:"size:32;not null;default:pending"`
	ContractExpiresOn    string                `gorm:"index;size:10"`
	PerformanceTotal     float64               `gorm:"not null;default:0"`
	ReceivedTotal        float64               `gorm:"not null;default:0"`
	PaidTotal            float64               `gorm:"not null;default:0"`
	DirectPaymentTotal   float64               `gorm:"not null;default:0"`
	Note                 string                `gorm:"type:text"`
	Talents              []DeliveryOrderTalent `gorm:"foreignKey:DeliveryOrderID;constraint:OnDelete:CASCADE"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (d *DeliveryOrder) BeforeCreate(_ *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return nil
}

type DeliveryOrderTalent struct {
	ID                string  `gorm:"primaryKey;size:36"`
	DeliveryOrderID   string  `gorm:"index;not null"`
	TalentID          string  `gorm:"index;not null" json:"talent_id"`
	CertificateID     string  `gorm:"index;size:36" json:"certificate_id"`
	Specialty         string  `gorm:"size:100" json:"specialty"`
	TalentQuote       float64 `json:"talent_quote"`
	PerformanceAmount float64 `json:"performance_amount"`
	ReceivedAmount    float64 `json:"received_amount"`
	PaidAmount        float64 `json:"paid_amount"`
	CompanyRebate     float64 `json:"company_rebate"`
	DirectPayment     float64 `json:"direct_payment"`
	Note              string  `gorm:"type:text" json:"note"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (d *DeliveryOrderTalent) BeforeCreate(_ *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	return nil
}
