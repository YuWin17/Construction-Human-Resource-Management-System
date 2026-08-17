package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ContractStatusActive     = "active"
	ContractStatusExpired    = "expired"
	ContractStatusTerminated = "terminated"
	ContractStatusRenewed    = "renewed"
)

type Contract struct {
	ID                    string    `gorm:"primaryKey;size:36"`
	TalentID              string    `gorm:"index;not null;size:36"`
	ContractNumber        string    `gorm:"uniqueIndex;not null;size:100"`
	CompanyName           string    `gorm:"not null;size:255"`
	ContractType          string    `gorm:"not null;size:32"`
	StartDate             string    `gorm:"size:10;not null"`
	EndDate               string    `gorm:"index;size:10;not null"`
	Status                string    `gorm:"index;not null;size:32"`
	Note                  string    `gorm:"type:text"`
	TerminatedOn          string    `gorm:"size:10"`
	TerminationReason     string    `gorm:"type:text"`
	RenewedFromContractID string    `gorm:"index;size:36"`
	CreatedAt             time.Time `gorm:"not null"`
	UpdatedAt             time.Time `gorm:"not null"`
}

func (c *Contract) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}
