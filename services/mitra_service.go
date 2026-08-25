package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------
// MitraProfileService Interface & Implementation
// ---------------------------------------------------------------------

type MitraProfileService interface {
	CreateProfile(userID uuid.UUID, req dto.CreateMitraProfileRequest) (*dto.MitraProfileResponse, error)
	GetMyProfile(userID uuid.UUID) (*dto.MitraProfileResponse, error)
	FindByID(userID uuid.UUID) (*dto.MitraProfileResponse, error)
	FindAllVerified(req dto.PaginationRequest) (*dto.PaginationResponse, error)
	GetPendingVerifications(req dto.PaginationRequest) (*dto.PaginationResponse, error)
	ReviewVerification(adminID uuid.UUID, mitraUserID uuid.UUID, req dto.ReviewMitraVerificationRequest) error
}

type mitraProfileService struct {
	mitraRepo repositories.MitraProfileRepository
	userRepo  repositories.UserRepository
	db        *gorm.DB
}

func NewMitraProfileService(mitraRepo repositories.MitraProfileRepository, userRepo repositories.UserRepository, db *gorm.DB) MitraProfileService {
	return &mitraProfileService{
		mitraRepo: mitraRepo,
		userRepo:  userRepo,
		db:        db,
	}
}

func (s *mitraProfileService) CreateProfile(userID uuid.UUID, req dto.CreateMitraProfileRequest) (*dto.MitraProfileResponse, error) {
	// Check existing profile
	existing, err := s.mitraRepo.FindByUserID(userID)
	if err == nil && existing != nil {
		return nil, errors.New("profile mitra sudah ada untuk pengguna ini")
	}

	// Legal document validation based on jenis_mitra
	jenis := strings.ToLower(req.JenisMitra)
	if jenis == "perusahaan" || jenis == "organisasi" {
		if req.NIB == nil || strings.TrimSpace(*req.NIB) == "" {
			return nil, errors.New("NIB wajib diisi untuk jenis mitra perusahaan/organisasi")
		}
	} else if jenis == "individu" {
		if req.NIKKTP == nil || strings.TrimSpace(*req.NIKKTP) == "" {
			return nil, errors.New("NIK KTP wajib diisi untuk jenis mitra individu")
		}
	} else {
		return nil, errors.New("jenis mitra tidak valid")
	}

	profile := &models.MitraProfile{
		UserID:                 userID,
		JenisMitra:             req.JenisMitra,
		NamaMitra:              req.NamaMitra,
		DeskripsiSingkat:       req.DeskripsiSingkat,
		NomorTeleponBisnis:     req.NomorTeleponBisnis,
		EmailBisnis:            req.EmailBisnis,
		Website:                req.Website,
		AlamatLengkap:          req.AlamatLengkap,
		Provinsi:               req.Provinsi,
		KotaKabupaten:          req.KotaKabupaten,
		NPWP:                   req.NPWP,
		NIB:                    req.NIB,
		NIKKTP:                 req.NIKKTP,
		DokumenLegalitas:       req.DokumenLegalitas,
		StatusVerifikasi:       "pending",
		NamaBank:               req.NamaBank,
		NomorRekening:          req.NomorRekening,
		AtasNamaRekening:      req.AtasNamaRekening,
		LogoMitra:              req.LogoMitra,
		RatingMitra:            0.0,
		TotalTransaksiBerhasil: 0,
	}

	if err := s.mitraRepo.Create(nil, profile); err != nil {
		return nil, fmt.Errorf("gagal membuat profil mitra: %w", err)
	}

	return s.toMitraProfileResponse(profile), nil
}

func (s *mitraProfileService) GetMyProfile(userID uuid.UUID) (*dto.MitraProfileResponse, error) {
	profile, err := s.mitraRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("profil mitra belum dibuat")
	}
	return s.toMitraProfileResponse(profile), nil
}

func (s *mitraProfileService) FindByID(userID uuid.UUID) (*dto.MitraProfileResponse, error) {
	profile, err := s.mitraRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("profil mitra tidak ditemukan")
	}
	return s.toMitraProfileResponse(profile), nil
}

