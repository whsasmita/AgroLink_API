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

// Data ringkasan layanan
type ServiceSummary struct {
	Name             string  `json:"name"`              // e.g. "Pekerja", "Ekspedisi", "E-Commerce", "Chatbot Premium", "Tukang", "Peternak", "Kemitraan"
	TransactionCount int     `json:"transaction_count"` // Jumlah transaksi pada layanan ini
	TotalAmount      float64 `json:"total_amount"`      // Total GMV / nilai transaksi
	GrossProfit      float64 `json:"gross_profit"`      // Keuntungan kotor
	GatewayFee       float64 `json:"gateway_fee"`       // Biaya Midtrans
	NetProfit        float64 `json:"net_profit"`        // Keuntungan bersih
	TotalMitraShare  float64 `json:"total_mitra_share"` // Total diterima mitra
	Percentage       float64 `json:"percentage"`        // Persentase kontribusi terhadap total GMV (%)
}

// Data ringkasan keuangan platform menyeluruh
type DashboardFinancialSummary struct {
	TotalTransactions  int     `json:"total_transactions"`
	TotalGMV           float64 `json:"total_gmv"`
	TotalGrossProfit   float64 `json:"total_gross_profit"`
	TotalGatewayFee    float64 `json:"total_gateway_fee"`
	TotalNetProfit     float64 `json:"total_net_profit"`
	TotalMitraShare    float64 `json:"total_mitra_share"`
	Phase1Transactions int     `json:"phase1_transactions"`
	Phase1GMV          float64 `json:"phase1_gmv"`
	Phase1GrossProfit  float64 `json:"phase1_gross_profit"`
	Phase1GatewayFee   float64 `json:"phase1_gateway_fee"`
	Phase1NetProfit    float64 `json:"phase1_net_profit"`
	Phase2Transactions int     `json:"phase2_transactions"`
	Phase2GMV          float64 `json:"phase2_gmv"`
	Phase2GrossProfit  float64 `json:"phase2_gross_profit"`
	Phase2GatewayFee   float64 `json:"phase2_gateway_fee"`
	Phase2NetProfit    float64 `json:"phase2_net_profit"`
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
	KPIs             DashboardKPIs             `json:"kpis"`
	FinancialSummary DashboardFinancialSummary `json:"financial_summary"`
	ServiceSummaries []ServiceSummary          `json:"service_summaries"`
	UserStats        DashboardUserStats        `json:"user_stats"`
	ActionQueue      DashboardActionQueue      `json:"action_queue"`
	RevenueTrend     []DailyDataPoint          `json:"revenue_trend"`
	UserTrend        []DailyDataPoint          `json:"user_trend"`
}