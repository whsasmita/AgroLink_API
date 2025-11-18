package dto

// Data untuk widget KPI
type DashboardKPIs struct {
	TotalRevenueMonthly   float64 `json:"total_revenue_monthly"`
	PendingPayoutsTotal   float64 `json:"pending_payouts_total"`
	NewUsersMonthly       int     `json:"new_users_monthly"`
	ActiveProjects        int     `json:"active_projects"`
	ActiveDeliveries      int     `json:"active_deliveries"`
	NewECommerceOrders    int     `json:"new_e_commerce_orders"`
}

// Data untuk antrean "Butuh Tindakan"
type DashboardActionQueue struct {
	PendingVerifications int `json:"pending_verifications"`
	PendingPayouts       int `json:"pending_payouts"`
	OpenDisputes         int `json:"open_disputes"`
}

// Data untuk grafik (contoh sederhana)
type DailyDataPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Respon DTO utama
type AdminDashboardResponse struct {
	KPIs         DashboardKPIs          `json:"kpis"`
	ActionQueue  DashboardActionQueue   `json:"action_queue"`
	RevenueTrend []DailyDataPoint       `json:"revenue_trend"`
	UserTrend    []DailyDataPoint       `json:"user_trend"`
}