func (s *mitraProfileService) FindAllVerified(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	profiles, total, err := s.mitraRepo.FindAllVerified(req)
	if err != nil {
		return nil, err
	}

	var respList []dto.MitraProfileResponse
	for _, p := range profiles {
		respList = append(respList, *s.toMitraProfileResponse(&p))
	}

	totalPages := int(total) / req.Limit
	if int(total)%req.Limit != 0 {
		totalPages++
	}

	return &dto.PaginationResponse{
		Data:       respList,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *mitraProfileService) GetPendingVerifications(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	profiles, total, err := s.mitraRepo.FindPendingVerifications(req)
	if err != nil {
		return nil, err
	}

	var respList []dto.MitraProfileResponse
	for _, p := range profiles {
		respList = append(respList, *s.toMitraProfileResponse(&p))
	}

	totalPages := int(total) / req.Limit
	if int(total)%req.Limit != 0 {
		totalPages++
	}

	return &dto.PaginationResponse{
		Data:       respList,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *mitraProfileService) ReviewVerification(adminID uuid.UUID, mitraUserID uuid.UUID, req dto.ReviewMitraVerificationRequest) error {
	profile, err := s.mitraRepo.FindByUserID(mitraUserID)
	if err != nil {
		return errors.New("profil mitra tidak ditemukan")
	}

	return s.mitraRepo.UpdateStatus(nil, profile.UserID, req.Status)
}

func (s *mitraProfileService) toMitraProfileResponse(p *models.MitraProfile) *dto.MitraProfileResponse {
	if p == nil {
		return nil
	}
	return &dto.MitraProfileResponse{
		UserID:                 p.UserID,
		JenisMitra:             p.JenisMitra,
		NamaMitra:              p.NamaMitra,
		DeskripsiSingkat:       p.DeskripsiSingkat,
		NomorTeleponBisnis:     p.NomorTeleponBisnis,
		EmailBisnis:            p.EmailBisnis,
		Website:                p.Website,
		AlamatLengkap:          p.AlamatLengkap,
		Provinsi:               p.Provinsi,
		KotaKabupaten:          p.KotaKabupaten,
		NPWP:                   p.NPWP,
		NIB:                    p.NIB,
		NIKKTP:                 p.NIKKTP,
		DokumenLegalitas:       p.DokumenLegalitas,
		StatusVerifikasi:       p.StatusVerifikasi,
		NamaBank:               p.NamaBank,
		NomorRekening:          p.NomorRekening,
		AtasNamaRekening:      p.AtasNamaRekening,
		LogoMitra:              p.LogoMitra,
		RatingMitra:            p.RatingMitra,
		TotalTransaksiBerhasil: p.TotalTransaksiBerhasil,
		CreatedAt:              p.CreatedAt,
	}
}

// ---------------------------------------------------------------------
// MitraCooperationService Interface & Implementation
// ---------------------------------------------------------------------

type MitraCooperationService interface {
	CreateOffer(mitraUserID uuid.UUID, req dto.CreateOfferRequest) (*dto.CooperationDetailResponse, error)
	CreateApplication(farmerUserID uuid.UUID, req dto.CreateApplicationRequest) (*dto.CooperationDetailResponse, error)
	FindMyCooperations(userID uuid.UUID, role string, req dto.PaginationRequest) (*dto.PaginationResponse, error)
	FindByID(cooperationID string, userID uuid.UUID) (*dto.CooperationDetailResponse, error)
	ReviewCooperation(cooperationID string, userID uuid.UUID, req dto.ReviewCooperationRequest) error
	ApproveCooperation(cooperationID string, userID uuid.UUID, req dto.ApproveCooperationRequest) error
	RejectCooperation(cooperationID string, userID uuid.UUID, req dto.RejectCooperationRequest) error
	CreateReview(cooperationID string, reviewerID uuid.UUID, reviewerRole string, req dto.CreateMitraReviewRequest) error
}

type mitraCooperationService struct {
	coopRepo    repositories.MitraCooperationRepository
	mitraRepo   repositories.MitraProfileRepository
	invoiceRepo repositories.InvoiceRepository
	reviewRepo  repositories.MitraReviewRepository
	userRepo    repositories.UserRepository
	db          *gorm.DB
}

func NewMitraCooperationService(
	coopRepo repositories.MitraCooperationRepository,
	mitraRepo repositories.MitraProfileRepository,
	invoiceRepo repositories.InvoiceRepository,
	reviewRepo repositories.MitraReviewRepository,
	userRepo repositories.UserRepository,
	db *gorm.DB,
) MitraCooperationService {
	return &mitraCooperationService{
		coopRepo:    coopRepo,
		mitraRepo:   mitraRepo,
		invoiceRepo: invoiceRepo,
		reviewRepo:  reviewRepo,
		userRepo:    userRepo,
		db:          db,
	}
}

func (s *mitraCooperationService) CreateOffer(mitraUserID uuid.UUID, req dto.CreateOfferRequest) (*dto.CooperationDetailResponse, error) {
	// Verify mitra status
	mitra, err := s.mitraRepo.FindByUserID(mitraUserID)
	if err != nil || mitra.StatusVerifikasi != "verified" {
		return nil, errors.New("akun Mitra belum terverifikasi oleh Admin")
	}

	// Verify farmer existence
	farmerUser, err := s.userRepo.FindByID(req.FarmerID.String())
	if err != nil || farmerUser.Role != "farmer" {
		return nil, errors.New("petani tujuan tidak ditemukan")
	}

	agreed := req.ProposedAmount
	coop := &models.MitraCooperation{
		MitraID:                mitraUserID,
		FarmerID:               req.FarmerID,
		InitiatorType:          "mitra",
		Title:                  req.Title,
		Description:            req.Description,
		ProposedAmount:         req.ProposedAmount,
		AgreedAmount:           &agreed,
		PlatformFeePercentage: 11.00,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		Status:                 "submitted",
		Notes:                  req.Notes,
	}

	if err := s.coopRepo.Create(nil, coop); err != nil {
		return nil, fmt.Errorf("gagal membuat penawaran kerjasama: %w", err)
	}

	return s.FindByID(coop.ID.String(), mitraUserID)
}

func (s *mitraCooperationService) CreateApplication(farmerUserID uuid.UUID, req dto.CreateApplicationRequest) (*dto.CooperationDetailResponse, error) {
	// Verify target mitra status
	mitra, err := s.mitraRepo.FindByUserID(req.MitraID)
	if err != nil || mitra.StatusVerifikasi != "verified" {
		return nil, errors.New("mitra tujuan belum terverifikasi oleh Admin")
	}

	agreed := req.ProposedAmount
	coop := &models.MitraCooperation{
		MitraID:                req.MitraID,
		FarmerID:               farmerUserID,
		InitiatorType:          "farmer",
		Title:                  req.Title,
		Description:            req.Description,
		ProposedAmount:         req.ProposedAmount,
		AgreedAmount:           &agreed,
		PlatformFeePercentage: 11.00,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		Status:                 "submitted",
		Notes:                  req.Notes,
	}

	if err := s.coopRepo.Create(nil, coop); err != nil {
		return nil, fmt.Errorf("gagal membuat pengajuan kerjasama: %w", err)
	}

	return s.FindByID(coop.ID.String(), farmerUserID)
}

func (s *mitraCooperationService) FindMyCooperations(userID uuid.UUID, role string, req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	var coops []models.MitraCooperation
	var total int64
	var err error

	if role == "mitra" {
		coops, total, err = s.coopRepo.FindAllByMitraID(userID, req)
	} else {
		coops, total, err = s.coopRepo.FindAllByFarmerID(userID, req)
	}

	if err != nil {
		return nil, err
	}

	var briefList []dto.CooperationBriefResponse
	for _, c := range coops {
		briefList = append(briefList, *s.toBriefResponse(&c))
	}

	totalPages := int(total) / req.Limit
	if int(total)%req.Limit != 0 {
		totalPages++
	}

	return &dto.PaginationResponse{
		Data:       briefList,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *mitraCooperationService) FindByID(cooperationID string, userID uuid.UUID) (*dto.CooperationDetailResponse, error) {
	coop, err := s.coopRepo.FindByIDWithDetails(cooperationID)
	if err != nil {
		return nil, errors.New("kerjasama tidak ditemukan")
	}

	// Verify authorization (Must be Farmer, Mitra, or Admin)
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, errors.New("pengguna tidak valid")
	}

	if user.Role != "admin" && coop.MitraID != userID && coop.FarmerID != userID {
		return nil, errors.New("akses ditolak: Anda bukan bagian dari kerjasama ini")
	}

	return s.toDetailResponse(coop), nil
}

func (s *mitraCooperationService) ReviewCooperation(cooperationID string, userID uuid.UUID, req dto.ReviewCooperationRequest) error {
	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		return errors.New("kerjasama tidak ditemukan")
	}

	if !s.isRecipientParty(coop, userID) {
		return errors.New("hanya pihak penerima pengajuan yang dapat melakukan peninjauan")
	}

	if coop.Status != "submitted" && coop.Status != "reviewed" {
		return errors.New("kerjasama tidak dalam status diajukan/ditinjau")
	}

	coop.Status = "reviewed"
	if req.Notes != nil {
		coop.Notes = req.Notes
	}

	return s.coopRepo.Update(nil, coop)
}

func (s *mitraCooperationService) ApproveCooperation(cooperationID string, userID uuid.UUID, req dto.ApproveCooperationRequest) error {
	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		return errors.New("kerjasama tidak ditemukan")
	}

	if !s.isRecipientParty(coop, userID) {
		return errors.New("hanya pihak penerima pengajuan yang dapat menyetujui kerjasama")
	}

	if coop.Status != "submitted" && coop.Status != "reviewed" {
		return errors.New("kerjasama tidak dapat disetujui pada status saat ini")
	}

	finalAmount := coop.ProposedAmount
	if req.AgreedAmount != nil && *req.AgreedAmount > 0 {
		finalAmount = *req.AgreedAmount
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	feePct := coop.PlatformFeePercentage
	if feePct == 0 {
		feePct = 11.00
	}
	platformFee := finalAmount * (feePct / 100.0)
	netAmount := finalAmount - platformFee

	// Generate Invoice in DB transaction
	invoice := &models.Invoice{
		MitraCooperationID: &coop.ID,
		FarmerID:           coop.FarmerID,
		Amount:             netAmount,
		PlatformFee:        platformFee,
		TotalAmount:        finalAmount,
		Status:             "pending",
	}

	if err := s.invoiceRepo.Create(tx, invoice); err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat invoice: %w", err)
	}

	coop.AgreedAmount = &finalAmount
	coop.Status = "waiting_payment"
	if req.StartDate != nil {
		coop.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		coop.EndDate = req.EndDate
	}
	if req.Notes != nil {
		coop.Notes = req.Notes
	}

	if err := s.coopRepo.Update(tx, coop); err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal memperbarui status kerjasama: %w", err)
	}

	return tx.Commit().Error
}

