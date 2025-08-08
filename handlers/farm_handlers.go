package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type FarmHandler struct {
	FarmService services.FarmService
}

func NewFarmHandler(farmService services.FarmService) *FarmHandler {
	return &FarmHandler{
		FarmService: farmService,
	}
}

// CreateFarm handles POST /farms
func (h *FarmHandler) CreateFarm(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: User not found in context", nil)
		return
	}
	// 2. Lakukan Type Assertion untuk mengubah interface{} menjadi *models.User
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error: Failed to process user data", nil)
		return
	}

	// 3. Bind input JSON dari body request
	var input services.CreateFarmInput // Pastikan tipe input ini benar
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 4. Panggil service dengan ID dari pengguna yang sudah diautentikasi.
	//    Sekarang kita bisa mengakses currentUser.ID dengan aman.
	farm, err := h.FarmService.CreateFarm(input, currentUser.ID.String())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create farm", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Farm created successfully", farm)
}

// GetFarmByID handles GET /farms/:id
func (h *FarmHandler) GetFarmByID(c *gin.Context) {
	farmID := c.Param("id")
	if farmID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Farm ID is required", nil)
		return
	}

	farm, err := h.FarmService.FindFarmByID(farmID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Farm not found", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Farm retrieved successfully", farm)
}

/*
GetAllFarms handles GET /farms?farmer_id=uuid
Handles a GET request to the /farms endpoint with a farmer_id query parameter, retrieving a list of farms associated with the specified farmer ID.
*/
func (h *FarmHandler) GetAllFarms(c *gin.Context) {
	farmerID := c.Query("farmer_id")
	if farmerID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Farmer ID is required", nil)
		return
	}

	farms, err := h.FarmService.GetAllFarms(farmerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get farms", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Farms retrieved successfully", farms)
}

// GetMyFarms handles GET /farms/my - gets farms for the authenticated user
func (h *FarmHandler) GetMyFarms(c *gin.Context) {
	// Get farmer ID from authenticated user context
	userInterface, exists := c.Get("user")
	// log.Println(userInterface)

	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	// 2. Lakukan Type Assertion ke tipe data yang benar, yaitu *models.User
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	farms, err := h.FarmService.GetAllFarms(currentUser.ID.String())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get farms", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "My farms retrieved successfully", farms)
}

func (h *FarmHandler) UpdateFarm(c *gin.Context) {
	// 1. Get authenticated user
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: User not found in context", nil)
		return
	}

	// 2. Type assertion to get user data
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error: Failed to process user data", nil)
		return
	}

	// 3. Get farm ID from URL parameter
	farmID := c.Param("id")
	if farmID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Farm ID is required", nil)
		return
	}

	// 4. Bind input JSON from request body
	var input services.CreateFarmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 5. Check if the farm belongs to the current user (optional security check)
	existingFarm, err := h.FarmService.FindFarmByID(farmID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Farm not found", err)
		return
	}

	// Verify ownership - only allow farmer to update their own farms
	if existingFarm.FarmerID.String() != currentUser.ID.String() {
		utils.ErrorResponse(c, http.StatusForbidden, "You can only update your own farms", nil)
		return
	}

	// 6. Call service to update farm
	updatedFarm, err := h.FarmService.UpdateFarm(farmID, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update farm", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Farm updated successfully", updatedFarm)
}

// DeleteFarm handles DELETE /farms/:id
func (h *FarmHandler) DeleteFarm(c *gin.Context) {
	// 1. Get authenticated user
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: User not found in context", nil)
		return
	}

	// 2. Type assertion to get user data
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error: Failed to process user data", nil)
		return
	}

	// 3. Get farm ID from URL parameter
	farmID := c.Param("id")
	if farmID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Farm ID is required", nil)
		return
	}

	// 4. Check if the farm exists and belongs to the current user
	existingFarm, err := h.FarmService.FindFarmByID(farmID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Farm not found", err)
		return
	}

	// Verify ownership - only allow farmer to delete their own farms
	if existingFarm.FarmerID.String() != currentUser.ID.String() {
		utils.ErrorResponse(c, http.StatusForbidden, "You can only delete your own farms", nil)
		return
	}

	// 5. Call service to delete farm
	err = h.FarmService.DeleteFarm(farmID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete farm", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Farm deleted successfully", nil)
}
