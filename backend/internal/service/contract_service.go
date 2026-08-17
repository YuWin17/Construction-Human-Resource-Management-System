package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrContractNotFound = errors.New("contract not found")
	ErrActiveContract   = errors.New("active contract exists")
	ErrContractNumber   = errors.New("contract number exists")
)

type ContractInput struct {
	ContractNumber string `json:"contract_number"`
	CompanyName    string `json:"company_name"`
	ContractType   string `json:"contract_type"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Note           string `json:"note"`
}

type TerminateContractInput struct {
	TerminatedOn      string `json:"terminated_on"`
	TerminationReason string `json:"termination_reason"`
}

type ContractService struct {
	repo *repository.ContractRepository
}

func NewContractService(repo *repository.ContractRepository) *ContractService {
	return &ContractService{repo: repo}
}

func (s *ContractService) List(ctx context.Context, talentID string) ([]domain.Contract, error) {
	return s.repo.ListByTalent(ctx, talentID)
}

func (s *ContractService) Create(ctx context.Context, talentID string, input ContractInput, adminID string) (domain.Contract, error) {
	contract, err := contractFromInput(input)
	if err != nil {
		return domain.Contract{}, err
	}
	contract.TalentID = talentID
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var talent domain.Talent
		if err := tx.First(&talent, "id = ?", talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTalentNotFound
			}
			return err
		}
		if talent.Status != domain.TalentStatusActive {
			return ErrValidation
		}
		var count int64
		if err := tx.Model(&domain.Contract{}).Where("talent_id = ? AND status = ?", talentID, domain.ContractStatusActive).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrActiveContract
		}
		if err := tx.Model(&domain.Contract{}).Where("contract_number = ?", contract.ContractNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrContractNumber
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "contract.created", ResourceType: "contract", ResourceID: contract.ID, DisplayName: contract.ContractNumber, Summary: "talent:" + talentID + " 新增签约合同"}).Error
	})
	return contract, err
}

func (s *ContractService) Update(ctx context.Context, talentID, contractID string, input ContractInput, adminID string) (domain.Contract, error) {
	updated, err := contractFromInput(input)
	if err != nil {
		return domain.Contract{}, err
	}
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var contract domain.Contract
		if err := tx.First(&contract, "id = ? AND talent_id = ?", contractID, talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrContractNotFound
			}
			return err
		}
		if contract.Status != domain.ContractStatusActive {
			return ErrValidation
		}
		var count int64
		if err := tx.Model(&domain.Contract{}).Where("id <> ? AND contract_number = ?", contractID, updated.ContractNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrContractNumber
		}
		if err := tx.Model(&contract).Updates(map[string]any{"contract_number": updated.ContractNumber, "company_name": updated.CompanyName, "contract_type": updated.ContractType, "start_date": updated.StartDate, "end_date": updated.EndDate, "note": updated.Note}).Error; err != nil {
			return err
		}
		updated = contract
		updated.ContractNumber = input.ContractNumber
		updated.CompanyName = input.CompanyName
		updated.ContractType = input.ContractType
		updated.StartDate = input.StartDate
		updated.EndDate = input.EndDate
		updated.Note = input.Note
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "contract.updated", ResourceType: "contract", ResourceID: contractID, DisplayName: input.ContractNumber, Summary: "talent:" + talentID + " 更新签约合同"}).Error
	})
	return updated, err
}

func (s *ContractService) Renew(ctx context.Context, talentID, contractID string, input ContractInput, adminID string) (domain.Contract, error) {
	created, err := contractFromInput(input)
	if err != nil {
		return domain.Contract{}, err
	}
	created.TalentID = talentID
	created.RenewedFromContractID = contractID
	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var original domain.Contract
		if err := tx.First(&original, "id = ? AND talent_id = ?", contractID, talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrContractNotFound
			}
			return err
		}
		var count int64
		if err := tx.Model(&domain.Contract{}).Where("talent_id = ? AND status = ? AND id <> ?", talentID, domain.ContractStatusActive, contractID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrActiveContract
		}
		if err := tx.Model(&domain.Contract{}).Where("contract_number = ?", created.ContractNumber).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrContractNumber
		}
		if err := tx.Model(&original).Update("status", domain.ContractStatusRenewed).Error; err != nil {
			return err
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "contract.renewed", ResourceType: "contract", ResourceID: created.ID, DisplayName: created.ContractNumber, Summary: "talent:" + talentID + " 续约合同"}).Error
	})
	return created, err
}

func (s *ContractService) Terminate(ctx context.Context, talentID, contractID string, input TerminateContractInput, adminID string) error {
	terminatedOn, err := parseDate(input.TerminatedOn)
	if err != nil || strings.TrimSpace(input.TerminationReason) == "" {
		return ErrValidation
	}
	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var contract domain.Contract
		if err := tx.First(&contract, "id = ? AND talent_id = ?", contractID, talentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrContractNotFound
			}
			return err
		}
		if contract.Status != domain.ContractStatusActive {
			return ErrValidation
		}
		if terminatedOn.Before(mustDate(contract.StartDate)) {
			return ErrValidation
		}
		if err := tx.Model(&contract).Updates(map[string]any{"status": domain.ContractStatusTerminated, "terminated_on": input.TerminatedOn, "termination_reason": strings.TrimSpace(input.TerminationReason)}).Error; err != nil {
			return err
		}
		return tx.Create(&domain.AuditLog{AdminID: adminID, Action: "contract.terminated", ResourceType: "contract", ResourceID: contractID, DisplayName: contract.ContractNumber, Summary: "talent:" + talentID + " 解除合同"}).Error
	})
}

func contractFromInput(input ContractInput) (domain.Contract, error) {
	start, err := parseDate(input.StartDate)
	if err != nil {
		return domain.Contract{}, ErrValidation
	}
	end, err := parseDate(input.EndDate)
	if err != nil || !end.After(start) || strings.TrimSpace(input.ContractNumber) == "" || strings.TrimSpace(input.CompanyName) == "" || strings.TrimSpace(input.ContractType) == "" || len([]rune(input.Note)) > 1000 {
		return domain.Contract{}, ErrValidation
	}
	return domain.Contract{ContractNumber: strings.TrimSpace(input.ContractNumber), CompanyName: strings.TrimSpace(input.CompanyName), ContractType: strings.TrimSpace(input.ContractType), StartDate: input.StartDate, EndDate: input.EndDate, Note: input.Note, Status: domain.ContractStatusActive}, nil
}

func parseDate(value string) (time.Time, error) { return time.Parse("2006-01-02", value) }
func mustDate(value string) time.Time           { parsed, _ := parseDate(value); return parsed }
