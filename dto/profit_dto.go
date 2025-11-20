package dto

type PlatformProfitTotalSummaryResponse struct {
	TotalGrossProfit  float64 `json:"total_gross_profit"`
	TotalGatewayFee   float64 `json:"total_gateway_fee"`
	TotalNetProfit    float64 `json:"total_net_profit"`
	TotalTransactions int64   `json:"total_transactions"`
}

type PlatformProfitDailySummaryResponse struct {
	Date             string  `json:"date"` // "YYYY-MM-DD"
	GrossProfit      float64 `json:"gross_profit"`
	GatewayFee       float64 `json:"gateway_fee"`
	NetProfit        float64 `json:"net_profit"`
	TransactionCount int64   `json:"transaction_count"`
}