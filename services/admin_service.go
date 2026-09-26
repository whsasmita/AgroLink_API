package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
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
	GetCombinedTransactions(page, limit int, search, status, layanan, sektor, startDate, endDate string) (*dto.AdminPaginationResponse, error)
	GetAllUsers(page, limit int, search string, roleFilter string) (*dto.AdminPaginationResponse, error)
	GetRevenueAnalytics(startDate, endDate time.Time) (*dto.RevenueAnalyticsResponse, error)
	ExportTransactionsToExcel(search, status, layanan, sektor, startDate, endDate string) (*bytes.Buffer, error)
}

// adminService sekarang menampung repo yang dibutuhkan untuk Dashboard & Payout
type adminService struct {
	payoutRepo           repositories.PayoutRepository
	userRepo             repositories.UserRepository
	userVerificationRepo repositories.UserVerificationRepository
	transactionRepo      repositories.TransactionRepository
	projectRepo          repositories.ProjectRepository
	ecommPaymentRepo     repositories.ECommercePaymentRepository
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
		deliveryRepo:         deliveryRepo,
		ecommPaymentRepo:     ecommPaymentRepo,
		orderRepo:            orderRepo,
		db:                   db,
	}
}

type seedItem struct {
	IDTransaksi        string   `json:"IDTransaksi"`
	Tanggal            string   `json:"Tanggal"`
	Timestamp          string   `json:"Timestamp"`
	BulanTahun         string   `json:"Bulan_Tahun"`
	Layanan            string   `json:"Layanan"`
	StatusTransaksi    string   `json:"StatusTransaksi"`
	Keterangan         string   `json:"Keterangan"`
	KomentarUser       string   `json:"KomentarUser"`
	MetodePembayaran   string   `json:"MetodePembayaran"`
	NominalTransaksi   *float64 `json:"NominalTransaksi"`
	PersentaseKomisi   *float64 `json:"PersentaseKomisi"`
	KeuntunganKotor    *float64 `json:"KeuntunganKotor"`
	BiayaMidtrans      *float64 `json:"BiayaMidtrans"`
	KeuntunganBersih   *float64 `json:"KeuntunganBersih"`
	TotalDiterimaMitra *float64 `json:"TotalDiterimaMitra"`

	FarmerEmail       *string `json:"FarmerEmail"`
	FarmerName        *string `json:"FarmerName"`
	WorkerEmail       *string `json:"WorkerEmail"`
	WorkerName        *string `json:"WorkerName"`
	DriverEmail       *string `json:"DriverEmail"`
	DriverName        *string `json:"DriverName"`
	PenjualEmail      *string `json:"PenjualEmail"`
	PembeliEmail      *string `json:"PembeliEmail"`
	BuyerEmail        *string `json:"BuyerEmail"`
	BuyerName         *string `json:"BuyerName"`
	MitraEmail        *string `json:"MitraEmail"`
	MitraName         *string `json:"MitraName"`
	UserEmail         *string `json:"UserEmail"`
	PemberiKerjaEmail *string `json:"PemberiKerjaEmail"`
	PekerjaEmail      *string `json:"PekerjaEmail"`
}

