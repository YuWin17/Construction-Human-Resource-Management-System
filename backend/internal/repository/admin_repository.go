// Package repository 提供数据库访问实现。
package repository

import (
	"context"
	"errors"

	"construction-hrms/backend/internal/domain"
	"gorm.io/gorm"
)

// AdminRepository 将管理员数据库操作与认证逻辑隔离。
type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (domain.Admin, error) {
	var admin domain.Admin
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error
	return admin, err
}

func (r *AdminRepository) FindByID(ctx context.Context, id string) (domain.Admin, error) {
	var admin domain.Admin
	err := r.db.WithContext(ctx).First(&admin, "id = ?", id).Error
	return admin, err
}

func (r *AdminRepository) Create(ctx context.Context, admin *domain.Admin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

func (r *AdminRepository) HasAny(ctx context.Context) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Admin{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
