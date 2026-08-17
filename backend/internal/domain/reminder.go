package domain

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Reminder struct {
	ID               string `gorm:"primaryKey;size:36"`
	ReminderType     string `gorm:"uniqueIndex:idx_reminder_source;index;not null"`
	SourceID         string `gorm:"uniqueIndex:idx_reminder_source;not null"`
	TalentID         string `gorm:"index;not null"`
	DueDate          string `gorm:"index;not null"`
	Status           string `gorm:"index;not null;default:pending"`
	HandledAt        *time.Time
	HandledByAdminID string `gorm:"size:36"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r *Reminder) BeforeCreate(_ *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

type SystemSetting struct {
	Key              string `gorm:"primaryKey;size:100"`
	Value            string `gorm:"not null;size:255"`
	UpdatedByAdminID string `gorm:"size:36"`
	UpdatedAt        time.Time
}