func loadSeedTransactions() ([]seedItem, error) {
	paths := []string{"seeders/new_transaction.json", "../seeders/new_transaction.json", "./new_transaction.json"}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}

	var rows []seedItem
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seed json: %w", err)
	}
	return rows, nil
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

	// 2. Ambil Statistik Pengguna Lengkap dari DB (TotalUsers mengecualikan admin)
	var userStats dto.DashboardUserStats
	var userCounts []struct {
		Role  string
		Count int64
	}
	s.db.Model(&models.User{}).Select("role, count(*) as count").Group("role").Scan(&userCounts)
	for _, uc := range userCounts {
		switch uc.Role {
		case "worker":
			userStats.TotalWorker = uc.Count
			userStats.TotalUsers += uc.Count
		case "farmer":
			userStats.TotalFarmer = uc.Count
			userStats.TotalUsers += uc.Count
		case "driver":
			userStats.TotalDriver = uc.Count
			userStats.TotalUsers += uc.Count
		case "general":
			userStats.TotalGeneral = uc.Count
			userStats.TotalUsers += uc.Count
		case "mitra":
			userStats.TotalMitra = uc.Count
			userStats.TotalUsers += uc.Count
		case "admin":
			userStats.TotalAdmin = uc.Count
		}
	}

	// 3. Ambil Breakdown Layanan, Sektor & Financial Summary dari Seed JSON
	serviceOrder := []string{"Pekerja", "Ekspedisi", "E-Commerce", "Chatbot Premium", "Tukang", "Peternak", "Kemitraan"}
	serviceMap := make(map[string]*dto.ServiceSummary)
	for _, name := range serviceOrder {
		sektor := "-"
		switch name {
		case "Pekerja":
			sektor = "Pertanian"
		case "Peternak":
			sektor = "Peternakan"
		case "Tukang":
			sektor = "Tukang Bangunan"
		}
		serviceMap[name] = &dto.ServiceSummary{
			Name:   name,
			Sector: sektor,
		}
	}

	workerSectorMap := map[string]*dto.WorkerSectorSummary{
		"Pertanian":       {Sector: "Pertanian"},
		"Peternakan":      {Sector: "Peternakan"},
		"Tukang Bangunan": {Sector: "Tukang Bangunan"},
	}

	months := []string{
		"2025-09", "2025-10", "2025-11", "2025-12",
		"2026-01", "2026-02", "2026-03", "2026-04", "2026-05",
		"2026-06", "2026-07", "2026-08", "2026-09",
	}
	monthLabels := map[string]string{
		"2025-09": "Sep 2025", "2025-10": "Okt 2025", "2025-11": "Nov 2025", "2025-12": "Des 2025",
		"2026-01": "Jan 2026", "2026-02": "Feb 2026", "2026-03": "Mar 2026", "2026-04": "Apr 2026",
		"2026-05": "Mei 2026", "2026-06": "Jun 2026", "2026-07": "Jul 2026", "2026-08": "Agu 2026",
		"2026-09": "Sep 2026",
	}

	monthlyMap := make(map[string]*dto.MonthlyProfitPoint)
	for _, m := range months {
		monthlyMap[m] = &dto.MonthlyProfitPoint{
			Month:      m,
			MonthLabel: monthLabels[m],
		}
	}

	var financialSummary dto.DashboardFinancialSummary
	var trendMap = make(map[string]float64)

	rows, err := loadSeedTransactions()
	if err == nil {
		for _, r := range rows {
			nom := 0.0
			if r.NominalTransaksi != nil {
				nom = *r.NominalTransaksi
			}
			kotor := 0.0
			if r.KeuntunganKotor != nil {
				kotor = *r.KeuntunganKotor
			}
			fee := 0.0
			if r.BiayaMidtrans != nil {
				fee = *r.BiayaMidtrans
			}
			bersih := 0.0
			if r.KeuntunganBersih != nil {
				bersih = *r.KeuntunganBersih
			}
			mitra := 0.0
			if r.TotalDiterimaMitra != nil {
				mitra = *r.TotalDiterimaMitra
			}

			isSuccess := strings.EqualFold(strings.TrimSpace(r.StatusTransaksi), "Sukses")

			financialSummary.TotalTransactions++
			if isSuccess {
				financialSummary.SuccessfulTransactions++
			} else {
				financialSummary.FailedTransactions++
			}
			financialSummary.TotalGMV += nom
			financialSummary.TotalGrossProfit += kotor
			financialSummary.TotalGatewayFee += fee
			financialSummary.TotalNetProfit += bersih
			financialSummary.TotalMitraShare += mitra

			if r.Tanggal <= "2026-05-31" {
				financialSummary.Phase1Transactions++
				if isSuccess {
					financialSummary.Phase1Success++
				} else {
					financialSummary.Phase1Failed++
				}
				financialSummary.Phase1GMV += nom
				financialSummary.Phase1GrossProfit += kotor
				financialSummary.Phase1GatewayFee += fee
				financialSummary.Phase1NetProfit += bersih
			} else {
				financialSummary.Phase2Transactions++
				if isSuccess {
					financialSummary.Phase2Success++
				} else {
					financialSummary.Phase2Failed++
				}
				financialSummary.Phase2GMV += nom
				financialSummary.Phase2GrossProfit += kotor
				financialSummary.Phase2GatewayFee += fee
				financialSummary.Phase2NetProfit += bersih
			}

			// Update Service Summary
			if sm, exists := serviceMap[r.Layanan]; exists {
				sm.TransactionCount++
				if isSuccess {
					sm.SuccessCount++
				} else {
					sm.FailedCount++
				}
				sm.TotalAmount += nom
				sm.GrossProfit += kotor
				sm.GatewayFee += fee
				sm.NetProfit += bersih
				sm.TotalMitraShare += mitra
			}

			// Update Worker Sector Summary
			var workerSector string
			switch r.Layanan {
			case "Pekerja":
				workerSector = "Pertanian"
			case "Peternak":
				workerSector = "Peternakan"
			case "Tukang":
				workerSector = "Tukang Bangunan"
			}
			if workerSector != "" {
				if ws, exists := workerSectorMap[workerSector]; exists {
					ws.TotalCount++
					if isSuccess {
						ws.SuccessCount++
					} else {
						ws.FailedCount++
					}
					ws.TotalAmount += nom
					ws.GrossProfit += kotor
					ws.GatewayFee += fee
					ws.NetProfit += bersih
					ws.TotalMitraShare += mitra
				}
			}

			// Update Monthly Profit Point
			if len(r.Tanggal) >= 7 {
				mKey := r.Tanggal[:7]
				if mp, ok := monthlyMap[mKey]; ok {
					mp.TotalCount++
					if isSuccess {
						mp.SuccessCount++
					} else {
						mp.FailedCount++
					}
					mp.GMV += nom
					mp.GrossProfit += kotor
					mp.GatewayFee += fee
					mp.NetProfit += bersih
				}
			}

			if len(r.Tanggal) >= 10 {
				trendMap[r.Tanggal[:10]] += nom
			}
		}
	}

	var serviceSummaries []dto.ServiceSummary
	for _, name := range serviceOrder {
		if sm, ok := serviceMap[name]; ok {
			if financialSummary.TotalGMV > 0 {
				sm.Percentage = (sm.TotalAmount / financialSummary.TotalGMV) * 100.0
			}
			serviceSummaries = append(serviceSummaries, *sm)
		}
	}

	var workerSectorSummaries []dto.WorkerSectorSummary
	for _, sec := range []string{"Pertanian", "Peternakan", "Tukang Bangunan"} {
		if ws, ok := workerSectorMap[sec]; ok {
			workerSectorSummaries = append(workerSectorSummaries, *ws)
		}
	}

	// Calculate Cumulative Net Profit across months
	var monthlyProfitTrend []dto.MonthlyProfitPoint
	runningCumulativeNet := 0.0
	for _, m := range months {
		if mp, ok := monthlyMap[m]; ok {
			runningCumulativeNet += mp.NetProfit
			mp.CumulativeNetProfit = runningCumulativeNet
			monthlyProfitTrend = append(monthlyProfitTrend, *mp)
		}
	}

	// Status breakdown
	var statusBreakdown []dto.TransactionStatusBreakdown
	for _, sm := range serviceSummaries {
		statusBreakdown = append(statusBreakdown, dto.TransactionStatusBreakdown{
			Layanan:     sm.Name,
			Sektor:      sm.Sector,
			Sukses:      sm.SuccessCount,
			Gagal:       sm.FailedCount,
			Total:       sm.TransactionCount,
			GrossProfit: sm.GrossProfit,
			NetProfit:   sm.NetProfit,
		})
	}

	// 4. Data Antrean Tindakan
	actionQueue := dto.DashboardActionQueue{
		PendingPayouts: int(pendingPayoutCount),
		OpenDisputes:   0,
	}

	// 5. Data Grafik Revenue Trend
	var revenueTrend []dto.DailyDataPoint
	for date, val := range trendMap {
		revenueTrend = append(revenueTrend, dto.DailyDataPoint{
			Date:  date,
			Value: val,
		})
	}
	sort.Slice(revenueTrend, func(i, j int) bool {
		return revenueTrend[i].Date < revenueTrend[j].Date
	})

	userTrend, _ := s.userRepo.GetDailyUserTrend(time.Now().AddDate(-1, 0, 0))

	// 6. Susun Respons Final
	response := &dto.AdminDashboardResponse{
		KPIs:                  kpis,
		FinancialSummary:      financialSummary,
		ServiceSummaries:      serviceSummaries,
		WorkerSectorSummaries: workerSectorSummaries,
		MonthlyProfitTrend:    monthlyProfitTrend,
		StatusBreakdown:       statusBreakdown,
		UserStats:             userStats,
		ActionQueue:           actionQueue,
		RevenueTrend:          revenueTrend,
		UserTrend:             userTrend,
	}

	return response, nil
}

