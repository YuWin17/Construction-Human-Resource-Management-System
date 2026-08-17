package repository

import (
	"context"
	"strings"

	"construction-hrms/backend/internal/domain"
	"gorm.io/gorm"
)

type TalentRepository struct{ db *gorm.DB }

func NewTalentRepository(db *gorm.DB) *TalentRepository { return &TalentRepository{db: db} }

func (r *TalentRepository) DB() *gorm.DB { return r.db }

func (r *TalentRepository) Find(ctx context.Context, id string) (domain.Talent, error) {
	var talent domain.Talent
	err := r.db.WithContext(ctx).Preload("Certificates").First(&talent, "id = ?", id).Error
	return talent, err
}

func (r *TalentRepository) List(ctx context.Context, keyword, status, currentLocation, certificateName, certificateCategory string, available *bool, page, pageSize int) ([]domain.Talent, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Talent{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR id_card_number LIKE ? OR phone LIKE ?", pattern, pattern, pattern)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if currentLocation != "" {
		query = query.Where("current_location LIKE ?", "%"+currentLocation+"%")
	}
	hasCertificateFilter := certificateName != "" || certificateCategory != "" || available != nil
	if hasCertificateFilter {
		query = query.Joins("JOIN certificates ON certificates.talent_id = talents.id")
		if certificateName != "" {
			query = query.Where("certificates.certificate_name_snapshot LIKE ?", "%"+certificateName+"%")
		}
		if certificateCategory != "" {
			query = query.Where("certificates.category = ?", certificateCategory)
		}
		if available != nil {
			query = query.Where("certificates.is_available = ?", *available)
		}
	}
	var total int64
	countQuery := query
	if hasCertificateFilter {
		countQuery = countQuery.Distinct("talents.id")
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var talents []domain.Talent
	if hasCertificateFilter {
		query = query.Distinct("talents.*")
	}
	err := query.Preload("Certificates").Order("talents.updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&talents).Error
	return talents, total, err
}

func (r *TalentRepository) FindByIDCard(ctx context.Context, idCard string) (domain.Talent, error) {
	var talent domain.Talent
	err := r.db.WithContext(ctx).First(&talent, "id_card_number = ?", idCard).Error
	return talent, err
}

func (r *TalentRepository) ListCatalogs(ctx context.Context, enabledOnly bool) ([]domain.CertificateCatalog, error) {
	query := r.db.WithContext(ctx).Order("sort_order ASC, name ASC")
	if enabledOnly {
		query = query.Where("is_enabled = ?", true)
	}
	var catalogs []domain.CertificateCatalog
	err := query.Find(&catalogs).Error
	return catalogs, err
}

func (r *TalentRepository) ListAuditLogs(ctx context.Context, talentID string) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).Where("resource_id = ? OR summary LIKE ?", talentID, "%talent:"+talentID+"%").Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *TalentRepository) DashboardCounts(ctx context.Context) (total, active, certificates int64, recent []domain.Talent, err error) {
	base := r.db.WithContext(ctx)
	if err = base.Model(&domain.Talent{}).Count(&total).Error; err != nil {
		return
	}
	if err = base.Model(&domain.Talent{}).Where("status = ?", domain.TalentStatusActive).Count(&active).Error; err != nil {
		return
	}
	if err = base.Model(&domain.Certificate{}).Count(&certificates).Error; err != nil {
		return
	}
	err = base.Preload("Certificates").Order("created_at DESC").Limit(5).Find(&recent).Error
	return
}
