package httpapi

import (
	"errors"
	"net/http"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type contractResponse struct {
	ID                    string `json:"id"`
	ContractNumber        string `json:"contract_number"`
	CompanyName           string `json:"company_name"`
	ContractType          string `json:"contract_type"`
	StartDate             string `json:"start_date"`
	EndDate               string `json:"end_date"`
	Status                string `json:"status"`
	Note                  string `json:"note"`
	TerminatedOn          string `json:"terminated_on"`
	TerminationReason     string `json:"termination_reason"`
	RenewedFromContractID string `json:"renewed_from_contract_id"`
}

func (h *Handler) ListContracts(c *gin.Context) {
	if !h.ensureContractService(c) {
		return
	}
	contracts, err := h.contracts.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "查询合同失败")
		return
	}
	items := make([]contractResponse, 0, len(contracts))
	for _, contract := range contracts {
		items = append(items, contractDTO(contract))
	}
	RespondData(c, http.StatusOK, items)
}

func (h *Handler) CreateContract(c *gin.Context) {
	if !h.ensureContractService(c) {
		return
	}
	var input service.ContractInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查合同资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	contract, err := h.contracts.Create(c.Request.Context(), c.Param("id"), input, admin.ID)
	if !h.respondContractError(c, err) {
		return
	}
	RespondData(c, http.StatusCreated, contractDTO(contract))
}

func (h *Handler) UpdateContract(c *gin.Context) {
	if !h.ensureContractService(c) {
		return
	}
	var input service.ContractInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查合同资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	contract, err := h.contracts.Update(c.Request.Context(), c.Param("id"), c.Param("contractId"), input, admin.ID)
	if !h.respondContractError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, contractDTO(contract))
}

func (h *Handler) RenewContract(c *gin.Context) {
	if !h.ensureContractService(c) {
		return
	}
	var input service.ContractInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查合同资料格式")
		return
	}
	admin, _ := CurrentAdmin(c)
	contract, err := h.contracts.Renew(c.Request.Context(), c.Param("id"), c.Param("contractId"), input, admin.ID)
	if !h.respondContractError(c, err) {
		return
	}
	RespondData(c, http.StatusCreated, contractDTO(contract))
}

func (h *Handler) TerminateContract(c *gin.Context) {
	if !h.ensureContractService(c) {
		return
	}
	var input service.TerminateContractInput
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请填写解除日期和原因")
		return
	}
	admin, _ := CurrentAdmin(c)
	err := h.contracts.Terminate(c.Request.Context(), c.Param("id"), c.Param("contractId"), input, admin.ID)
	if !h.respondContractError(c, err) {
		return
	}
	RespondData(c, http.StatusOK, gin.H{"message": "合同已解除"})
}

func (h *Handler) ensureContractService(c *gin.Context) bool {
	if h.contracts != nil {
		return true
	}
	RespondError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "合同服务尚未配置")
	return false
}
func (h *Handler) respondContractError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, service.ErrTalentNotFound):
		RespondError(c, http.StatusNotFound, "TALENT_NOT_FOUND", "未找到人才档案")
	case errors.Is(err, service.ErrContractNotFound):
		RespondError(c, http.StatusNotFound, "CONTRACT_NOT_FOUND", "未找到合同")
	case errors.Is(err, service.ErrActiveContract):
		RespondError(c, http.StatusConflict, "ACTIVE_CONTRACT_EXISTS", "该人才已有履约中合同，请先续约或解除")
	case errors.Is(err, service.ErrContractNumber):
		RespondError(c, http.StatusConflict, "DUPLICATE_CONTRACT_NUMBER", "合同编号已存在")
	case errors.Is(err, service.ErrValidation):
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "请检查合同资料、状态和日期")
	default:
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "操作失败，请稍后重试")
	}
	return false
}
func contractDTO(contract domain.Contract) contractResponse {
	return contractResponse{ID: contract.ID, ContractNumber: contract.ContractNumber, CompanyName: contract.CompanyName, ContractType: contract.ContractType, StartDate: contract.StartDate, EndDate: contract.EndDate, Status: contract.Status, Note: contract.Note, TerminatedOn: contract.TerminatedOn, TerminationReason: contract.TerminationReason, RenewedFromContractID: contract.RenewedFromContractID}
}
