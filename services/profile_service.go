package services

import (
	"errors"
	"fmt"

	"encoding/json"

	"github.com/google/uuid" // <-- Tambahkan ini untuk mem-parsing ID
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type ProfileService interface {
	UpdateProfile(id, name, phoneNumber, profilePicture string) (*models.User, error)
	UpdateRoleDetails(userID string, userRole string, input RoleDetailsInput) (*models.User, error)
	SubmitVerificationDocument(userID uuid.UUID, docType string, filePath string) (*models.UserVerification, error)
	CheckVerificationStatus(userID uuid.UUID, role string) (bool, []string, error)
}

type profileService struct {
	UserRepo             repositories.UserRepository
	UserVerificationRepo repositories.UserVerificationRepository
}

// [PERBAIKI] Perbarui Constructor
func NewProfileService(userRepo repositories.UserRepository, userVerificationRepo repositories.UserVerificationRepository) ProfileService {
	return &profileService{
		UserRepo:             userRepo,
		UserVerificationRepo: userVerificationRepo,
	}
}

type RoleDetailsInput struct {
	Details json.RawMessage `json:"details" binding:"required"`
}

// Definisikan struct input sementara untuk setiap peran
type farmerInput struct {
	Address        *string `json:"address"`
	AdditionalInfo *string `json:"additional_info"`
}

type workerInput struct {
	Skills               []string        `json:"skills"` // Terima sebagai array
	HourlyRate           *float64        `json:"hourly_rate"`
	DailyRate            *float64        `json:"daily_rate"`
	Address              *string         `json:"address"`
	AvailabilitySchedule json.RawMessage `json:"availability_schedule"` // Terima sebagai JSON object
	CurrentLocationLat   *float64        `json:"current_location_lat"`
	CurrentLocationLng   *float64        `json:"current_location_lng"`
	NationalID           *string         `json:"national_id"` // NIK
	BankName             *string         `json:"bank_name"`
	BankAccountNumber    *string         `json:"bank_account_number"`
	BankAccountHolder    *string         `json:"bank_account_holder"`
}

type driverInput struct {
	Address         *string         `json:"company_address"`
	PricingScheme   json.RawMessage `json:"pricing_scheme"`
	VehicleTypes    []string        `json:"vehicle_types"`
	CurrentLat      *float64        `json:"current_lat"` // <-- Diubah dari string ke *float64
	CurrentLng      *float64        `json:"current_lng"`
	BankName        *string         `json:"bank_name"`
	BankAccountNumber *string       `json:"bank_account_number"`
	BankAccountHolder *string       `json:"bank_account_holder"`
}

// UpdateProfile mengurus logika bisnis untuk memperbarui profil pengguna.
func (s *profileService) UpdateProfile(id, name, phoneNumber, profilePicture string) (*models.User, error) {
	// 1. Konversi ID dari string ke tipe uuid.UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		// Jika ID yang diberikan tidak valid, kembalikan error.
		return nil, errors.New("invalid user ID format")
	}

	// 2. Siapkan data yang akan diperbarui dalam sebuah struct User.
	// Kita hanya perlu mengisi ID dan field yang akan diubah.
	// Perhatikan penggunaan '&' untuk tipe data pointer (*string) di model.
	updateData := &models.User{
		ID:             userID,
		Name:           name,
		PhoneNumber:    &phoneNumber,
		ProfilePicture: &profilePicture,
	}

	// 3. Panggil repository untuk melakukan update ke database.
	err = s.UserRepo.UpdateProfile(updateData)
	if err != nil {
		return nil, err // Jika ada error dari DB, teruskan.
	}
	updatedUser, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *profileService) UpdateRoleDetails(userID string, userRole string, input RoleDetailsInput) (*models.User, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Gunakan switch case berdasarkan peran pengguna
	switch userRole {
	case "farmer":
		var details farmerInput
		if err := json.Unmarshal(input.Details, &details); err != nil {
			return nil, fmt.Errorf("invalid farmer details format: %w", err)
		}

		farmerModel := models.Farmer{
			UserID:         parsedUserID,
			Address:        details.Address,
			AdditionalInfo: details.AdditionalInfo,
		}
		if err := s.UserRepo.CreateOrUpdateFarmer(&farmerModel); err != nil {
			return nil, err
		}

	case "worker":
		var details workerInput
		if err := json.Unmarshal(input.Details, &details); err != nil {
			return nil, fmt.Errorf("invalid worker details format: %w", err)
		}

		// Konversi array/object menjadi string JSON sebelum disimpan
		skillsJSON, _ := json.Marshal(details.Skills)
		scheduleJSON, _ := json.Marshal(details.AvailabilitySchedule)

		workerModel := models.Worker{
			UserID:               parsedUserID,
			Skills:               string(skillsJSON),
			HourlyRate:           details.HourlyRate,
			DailyRate:            details.DailyRate,
			Address:              details.Address,
			AvailabilitySchedule: Ptr(string(scheduleJSON)), // Helper Ptr untuk *string
			CurrentLocationLat:   details.CurrentLocationLat,
			CurrentLocationLng:   details.CurrentLocationLng,
			NationalID:           details.NationalID,
			BankName:             details.BankName,
			BankAccountNumber:    details.BankAccountNumber,
			BankAccountHolder:    details.BankAccountHolder,
		}
		if err := s.UserRepo.CreateOrUpdateWorker(&workerModel); err != nil {
			return nil, err
		}

	case "driver":
		var details driverInput
		if err := json.Unmarshal(input.Details, &details); err != nil {
			return nil, fmt.Errorf("invalid expedition details format: %w", err)
		}
		pricingJSON, _ := json.Marshal(details.PricingScheme)
		vehiclesJSON, _ := json.Marshal(details.VehicleTypes)

		driverModel := models.Driver{
			UserID:        parsedUserID,
			Address:       details.Address,
			PricingScheme: string(pricingJSON),
			VehicleTypes:  string(vehiclesJSON),
			CurrentLat:        details.CurrentLat,  
			CurrentLng:        details.CurrentLng, 
		}
		if err := s.UserRepo.CreateOrUpdateDriver(&driverModel); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("role does not support details update")
	}

	// Setelah berhasil, kembalikan profil pengguna yang sudah ter-update
	return s.UserRepo.FindByID(userID)
}

