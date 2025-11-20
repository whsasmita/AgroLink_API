package dto

// Ringkasan per hari per sumber (utama/ecommerce)
type PlatformProfitDailySummaryResponse struct {
	Date            string  `json:"date"`              // format: 2006-01-02
	SourceType      string  `json:"source_type"`       // "utama" / "ecommerce"
	TotalGross      float64 `json:"total_gross_profit"`
	TotalGatewayFee float64 `json:"total_gateway_fee"`
	TotalNet        float64 `json:"total_net_profit"`
}

// Ringkasan total seluruh transaksi di periode
type PlatformProfitTotalSummaryResponse struct {
	TotalGross      float64 `json:"total_gross_profit"`
	TotalGatewayFee float64 `json:"total_gateway_fee"`
	TotalNet        float64 `json:"total_net_profit"`
}
