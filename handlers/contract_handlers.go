package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

	// Ambil data user dari konteks JWT
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	if currentUser.Worker == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only workers can sign contracts", nil)
		return
	}
	workerID := currentUser.Worker.UserID

	contract, err := h.contractService.SignContract(contractID, workerID)
	if err != nil {
        // ... (penanganan error spesifik) ...
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Contract signed successfully", contract)
}

func (h *ContractHandler) DownloadContractPDF(c *gin.Context) {
    contractID := c.Param("id")
    
    // (Tambahkan validasi: pastikan user yang request adalah farmer atau worker dari kontrak ini)

    pdfBuffer, err := h.contractService.GenerateContractPDF(contractID)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate PDF", err)
        return
    }

    fileName := fmt.Sprintf("kontrak_%s.pdf", contractID)
    c.Header("Content-Disposition", "attachment; filename="+fileName)
    c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}
