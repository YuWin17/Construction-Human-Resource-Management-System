package repository

import (
	"context"

	"construction-hrms/backend/internal/domain"
	"gorm.io/gorm"
)

type ContractRepository struct{ db *gorm.DB }

func NewContractRepository(db *gorm.DB) *ContractRepository { return &ContractRepository{db: db} }
func (r *ContractRepository) DB() *gorm.DB                  { return r.db }

func (r *ContractRepository) ListByTalent(ctx context.Context, talentID string) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.WithContext(ctx).Where("talent_id = ?", talentID).Order("start_date DESC, created_at DESC").Find(&contracts).Error
	return contracts, err
}
