package services

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/xuri/excelize/v2"
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
	GetCombinedTransactions(page, limit int) (*dto.AdminPaginationResponse, error)
	GetAllUsers(page, limit int, search string, roleFilter string) (*dto.AdminPaginationResponse, error)
	GetRevenueAnalytics(startDate, endDate time.Time) (*dto.RevenueAnalyticsResponse, error)
	ExportTransactionsToExcel() (*bytes.Buffer, error)
}

// adminService sekarang menampung repo yang dibutuhkan untuk Dashboard & Payout
type adminService struct {
	payoutRepo           repositories.PayoutRepository
	userRepo             repositories.UserRepository
	userVerificationRepo repositories.UserVerificationRepository
	transactionRepo      repositories.TransactionRepository
	projectRepo          repositories.ProjectRepository
	ecommPaymentRepo repositories.ECommercePaymentRepository
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
	ecommPaymentRepo repositories.ECommercePaymentRepository,
	orderRepo repositories.OrderRepository,
	db *gorm.DB,
) AdminService {
	return &adminService{
		payoutRepo:           payoutRepo,
		userRepo:             userRepo,
		userVerificationRepo: userVerificationRepo,
		transactionRepo:      transactionRepo,
		projectRepo:          projectRepo,
		ecommPaymentRepo: ecommPaymentRepo,
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


func (s *adminService) GetCombinedTransactions(page, limit int) (*dto.AdminPaginationResponse, error) {
	// 1. Ambil Transaksi Jasa (Project/Delivery)
	serviceTransactions, totalService, err := s.transactionRepo.GetAllTransactions(page, limit)
	if err != nil {
		return nil, err
	}

	// 2. Ambil Transaksi Produk (E-commerce)
	productPayments, totalProduct, err := s.ecommPaymentRepo.GetAllPayments(page, limit)
	if err != nil {
		return nil, err
	}

	var combinedList []dto.TransactionDetailResponse

	// 3. Mapping Transaksi Jasa ke DTO
	for _, t := range serviceTransactions {
		detail := dto.TransactionDetailResponse{
			TransactionID:   t.ID.String(),
			TransactionDate: t.TransactionDate, // Asumsi field ini ada di model Transaction
			AmountPaid:      t.AmountPaid,
			TransactionType: "Jasa",
		}
		
		// Handling pointer nil untuk PaymentMethod
		if t.PaymentMethod != nil {
			detail.PaymentMethod = *t.PaymentMethod
		}

		// Isi Nama Payer & Context
		if t.Invoice.FarmerID != uuid.Nil {
             // Note: Pastikan repository mem-preload Invoice.Farmer.User
             // Jika Farmer nil, cek logika preload di repo
             if t.Invoice.Farmer != nil {
                  detail.PayerName = t.Invoice.Farmer.User.Name
             }
		}

		if t.Invoice.ProjectID != nil && t.Invoice.Project != nil {
			detail.ContextInfo = "Proyek: " + t.Invoice.Project.Title
		} else if t.Invoice.DeliveryID != nil && t.Invoice.Delivery != nil {
			detail.ContextInfo = "Pengiriman: " + t.Invoice.Delivery.ItemDescription
		}

		combinedList = append(combinedList, detail)
	}

	// 4. Mapping Transaksi Produk ke DTO
	for _, p := range productPayments {
		detail := dto.TransactionDetailResponse{
			TransactionID:   p.ID.String(),
			TransactionDate: p.CreatedAt, // Gunakan CreatedAt sebagai tanggal transaksi
			AmountPaid:      p.GrandTotal,
			PaymentMethod:   "midtrans", // Default atau ambil dari response midtrans jika disimpan
			TransactionType: "Produk",
			PayerName:       p.User.Name,
			ContextInfo:     "Pesanan E-commerce", // Bisa diperdetail jika perlu
		}
		combinedList = append(combinedList, detail)
	}

	// 5. Sorting Gabungan (Terbaru di atas)
	sort.Slice(combinedList, func(i, j int) bool {
		return combinedList[i].TransactionDate.After(combinedList[j].TransactionDate)
	})

	// Hitung total gabungan
	totalCombined := totalService + totalProduct
	
	// Hitung total pages (perkiraan kasar karena kita merge 2 limit)
	// Agar pagination UI tidak bingung, kita gunakan total item gabungan
	// dibagi limit yang diminta.
	totalPages := int(totalCombined) / limit
	if int(totalCombined)%limit != 0 {
		totalPages++
	}

	return &dto.AdminPaginationResponse{
		Data:       combinedList,
		TotalItems: totalCombined,
		TotalPages: totalPages,
		CurrentPage: page,
	}, nil
}

func (s *adminService) GetAllUsers(page, limit int, search string, roleFilter string) (*dto.AdminPaginationResponse, error) {
	users, total, err := s.userRepo.FindAllUsers(page, limit, search, roleFilter)
	if err != nil {
		return nil, err
	}

	var userResponses []dto.UserDetailResponse
	for _, u := range users {
		userResponses = append(userResponses, dto.UserDetailResponse{
			ID:            u.ID,
			Name:          u.Name,
			Email:         u.Email,
			PhoneNumber:   u.PhoneNumber,
			Role:          u.Role,
			IsActive:      u.IsActive,
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt,
		})
	}

	// Hitung total halaman
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.AdminPaginationResponse{
		Data:        userResponses,
		TotalItems:  total,
		TotalPages:  totalPages,
		CurrentPage: page,
	}, nil
}

func (s *adminService) GetRevenueAnalytics(startDate, endDate time.Time) (*dto.RevenueAnalyticsResponse, error) {
	// 1. Ambil Data Jasa
	svcTotal, svcTrend, err := s.transactionRepo.GetRevenueStats(startDate, endDate)
	if err != nil { return nil, err }

	// 2. Ambil Data Produk
	prodTotal, prodTrend, err := s.ecommPaymentRepo.GetRevenueStats(startDate, endDate)
	if err != nil { return nil, err }

	// 3. Gabungkan Tren Harian (Merging logic)
	// Gunakan map untuk menjumlahkan value pada tanggal yang sama
	trendMap := make(map[string]float64)

	for _, t := range svcTrend {
		// Asumsi format date dari DB adalah "YYYY-MM-DD" atau time.Time string
		// Kita ambil substring 10 karakter pertama (YYYY-MM-DD) untuk aman
		dateStr := t.Date
		if len(dateStr) > 10 { dateStr = dateStr[:10] }
		trendMap[dateStr] += t.Value
	}
	for _, t := range prodTrend {
		dateStr := t.Date
		if len(dateStr) > 10 { dateStr = dateStr[:10] }
		trendMap[dateStr] += t.Value
	}

	// Konversi map kembali ke slice untuk response
	var combinedTrend []dto.DailyDataPoint
	for date, value := range trendMap {
		combinedTrend = append(combinedTrend, dto.DailyDataPoint{
			Date:  date,
			Value: value,
		})
	}

	// Urutkan berdasarkan tanggal (karena map mengacak urutan)
	// (Anda perlu import "sort")
	sort.Slice(combinedTrend, func(i, j int) bool {
		return combinedTrend[i].Date < combinedTrend[j].Date
	})

	return &dto.RevenueAnalyticsResponse{
		TotalRevenue:     svcTotal + prodTotal,
		RevenueByService: svcTotal,
		RevenueByProduct: prodTotal,
		DailyTrend:       combinedTrend,
	}, nil
}

func (s *adminService) ExportTransactionsToExcel() (*bytes.Buffer, error) {
    // 1. Ambil SEMUA data dari kedua sumber
    svcTrx, err := s.transactionRepo.GetAllTransactionsNoPaging()
    if err != nil { return nil, err }
    
    prodTrx, err := s.ecommPaymentRepo.GetAllPaymentsNoPaging()
    if err != nil { return nil, err }

    // 2. Gabungkan Data (Logic sama seperti GetCombinedTransactions tapi tanpa paginasi)
    var combinedList []dto.TransactionDetailResponse

	// Mapping Jasa
	for _, t := range svcTrx {
		item := dto.TransactionDetailResponse{
			TransactionID:   t.ID.String(),
			TransactionDate: t.TransactionDate,
			AmountPaid:      t.AmountPaid,
			TransactionType: "Jasa",
		}
		if t.PaymentMethod != nil { item.PaymentMethod = *t.PaymentMethod }
		if t.Invoice.Farmer != nil { item.PayerName = t.Invoice.Farmer.User.Name }
		
		if t.Invoice.ProjectID != nil && t.Invoice.Project != nil {
			item.ContextInfo = "Proyek: " + t.Invoice.Project.Title
		} else if t.Invoice.DeliveryID != nil && t.Invoice.Delivery != nil {
			item.ContextInfo = "Pengiriman: " + t.Invoice.Delivery.ItemDescription
		}
		combinedList = append(combinedList, item)
	}

    // Mapping Produk
    for _, p := range prodTrx {
        item := dto.TransactionDetailResponse{
            TransactionID:   p.ID.String(),
            TransactionDate: p.CreatedAt,
            AmountPaid:      p.GrandTotal,
            TransactionType: "Produk",
            PaymentMethod:   "Midtrans",
            PayerName:       p.User.Name,
            ContextInfo:     "Pesanan E-commerce",
        }
        combinedList = append(combinedList, item)
    }

    // Sorting (Terbaru di atas)
    sort.Slice(combinedList, func(i, j int) bool {
        return combinedList[i].TransactionDate.After(combinedList[j].TransactionDate)
    })

    // 3. BUAT FILE EXCEL
    f := excelize.NewFile()
    sheetName := "Transactions"
    index, _ := f.NewSheet(sheetName)
    f.SetActiveSheet(index)
    f.DeleteSheet("Sheet1") // Hapus sheet default

    // Header Style
    style, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
    })

    // Tulis Header
    headers := []string{"Tanggal", "ID Transaksi", "Tipe", "Konteks", "Pembayar", "Metode", "Jumlah (Rp)", "Status"}
    for i, header := range headers {
        cell := fmt.Sprintf("%c1", 'A'+i) // A1, B1, C1...
        f.SetCellValue(sheetName, cell, header)
        f.SetCellStyle(sheetName, cell, cell, style)
    }

    // Tulis Data
    for i, row := range combinedList {
        rowNum := i + 2
        f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), row.TransactionDate.Format("2006-01-02 15:04"))
        f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row.TransactionID)
        f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row.TransactionType)
        f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row.ContextInfo)
        f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row.PayerName)
        f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row.PaymentMethod)
        f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), row.AmountPaid)
    }

    // Atur lebar kolom otomatis (opsional, manual lebih aman)
    f.SetColWidth(sheetName, "A", "A", 20)
    f.SetColWidth(sheetName, "B", "B", 36) // UUID panjang
    f.SetColWidth(sheetName, "D", "D", 30)

    // Tulis ke Buffer
    buffer, err := f.WriteToBuffer()
    if err != nil {
        return nil, err
    }

    return buffer, nil
}