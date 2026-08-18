// Package domain 定义服务层共享的业务实体。
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Admin 表示当前支持的系统用户。首版仅初始化一个账号，但表结构支持后续扩展多个管理员。
type Admin struct {
	ID           string    `gorm:"primaryKey;size:36"`
	Username     string    `gorm:"uniqueIndex;not null;size:64"`
	PasswordHash string    `gorm:"not null;size:255"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// BeforeCreate 在应用层生成 UUID，使 SQLite 与 PostgreSQL 无需依赖各自的 UUID 函数。
func (a *Admin) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}
