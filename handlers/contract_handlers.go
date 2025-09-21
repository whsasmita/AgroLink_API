package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ContractHandler struct {
	contractService services.ContractService
}

func NewContractHandler(service services.ContractService) *ContractHandler {
	return &ContractHandler{contractService: service}
}

func (h *ContractHandler) SignContract(c *gin.Context) {
	contractID := c.Param("id")

	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	var secondPartyID uuid.UUID

	switch currentUser.Role {
	case "worker":
		if currentUser.Worker == nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Worker profile not found", nil)
			return
		}
		secondPartyID = currentUser.Worker.UserID

	case "driver":
		if currentUser.Driver == nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Driver profile not found", nil)
			return
		}
		secondPartyID = currentUser.Driver.UserID
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only workers or drivers can sign contracts", nil)
		return
	}

	// Panggil service dengan ID pihak kedua & tipe nya
	response, err := h.contractService.SignContract(contractID, secondPartyID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Contract signed successfully", response)
}


func (h *ContractHandler) GetMyContracts(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)

	contracts, err := h.contractService.GetMyContracts(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve contracts", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Contracts retrieved successfully", contracts)
}


func (h *ContractHandler) DownloadContractPDF(c *gin.Context) {
	contractID := c.Param("id")
	// userInterface, _ := c.Get("user")
	// currentUser := userInterface.(*models.User)

	// TODO: Tambahkan validasi di service untuk memastikan currentUser
	// adalah Petani atau Pekerja yang ada di dalam kontrak ini.
	
	pdfBuffer, err := h.contractService.GenerateContractPDF(contractID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate PDF", err)
		return
	}

	fileName := fmt.Sprintf("kontrak_%s.pdf", contractID)
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}