func (s *mitraCooperationService) RejectCooperation(cooperationID string, userID uuid.UUID, req dto.RejectCooperationRequest) error {
	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		return errors.New("kerjasama tidak ditemukan")
	}

	if !s.isRecipientParty(coop, userID) {
		return errors.New("hanya pihak penerima pengajuan yang dapat menolak kerjasama")
	}

	if coop.Status == "completed" || coop.Status == "escrowed" || coop.Status == "contract_generated" {
		return errors.New("kerjasama yang sudah dibayar/selesai tidak dapat ditolak")
	}

	coop.Status = "rejected"
	if req.Notes != nil {
		coop.Notes = req.Notes
	}

	return s.coopRepo.Update(nil, coop)
}

func (s *mitraCooperationService) CreateReview(cooperationID string, reviewerID uuid.UUID, reviewerRole string, req dto.CreateMitraReviewRequest) error {
	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		return errors.New("kerjasama tidak ditemukan")
	}

	if coop.Status != "completed" {
		return errors.New("ulasan hanya dapat diberikan setelah kerjasama berstatus 'completed'")
	}

	var reviewerType string
	if reviewerID == coop.FarmerID {
		reviewerType = "farmer"
	} else if reviewerID == coop.MitraID {
		reviewerType = "mitra"
	} else {
		return errors.New("Anda bukan peserta dari kerjasama ini")
	}

	// Check for existing review
	existing, _ := s.reviewRepo.FindByCooperationAndReviewer(cooperationID, reviewerID)
	if existing != nil {
		return errors.New("Anda sudah memberikan ulasan untuk kerjasama ini")
	}

	review := &models.MitraReview{
		CooperationID: coop.ID,
		ReviewerType:  reviewerType,
		ReviewerID:    reviewerID,
		Rating:        req.Rating,
		Comment:       req.Comment,
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.reviewRepo.Create(tx, review); err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat ulasan: %w", err)
	}

	// Recalculate average rating for target party if target is Mitra
	if reviewerType == "farmer" {
		reviews, err := s.reviewRepo.FindAllByCooperationID(cooperationID)
		if err == nil && len(reviews) > 0 {
			var sum int
			var count int
			for _, r := range reviews {
				if r.ReviewerType == "farmer" {
					sum += r.Rating
					count++
				}
			}
			if count > 0 {
				avgRating := float64(sum) / float64(count)
				mitra, err := s.mitraRepo.FindByUserID(coop.MitraID)
				if err == nil && mitra != nil {
					mitra.RatingMitra = avgRating
					_ = s.mitraRepo.Update(tx, mitra)
				}
			}
		}
	}

	return tx.Commit().Error
}

