package dto

// Data untuk widget KPI
type DashboardKPIs struct {
	TotalRevenueMonthly float64 `json:"total_revenue_monthly"`
	PendingPayoutsTotal float64 `json:"pending_payouts_total"`
	NewUsersMonthly     int     `json:"new_users_monthly"`
	ActiveProjects      int     `json:"active_projects"`
	ActiveDeliveries    int     `json:"active_deliveries"`
	NewECommerceOrders  int     `json:"new_e_commerce_orders"`
}

// Data ringkasan sub-kategori sektor pekerja
type WorkerSectorSummary struct {
	Sector          string  `json:"sector"` // "Pertanian", "Peternakan", "Tukang Bangunan"
	SuccessCount    int     `json:"success_count"`
	FailedCount     int     `json:"failed_count"`
	TotalCount      int     `json:"total_count"`
	TotalAmount     float64 `json:"total_amount"`
	GrossProfit     float64 `json:"gross_profit"`
	GatewayFee      float64 `json:"gateway_fee"`
	NetProfit       float64 `json:"net_profit"`
	TotalMitraShare float64 `json:"total_mitra_share"`
}

// Data ringkasan layanan
type ServiceSummary struct {
	Name             string  `json:"name"`              // e.g. "Pekerja", "Ekspedisi", "E-Commerce", "Chatbot Premium", "Tukang", "Peternak", "Kemitraan"
	Sector           string  `json:"sector"`            // "Pertanian", "Peternakan", "Tukang Bangunan", "-"
	SuccessCount     int     `json:"success_count"`     // Transaksi berhasil
	FailedCount      int     `json:"failed_count"`      // Transaksi gagal
	TransactionCount int     `json:"transaction_count"` // Total transaksi pada layanan ini
	TotalAmount      float64 `json:"total_amount"`      // Total GMV / nilai transaksi
	GrossProfit      float64 `json:"gross_profit"`      // Keuntungan kotor
	GatewayFee       float64 `json:"gateway_fee"`       // Biaya Midtrans
	NetProfit        float64 `json:"net_profit"`        // Keuntungan bersih
	TotalMitraShare  float64 `json:"total_mitra_share"` // Total diterima mitra
	Percentage       float64 `json:"percentage"`        // Persentase kontribusi terhadap total GMV (%)
}

// Data ringkasan laba bulanan (Sep 2025 – Sep 2026)
type MonthlyProfitPoint struct {
	Month               string  `json:"month"`                 // "2025-09"
	MonthLabel          string  `json:"month_label"`           // "Sep 2025"
	SuccessCount        int     `json:"success_count"`
	FailedCount         int     `json:"failed_count"`
	TotalCount          int     `json:"total_count"`
	GMV                 float64 `json:"gmv"`
	GrossProfit         float64 `json:"gross_profit"`
	GatewayFee          float64 `json:"gateway_fee"`
	NetProfit           float64 `json:"net_profit"`
	CumulativeNetProfit float64 `json:"cumulative_net_profit"`
}

// Data status breakdown per Layanan x Status
type TransactionStatusBreakdown struct {
	Layanan     string  `json:"layanan"`
	Sektor      string  `json:"sektor"`
	Sukses      int     `json:"sukses"`
	Gagal       int     `json:"gagal"`
	Total       int     `json:"total"`
	GrossProfit float64 `json:"gross_profit"`
	NetProfit   float64 `json:"net_profit"`
}

// Data ringkasan keuangan platform menyeluruh
type DashboardFinancialSummary struct {
	TotalTransactions      int     `json:"total_transactions"`
	SuccessfulTransactions int     `json:"successful_transactions"`
	FailedTransactions     int     `json:"failed_transactions"`
	TotalGMV               float64 `json:"total_gmv"`
	TotalGrossProfit       float64 `json:"total_gross_profit"`
	TotalGatewayFee        float64 `json:"total_gateway_fee"`
	TotalNetProfit         float64 `json:"total_net_profit"`
	TotalMitraShare        float64 `json:"total_mitra_share"`
	Phase1Transactions     int     `json:"phase1_transactions"`
	Phase1Success          int     `json:"phase1_success"`
	Phase1Failed           int     `json:"phase1_failed"`
	Phase1GMV              float64 `json:"phase1_gmv"`
	Phase1GrossProfit      float64 `json:"phase1_gross_profit"`
	Phase1GatewayFee       float64 `json:"phase1_gateway_fee"`
	Phase1NetProfit        float64 `json:"phase1_net_profit"`
	Phase2Transactions     int     `json:"phase2_transactions"`
	Phase2Success          int     `json:"phase2_success"`
	Phase2Failed           int     `json:"phase2_failed"`
	Phase2GMV              float64 `json:"phase2_gmv"`
	Phase2GrossProfit      float64 `json:"phase2_gross_profit"`
	Phase2GatewayFee       float64 `json:"phase2_gateway_fee"`
	Phase2NetProfit        float64 `json:"phase2_net_profit"`
}

// Data statistik pengguna lengkap
type DashboardUserStats struct {
	TotalUsers   int64 `json:"total_users"`
	TotalWorker  int64 `json:"total_worker"`
	TotalFarmer  int64 `json:"total_farmer"`
	TotalDriver  int64 `json:"total_driver"`
	TotalGeneral int64 `json:"total_general"`
	TotalMitra   int64 `json:"total_mitra"`
	TotalAdmin   int64 `json:"total_admin"`
}

// Data untuk antrean "Butuh Tindakan"
type DashboardActionQueue struct {
	PendingVerifications int `json:"pending_verifications"`
	PendingPayouts       int `json:"pending_payouts"`
	OpenDisputes         int `json:"open_disputes"`
}

// Data untuk grafik
type DailyDataPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Respon DTO utama dashboard admin
type AdminDashboardResponse struct {
	KPIs                  DashboardKPIs                `json:"kpis"`
	FinancialSummary      DashboardFinancialSummary    `json:"financial_summary"`
	ServiceSummaries      []ServiceSummary             `json:"service_summaries"`
	WorkerSectorSummaries []WorkerSectorSummary        `json:"worker_sector_summaries"`
	MonthlyProfitTrend    []MonthlyProfitPoint         `json:"monthly_profit_trend"`
	StatusBreakdown       []TransactionStatusBreakdown `json:"status_breakdown"`
	UserStats             DashboardUserStats           `json:"user_stats"`
	ActionQueue           DashboardActionQueue         `json:"action_queue"`
	RevenueTrend          []DailyDataPoint             `json:"revenue_trend"`
	UserTrend             []DailyDataPoint             `json:"user_trend"`
}