func (s *adminService) GetCombinedTransactions(page, limit int, search, status, layanan, sektor, startDate, endDate string) (*dto.AdminPaginationResponse, error) {
	rows, err := loadSeedTransactions()
	if err != nil {
		return nil, err
	}

	var combinedList []dto.TransactionDetailResponse

	for _, item := range rows {
		nom := 0.0
		if item.NominalTransaksi != nil {
			nom = *item.NominalTransaksi
		}
		komisi := 0.0
		if item.PersentaseKomisi != nil {
			komisi = *item.PersentaseKomisi
		}
		kotor := 0.0
		if item.KeuntunganKotor != nil {
			kotor = *item.KeuntunganKotor
		}
		fee := 0.0
		if item.BiayaMidtrans != nil {
			fee = *item.BiayaMidtrans
		}
		bersih := 0.0
		if item.KeuntunganBersih != nil {
			bersih = *item.KeuntunganBersih
		}
		mitra := 0.0
		if item.TotalDiterimaMitra != nil {
			mitra = *item.TotalDiterimaMitra
		}

		sektorVal := "-"
		switch item.Layanan {
		case "Pekerja":
			sektorVal = "Pertanian"
		case "Peternak":
			sektorVal = "Peternakan"
		case "Tukang":
			sektorVal = "Tukang Bangunan"
		}

		statusTrx := item.StatusTransaksi
		if statusTrx == "" {
			statusTrx = "Sukses"
		}

		statusCode := "paid"
		if strings.EqualFold(statusTrx, "Gagal") {
			statusCode = "failed"
		}

		var txnDate time.Time
		if strings.TrimSpace(item.Timestamp) != "" {
			txnDate, _ = time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(item.Timestamp), time.Local)
		}
		if txnDate.IsZero() && strings.TrimSpace(item.Tanggal) != "" {
			txnDate, _ = time.ParseInLocation("2006-01-02", strings.TrimSpace(item.Tanggal), time.Local)
		}

		payerName := "Pengguna AgroLink"
		if item.FarmerName != nil && *item.FarmerName != "" {
			payerName = *item.FarmerName
		} else if item.BuyerName != nil && *item.BuyerName != "" {
			payerName = *item.BuyerName
		} else if item.MitraName != nil && *item.MitraName != "" {
			payerName = *item.MitraName
		} else if item.FarmerEmail != nil && *item.FarmerEmail != "" {
			payerName = *item.FarmerEmail
		} else if item.BuyerEmail != nil && *item.BuyerEmail != "" {
			payerName = *item.BuyerEmail
		} else if item.UserEmail != nil && *item.UserEmail != "" {
			payerName = *item.UserEmail
		}

		trxType := "Jasa"
		if item.Layanan == "E-Commerce" {
			trxType = "Produk"
		}

		// Apply Filters
		// 1. Status filter
		if status != "" && !strings.EqualFold(status, "all") {
			stLower := strings.ToLower(strings.TrimSpace(status))
			if stLower == "sukses" || stLower == "paid" || stLower == "success" || stLower == "berhasil" {
				if !strings.EqualFold(statusTrx, "Sukses") {
					continue
				}
			} else if stLower == "gagal" || stLower == "failed" || stLower == "error" {
				if !strings.EqualFold(statusTrx, "Gagal") {
					continue
				}
			}
		}

		// 2. Layanan filter
		if layanan != "" && !strings.EqualFold(layanan, "all") {
			layLower := strings.ToLower(strings.TrimSpace(layanan))
			itemLayLower := strings.ToLower(item.Layanan)
			if layLower == "pekerja" || layLower == "pekerja_tani" || layLower == "tani" {
				if itemLayLower != "pekerja" {
					continue
				}
			} else if layLower == "peternak" || layLower == "pekerja_ternak" || layLower == "ternak" {
				if itemLayLower != "peternak" {
					continue
				}
			} else if layLower == "tukang" || layLower == "pekerja_tukang" {
				if itemLayLower != "tukang" {
					continue
				}
			} else if layLower == "ekspedisi" {
				if itemLayLower != "ekspedisi" {
					continue
				}
			} else if layLower == "ecommerce" || layLower == "e-commerce" || layLower == "produk" {
				if itemLayLower != "e-commerce" {
					continue
				}
			} else if layLower == "chatbot" || layLower == "chatbot premium" || layLower == "ai" {
				if itemLayLower != "chatbot premium" {
					continue
				}
			} else if layLower == "kemitraan" || layLower == "mitra" || layLower == "b2b" {
				if itemLayLower != "kemitraan" {
					continue
				}
			} else {
				if !strings.Contains(itemLayLower, layLower) {
					continue
				}
			}
		}

		// 3. Sektor filter
		if sektor != "" && !strings.EqualFold(sektor, "all") {
			sekLower := strings.ToLower(strings.TrimSpace(sektor))
			itemSekLower := strings.ToLower(sektorVal)
			if !strings.Contains(itemSekLower, sekLower) {
				continue
			}
		}

		// 4. Date Range filter
		if startDate != "" && strings.TrimSpace(item.Tanggal) < startDate {
			continue
		}
		if endDate != "" && strings.TrimSpace(item.Tanggal) > endDate {
			continue
		}

		// 5. Search query
		if search != "" {
			query := strings.ToLower(strings.TrimSpace(search))
			matched := strings.Contains(strings.ToLower(item.IDTransaksi), query) ||
				strings.Contains(strings.ToLower(item.Keterangan), query) ||
				strings.Contains(strings.ToLower(item.KomentarUser), query) ||
				strings.Contains(strings.ToLower(item.MetodePembayaran), query) ||
				strings.Contains(strings.ToLower(payerName), query)

			if !matched && item.FarmerEmail != nil && strings.Contains(strings.ToLower(*item.FarmerEmail), query) {
				matched = true
			}
			if !matched && item.FarmerName != nil && strings.Contains(strings.ToLower(*item.FarmerName), query) {
				matched = true
			}
			if !matched && item.WorkerEmail != nil && strings.Contains(strings.ToLower(*item.WorkerEmail), query) {
				matched = true
			}
			if !matched && item.WorkerName != nil && strings.Contains(strings.ToLower(*item.WorkerName), query) {
				matched = true
			}
			if !matched && item.DriverEmail != nil && strings.Contains(strings.ToLower(*item.DriverEmail), query) {
				matched = true
			}
			if !matched && item.DriverName != nil && strings.Contains(strings.ToLower(*item.DriverName), query) {
				matched = true
			}
			if !matched && item.BuyerEmail != nil && strings.Contains(strings.ToLower(*item.BuyerEmail), query) {
				matched = true
			}
			if !matched && item.BuyerName != nil && strings.Contains(strings.ToLower(*item.BuyerName), query) {
				matched = true
			}
			if !matched && item.MitraEmail != nil && strings.Contains(strings.ToLower(*item.MitraEmail), query) {
				matched = true
			}
			if !matched && item.MitraName != nil && strings.Contains(strings.ToLower(*item.MitraName), query) {
				matched = true
			}

			if !matched {
				continue
			}
		}

		combinedList = append(combinedList, dto.TransactionDetailResponse{
			TransactionID:      item.IDTransaksi,
			TransactionDate:    txnDate,
			Layanan:            item.Layanan,
			Sektor:             sektorVal,
			StatusTransaksi:    statusTrx,
			Status:             statusCode,
			Keterangan:         item.Keterangan,
			KomentarUser:       item.KomentarUser,
			PaymentMethod:      item.MetodePembayaran,
			NominalTransaksi:   nom,
			PersentaseKomisi:   komisi,
			KeuntunganKotor:    kotor,
			BiayaMidtrans:      fee,
			KeuntunganBersih:   bersih,
			TotalDiterimaMitra: mitra,
			FarmerEmail:        item.FarmerEmail,
			FarmerName:         item.FarmerName,
			WorkerEmail:        item.WorkerEmail,
			WorkerName:         item.WorkerName,
			DriverEmail:        item.DriverEmail,
			DriverName:         item.DriverName,
			PenjualEmail:       item.PenjualEmail,
			PembeliEmail:       item.PembeliEmail,
			BuyerEmail:         item.BuyerEmail,
			BuyerName:          item.BuyerName,
			MitraEmail:         item.MitraEmail,
			MitraName:          item.MitraName,
			UserEmail:          item.UserEmail,
			PemberiKerjaEmail:  item.PemberiKerjaEmail,
			PekerjaEmail:       item.PekerjaEmail,
			AmountPaid:         nom,
			TransactionType:    trxType,
			ContextInfo:        item.Keterangan,
			PayerName:          payerName,
		})
	}

	// Sort newest first
	sort.Slice(combinedList, func(i, j int) bool {
		return combinedList[i].TransactionDate.After(combinedList[j].TransactionDate)
	})

	totalItems := int64(len(combinedList))
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	startIdx := (page - 1) * limit
	endIdx := startIdx + limit
	if startIdx > len(combinedList) {
		startIdx = len(combinedList)
	}
	if endIdx > len(combinedList) {
		endIdx = len(combinedList)
	}

	paginatedData := combinedList[startIdx:endIdx]

	return &dto.AdminPaginationResponse{
		Data:        paginatedData,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
	}, nil
}