func (s *mitraCooperationService) isRecipientParty(coop *models.MitraCooperation, userID uuid.UUID) bool {
	if coop.InitiatorType == "mitra" {
		// Initiator is Mitra -> Recipient is Farmer
		return coop.FarmerID == userID
	}
	// Initiator is Farmer -> Recipient is Mitra
	return coop.MitraID == userID
}

func (s *mitraCooperationService) toBriefResponse(c *models.MitraCooperation) *dto.CooperationBriefResponse {
	if c == nil {
		return nil
	}

	mitraBrief := dto.CooperationPartyBrief{UserID: c.MitraID}
	if c.Mitra != nil && c.Mitra.User.Name != "" {
		mitraBrief.Name = c.Mitra.NamaMitra
		mitraBrief.Email = c.Mitra.EmailBisnis
		mitraBrief.Phone = &c.Mitra.NomorTeleponBisnis
	}

	farmerBrief := dto.CooperationPartyBrief{UserID: c.FarmerID}
	if c.Farmer != nil {
		farmerBrief.Name = c.Farmer.User.Name
		farmerBrief.Email = c.Farmer.User.Email
		farmerBrief.Phone = c.Farmer.User.PhoneNumber
	}

	return &dto.CooperationBriefResponse{
		ID:            c.ID,
		Title:         c.Title,
		InitiatorType: c.InitiatorType,
		Status:        c.Status,
		ProposedAmount: c.ProposedAmount,
		AgreedAmount:  c.AgreedAmount,
		Mitra:         mitraBrief,
		Farmer:        farmerBrief,
		CreatedAt:     c.CreatedAt,
	}
}

