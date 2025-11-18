package services

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

// AdminService mendefinisikan semua logika bisnis untuk panel admin.
type AdminService interface {
	// Fitur Dashboard
	GetDashboardStats() (*dto.AdminDashboardResponse, error)
	// Fitur Payout
	GetPendingPayouts() ([]dto.PayoutDetailResponse, error)
	MarkPayoutAsCompleted(payoutID string, adminID uuid.UUID, transferProofURL string) error
	GetPendingVerifications() ([]models.UserVerification, error)
	ReviewVerification(verificationID uuid.UUID, input dto.ReviewVerificationInput, adminID uuid.UUID) error
}

// adminService sekarang menampung repo yang dibutuhkan untuk Dashboard & Payout
type adminService struct {
	payoutRepo           repositories.PayoutRepository
	userRepo             repositories.UserRepository
	userVerificationRepo repositories.UserVerificationRepository
	transactionRepo      repositories.TransactionRepository
	projectRepo          repositories.ProjectRepository
	deliveryRepo         repositories.DeliveryRepository
	orderRepo            repositories.OrderRepository
	db                   *gorm.DB
}

// NewAdminService sekarang menerima dependensi yang relevan
func NewAdminService(
	payoutRepo repositories.PayoutRepository,
	userRepo repositories.UserRepository,
	userVerificationRepo repositories.UserVerificationRepository,
	transactionRepo repositories.TransactionRepository,
	projectRepo repositories.ProjectRepository,
	deliveryRepo repositories.DeliveryRepository,
	orderRepo repositories.OrderRepository,
	db *gorm.DB,
) AdminService {
	return &adminService{
		payoutRepo:           payoutRepo,
		userRepo:             userRepo,
		userVerificationRepo: userVerificationRepo,
		transactionRepo:      transactionRepo,
		projectRepo:          projectRepo,
		deliveryRepo:         deliveryRepo,
		orderRepo:            orderRepo,
		db:                   db,
	}
}

// GetDashboardStats mengumpulkan semua data untuk halaman dashboard admin.
func (s *adminService) GetDashboardStats() (*dto.AdminDashboardResponse, error) {
	thirtyDaysAgo := time.Now().AddDate(0, -1, 0)

	// 1. Ambil Data KPI
	pendingPayoutCount, pendingPayoutTotal, _ := s.payoutRepo.GetPendingPayoutStats()
	newUsersCount, _ := s.userRepo.CountNewUsers(thirtyDaysAgo)
	totalRevenueMonthly, _ := s.transactionRepo.GetTotalRevenue(thirtyDaysAgo)
	activeProjects, _ := s.projectRepo.CountActiveProjects()
	activeDeliveries, _ := s.deliveryRepo.CountActiveDeliveries()
	newECommerceOrders, _ := s.orderRepo.CountNewOrders(thirtyDaysAgo)

	kpis := dto.DashboardKPIs{
		TotalRevenueMonthly: totalRevenueMonthly,
		PendingPayoutsTotal: pendingPayoutTotal,
		NewUsersMonthly:     int(newUsersCount),
		ActiveProjects:      int(activeProjects),
		ActiveDeliveries:    int(activeDeliveries),
		NewECommerceOrders:  int(newECommerceOrders),
	}

	// 2. Ambil Data Antrean "Butuh Tindakan"
	// openDisputes, _ := s.disputeRepo.CountOpen() // Perlu repo sengketa

	actionQueue := dto.DashboardActionQueue{
		PendingPayouts: int(pendingPayoutCount),
		OpenDisputes:   0, // Ganti dengan data asli nanti
	}

	// 3. Ambil Data Grafik
	revenueTrend, _ := s.transactionRepo.GetDailyRevenueTrend(thirtyDaysAgo)
	userTrend, _ := s.userRepo.GetDailyUserTrend(thirtyDaysAgo)

	// 4. Susun Respons Final
	response := &dto.AdminDashboardResponse{
		KPIs:         kpis,
		ActionQueue:  actionQueue,
		RevenueTrend: revenueTrend,
		UserTrend:    userTrend,
	}

	return response, nil
}

// --- Implementasi Fungsi Payout ---