func (s *adminService) ExportTransactionsToExcel(search, status, layanan, sektor, startDate, endDate string) (*bytes.Buffer, error) {
	// Ambil semua transaksi dengan filter yang sama tapi tanpa pagination (limit = 99999)
	res, err := s.GetCombinedTransactions(1, 99999, search, status, layanan, sektor, startDate, endDate)
	if err != nil {
		return nil, err
	}

	dataList, ok := res.Data.([]dto.TransactionDetailResponse)
	if !ok {
		return nil, errors.New("invalid transaction data format")
	}

	f := excelize.NewFile()
	sheetName := "Transactions"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Header Style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#1677FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	headers := []string{
		"No", "ID Transaksi", "Tanggal & Waktu", "Layanan", "Sektor", "Status Transaksi",
		"Keterangan", "Komentar User", "Metode Pembayaran", "Nominal (Rp)", "Komisi (%)",
		"Keuntungan Kotor (Rp)", "Biaya Midtrans (Rp)", "Keuntungan Bersih (Rp)", "Diterima Mitra (Rp)",
		"Pembayar / Pihak Terkait",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		if i >= 26 {
			cell = fmt.Sprintf("A%s1", string(rune('A'+i-26)))
		}
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	for i, row := range dataList {
		rowNum := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), row.TransactionID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), row.TransactionDate.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), row.Layanan)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), row.Sektor)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), row.StatusTransaksi)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), row.Keterangan)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), row.KomentarUser)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), row.PaymentMethod)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), row.NominalTransaksi)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), fmt.Sprintf("%.0f%%", row.PersentaseKomisi*100))
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), row.KeuntunganKotor)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowNum), row.BiayaMidtrans)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowNum), row.KeuntunganBersih)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowNum), row.TotalDiterimaMitra)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowNum), row.PayerName)
	}

	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 16)
	f.SetColWidth(sheetName, "E", "E", 18)
	f.SetColWidth(sheetName, "F", "F", 16)
	f.SetColWidth(sheetName, "G", "G", 35)
	f.SetColWidth(sheetName, "H", "H", 35)
	f.SetColWidth(sheetName, "I", "I", 18)
	f.SetColWidth(sheetName, "J", "O", 20)
	f.SetColWidth(sheetName, "P", "P", 25)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// --- Implementasi Fungsi Payout ---