func (s *mitraCooperationService) toDetailResponse(c *models.MitraCooperation) *dto.CooperationDetailResponse {
	if c == nil {
		return nil
	}

	mitraBrief := dto.CooperationPartyBrief{UserID: c.MitraID}
	if c.Mitra != nil {
		mitraBrief.Name = c.Mitra.NamaMitra
		mitraBrief.Email = c.Mitra.EmailBisnis
		mitraBrief.Phone = &c.Mitra.NomorTeleponBisnis
	}

	farmerBrief := dto.CooperationPartyBrief{UserID: c.FarmerID}
	if c.Farmer != nil {
		farmerBrief.Name = c.Farmer.User.Name
		farmerBrief.Email = c.Farmer.User.Email
		farmerBrief.Phone = c.Farmer.User.PhoneNumber
	}

	return &dto.CooperationDetailResponse{
		ID:                     c.ID,
		MitraID:                c.MitraID,
		FarmerID:               c.FarmerID,
		InitiatorType:          c.InitiatorType,
		Title:                  c.Title,
		Description:            c.Description,
		ProposedAmount:         c.ProposedAmount,
		AgreedAmount:           c.AgreedAmount,
		PlatformFeePercentage: c.PlatformFeePercentage,
		StartDate:              c.StartDate,
		EndDate:                c.EndDate,
		Status:                 c.Status,
		Notes:                  c.Notes,
		ContractID:             c.ContractID,
		Mitra:                  mitraBrief,
		Farmer:                 farmerBrief,
		CreatedAt:              c.CreatedAt,
		UpdatedAt:              c.UpdatedAt,
	}
}
