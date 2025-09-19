package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

// UpdateProfileInput adalah struct untuk memvalidasi input JSON dari body request.
// Menggunakan struct terpisah adalah praktik yang baik agar tidak mengekspos
// semua field dari model database ke API.
type UpdateProfileInput struct {
	Name           string `json:"name" binding:"required"`
	PhoneNumber    string `json:"phone_number"`
	ProfilePicture string `json:"profile_picture"`
}

type ProfileHandler interface {
	UpdateProfile(c *gin.Context)
	UploadProfilePhoto(c *gin.Context)
	UpdateRoleDetails(c *gin.Context)
}

type profileHandler struct {
	service services.ProfileService
}

// NewProfileHandler membuat instance baru dari profileHandler.
func NewProfileHandler(service services.ProfileService) ProfileHandler {
	return &profileHandler{service}
}

// UpdateProfile menangani permintaan untuk memperbarui profil pengguna yang sedang login.
func (h *profileHandler) UpdateProfile(c *gin.Context) {
	// 1. Ambil data pengguna yang sudah diautentikasi dari konteks (ditempatkan oleh middleware).
	// Ini cara paling aman untuk mendapatkan ID pengguna, bukan dari parameter URL.
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}

	// Lakukan type assertion ke *models.User
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	// 2. Bind dan validasi input JSON dari body request.
	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 3. Panggil service untuk melakukan pembaruan.
	// Kita gunakan ID dari pengguna yang sedang login (currentUser.ID).
	updatedUser, err := h.service.UpdateProfile(
		currentUser.ID.String(),
		input.Name,
		input.PhoneNumber,
		input.ProfilePicture,
	)

	if err != nil {
		// Jika service mengembalikan error, teruskan ke respons.
		// (Service kita sudah menangani error 'not found', dll.)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	// 4. Jika berhasil, kirimkan data pengguna yang sudah diperbarui.
	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", updatedUser)
}

func (h *profileHandler) UploadProfilePhoto(c *gin.Context) {
	// 1. Ambil data pengguna dari konteks (disiapkan oleh middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	currentUser := userInterface.(*models.User)

	// 2. Ambil file dari request multipart/form-data
	// "photo" adalah nama field yang harus dikirim oleh klien (misal: dari Postman).
	file, err := c.FormFile("photo")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File not provided or field name is not 'photo'", err)
		return
	}

	// 3. Buat nama file yang unik untuk menghindari tabrakan nama.
	// Format: userID-randomUUID.extension
	// Contoh: 1234-abcd-efgh-5678-random123.jpg
	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s-%s%s", currentUser.ID.String(), uuid.New().String(), extension)
	
	// Tentukan path lengkap untuk menyimpan file
	savePath := filepath.Join("public", "images", "profiles", newFileName)

	// 4. Simpan file yang diunggah ke path yang sudah ditentukan
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err)
		return
	}

	// 5. Buat URL publik yang akan disimpan ke database
	// Pastikan URL ini sesuai dengan konfigurasi static server Anda.
	// Di pengembangan, kita bisa hardcode. Di produksi, ini harus dari config/env.
	// appURL := utils.GetEnv("APP_URL", "http://localhost:8080")
	appUrl := config.AppConfig_.App.APP_URL
	fileURL := fmt.Sprintf("%s/static/images/profiles/%s", appUrl, newFileName)

	// 6. Panggil service untuk memperbarui URL gambar di database
	updatedUser, err := h.service.UpdateProfile(
		currentUser.ID.String(),
		currentUser.Name, // Nama tidak berubah
		*currentUser.PhoneNumber, // Nomor telepon tidak berubah
		fileURL,          // HANYA ProfilePicture yang diperbarui dengan URL baru
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile with new photo URL", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile photo uploaded successfully", updatedUser)
}

func (h *profileHandler) UpdateRoleDetails(c *gin.Context) {
	// 1. Ambil pengguna yang sedang login dari konteks
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	currentUser := userInterface.(*models.User)

	// 2. Bind input JSON ke struct RoleDetailsInput
	var input services.RoleDetailsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 3. Panggil service untuk memproses pembaruan
	updatedUser, err := h.service.UpdateRoleDetails(currentUser.ID.String(), currentUser.Role, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update role details", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Role details updated successfully", updatedUser)
}