// GetPendingPayouts mengambil daftar payout yang perlu dibayar oleh admin.
func (s *adminService) GetPendingPayouts() ([]dto.PayoutDetailResponse, error) {
	payouts, err := s.payoutRepo.FindPendingPayouts()
	if err != nil {
		return nil, err
	}

	var response []dto.PayoutDetailResponse
	for _, p := range payouts {
		if err := p.LoadPayee(s.db); err != nil {
			log.Printf("Warning: Failed to load payee %s: %v", p.PayeeID, err)
			continue
		}

		dtoItem := dto.PayoutDetailResponse{
			PayoutID:   p.ID,
			Amount:     p.Amount,
			ReleasedAt: p.ReleasedAt,
			PayeeID:    p.PayeeID,
			PayeeType:  p.PayeeType,
		}

		if p.PayeeType == "worker" && p.Worker != nil {
			dtoItem.PayeeName = p.Worker.User.Name
			if p.Worker.BankName != nil {
				dtoItem.BankName = *p.Worker.BankName
				dtoItem.BankAccountNumber = *p.Worker.BankAccountNumber
				dtoItem.BankAccountHolder = *p.Worker.BankAccountHolder
			}
			if p.Transaction.Invoice.Project != nil {
				dtoItem.ContextTitle = p.Transaction.Invoice.Project.Title
			}
		} else if p.PayeeType == "driver" && p.Driver != nil {
			dtoItem.PayeeName = p.Driver.User.Name
			if p.Driver.BankName != nil {
				dtoItem.BankName = *p.Driver.BankName
				dtoItem.BankAccountNumber = *p.Driver.BankAccountNumber
				dtoItem.BankAccountHolder = *p.Driver.BankAccountHolder
			}
			if p.Transaction.Invoice.DeliveryID != nil {
				dtoItem.ContextTitle = "Pengiriman: " + p.Transaction.Invoice.Delivery.ItemDescription
			}
		}
		response = append(response, dtoItem)
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

	payout.Status = "completed"
	payout.TransferProofURL = &transferProofURL

	if err := s.payoutRepo.Update(tx, payout); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *adminService) GetPendingVerifications() ([]models.UserVerification, error) {
	return s.userVerificationRepo.FindPending()
}

// ReviewVerification memproses keputusan admin (Setuju/Tolak).
func (s *adminService) ReviewVerification(verificationID uuid.UUID, input dto.ReviewVerificationInput, adminID uuid.UUID) error {
	verification, err := s.userVerificationRepo.FindByID(verificationID)
	if err != nil {
		return errors.New("verification request not found")
	}

	if verification.Status != "pending" {
		return errors.New("this document has already been reviewed")
	}

	verification.Status = input.Status
	verification.Notes = &input.Notes
	verification.ReviewedBy = &adminID

	tx := s.db.Begin()
	if err := s.userVerificationRepo.UpdateStatus(tx, verification); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *adminService) GetAllUsers(page, limit int, search string, roleFilter string) (*dto.AdminPaginationResponse, error) {
	users, total, err := s.userRepo.FindAllUsers(page, limit, search, roleFilter)
	if err != nil {
		return nil, err
	}

	var userResponses []dto.UserDetailResponse
	for _, u := range users {
		var farmerType *string
		if u.Farmer != nil && u.Farmer.Type != "" {
			farmerType = &u.Farmer.Type
		}

		var workerSkills []string
		if u.Worker != nil && u.Worker.Skills != "" {
			if err := json.Unmarshal([]byte(u.Worker.Skills), &workerSkills); err != nil {
				workerSkills = []string{u.Worker.Skills}
			}
		}

		userResponses = append(userResponses, dto.UserDetailResponse{
			ID:            u.ID,
			Name:          u.Name,
			Email:         u.Email,
			PhoneNumber:   u.PhoneNumber,
			Role:          u.Role,
			IsActive:      u.IsActive,
			EmailVerified: u.EmailVerified,
			Type:          farmerType,
			Skills:        workerSkills,
			CreatedAt:     u.CreatedAt,
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	stats, err := s.userRepo.GetUserRoleStats()
	if err != nil {
		return nil, err
	}

	return &dto.AdminPaginationResponse{
		Data:        userResponses,
		TotalItems:  total,
		TotalPages:  totalPages,
		CurrentPage: page,
		Stats:       stats,
	}, nil
}

func (s *adminService) GetRevenueAnalytics(startDate, endDate time.Time) (*dto.RevenueAnalyticsResponse, error) {
	svcTotal, svcTrend, err := s.transactionRepo.GetRevenueStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	prodTotal, prodTrend, err := s.ecommPaymentRepo.GetRevenueStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	trendMap := make(map[string]float64)

	for _, t := range svcTrend {
		dateStr := t.Date
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		trendMap[dateStr] += t.Value
	}
	for _, t := range prodTrend {
		dateStr := t.Date
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		trendMap[dateStr] += t.Value
	}

	var combinedTrend []dto.DailyDataPoint
	for date, value := range trendMap {
		combinedTrend = append(combinedTrend, dto.DailyDataPoint{
			Date:  date,
			Value: value,
		})
	}

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