// GetPendingPayouts mengambil daftar payout yang perlu dibayar oleh admin.
func (s *adminService) GetPendingPayouts() ([]dto.PayoutDetailResponse, error) {
	// 1. Ambil Payouts (tanpa data Worker/Driver)
	payouts, err := s.payoutRepo.FindPendingPayouts()
	if err != nil {
		return nil, err
	}

	var response []dto.PayoutDetailResponse
	for _, p := range payouts {
		// 2. [BARU] Muat data Worker/Driver secara manual
		// Kita gunakan 's.db' (koneksi DB utama) untuk memuat
		if err := p.LoadPayee(s.db); err != nil {
			log.Printf("Warning: Failed to load payee %s: %v", p.PayeeID, err)
			continue // Lewati jika payee tidak ditemukan
		}

		// 3. Buat DTO (Logika Anda sebelumnya sekarang berfungsi)
		dto := dto.PayoutDetailResponse{
			PayoutID:   p.ID,
			Amount:     p.Amount,
			ReleasedAt: p.ReleasedAt,
			PayeeID:    p.PayeeID,
			PayeeType:  p.PayeeType,
		}

		if p.PayeeType == "worker" && p.Worker != nil {
			dto.PayeeName = p.Worker.User.Name
			if p.Worker.BankName != nil {
				dto.BankName = *p.Worker.BankName
				dto.BankAccountNumber = *p.Worker.BankAccountNumber
				dto.BankAccountHolder = *p.Worker.BankAccountHolder
				// ... (info bank lain)
			}
			if p.Transaction.Invoice.Project != nil {
				dto.ContextTitle = p.Transaction.Invoice.Project.Title
			}
		} else if p.PayeeType == "driver" && p.Driver != nil {
			dto.PayeeName = p.Driver.User.Name
			if p.Driver.BankName != nil {
				dto.BankName = *p.Driver.BankName
				dto.BankAccountNumber = *p.Driver.BankAccountNumber
				dto.BankAccountHolder = *p.Driver.BankAccountHolder
				// ... (info bank lain)
			}
			if p.Transaction.Invoice.DeliveryID != nil {
				dto.ContextTitle = "Pengiriman: " + p.Transaction.Invoice.Delivery.ItemDescription
			}
		}
		response = append(response, dto)
	}
	return response, nil
}

// MarkPayoutAsCompleted menandai payout sebagai 'completed' oleh admin.
func (s *adminService) MarkPayoutAsCompleted(payoutID string, adminID uuid.UUID, transferProofURL string) error {
	payout, err := s.payoutRepo.FindByID(payoutID)
	if err != nil {
		return errors.New("payout not found")
	}
	if payout.Status != "pending_disbursement" {
		return errors.New("payout has already been processed")
	}

	tx := s.db.Begin()

	// Update status DAN URL bukti transfer
	payout.Status = "completed"
	payout.TransferProofURL = &transferProofURL // Simpan URL yang sudah jadi

	if err := s.payoutRepo.Update(tx, payout); err != nil {
		tx.Rollback()
		return err
	}

	// TODO: Log aksi admin

	return tx.Commit().Error
}

func (s *adminService) GetPendingVerifications() ([]models.UserVerification, error) {
	// Panggil repo untuk mengambil data, Preload User agar bisa tampilkan nama
	return s.userVerificationRepo.FindPending()
}

// ReviewVerification memproses keputusan admin (Setuju/Tolak).
func (s *adminService) ReviewVerification(verificationID uuid.UUID, input dto.ReviewVerificationInput, adminID uuid.UUID) error {
	// 1. Ambil data verifikasi
	verification, err := s.userVerificationRepo.FindByID(verificationID)
	if err != nil {
		return errors.New("verification request not found")
	}

	// 2. Cek apakah sudah diproses
	if verification.Status != "pending" {
		return errors.New("this document has already been reviewed")
	}

	// 3. Update data
	verification.Status = input.Status
	verification.Notes = &input.Notes
	verification.ReviewedBy = &adminID // Catat siapa admin yang mereview

	tx := s.db.Begin()
	if err := s.userVerificationRepo.UpdateStatus(tx, verification); err != nil {
		tx.Rollback()
		return err
	}

	// TODO: Anda bisa menambahkan logika di sini untuk memeriksa
	// apakah pengguna sekarang sudah "fully verified" dan meng-update
	// status di tabel 'users' jika perlu.

	return tx.Commit().Error
}
