// Package domain holds business entities shared by the service layer.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Admin is the currently supported system user. The schema allows more
// administrators later even though the first release initializes one account.
type Admin struct {
	ID           string    `gorm:"primaryKey;size:36"`
	Username     string    `gorm:"uniqueIndex;not null;size:64"`
	PasswordHash string    `gorm:"not null;size:255"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// BeforeCreate assigns UUID values inside the application so SQLite and
// PostgreSQL can use the same model without database-specific UUID functions.
func (a *Admin) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}