// Helper kecil untuk membuat pointer dari string, berguna untuk field opsional.
func Ptr(s string) *string {
	if s == "" || s == "null" { // Handle jika schedule kosong
		return nil
	}
	return &s
}

func (s *profileService) SubmitVerificationDocument(userID uuid.UUID, docType string, filePath string) (*models.UserVerification, error) {
	verification := &models.UserVerification{
		UserID:       userID,
		DocumentType: docType,
		FilePath:     filePath,
		Status:       "pending",
	}
	err := s.UserVerificationRepo.Create(verification)
	return verification, err
}

// [FUNGSI BARU]
// CheckVerificationStatus memeriksa apakah pengguna sudah mengunggah semua dokumen wajib
func (s *profileService) CheckVerificationStatus(userID uuid.UUID, role string) (bool, []string, error) {
	// Definisikan dokumen wajib untuk setiap peran
	var requiredDocs map[string][]string = map[string][]string{
		"worker": {"KTP", "SELFIE_KTP"},
		"driver": {"KTP", "SELFIE_KTP", "SIM", "STNK", "KIR"}, // KIR bisa ditambahkan
		"farmer": {"KTP", "SKU"},
	}

	required, ok := requiredDocs[role]
	if !ok {
		return true, nil, nil // Tidak butuh verifikasi untuk peran ini (misal: general)
	}

	approved, err := s.UserVerificationRepo.GetApprovedDocumentsForUser(userID)
	if err != nil {
		return false, nil, err
	}

	// Buat map dari dokumen yang sudah disetujui untuk pencarian cepat
	approvedMap := make(map[string]bool)
	for _, doc := range approved {
		approvedMap[doc] = true
	}

	var missingDocs []string
	isFullyVerified := true

	// Periksa apakah semua dokumen wajib sudah ada di map
	for _, reqDoc := range required {
		if !approvedMap[reqDoc] {
			isFullyVerified = false
			missingDocs = append(missingDocs, reqDoc)
		}
	}

	return isFullyVerified, missingDocs, nil
}