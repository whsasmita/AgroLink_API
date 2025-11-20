package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
)

type ProfitRepository interface {
	GetDailySummary(start, end time.Time, sourceType string) ([]dto.PlatformProfitDailySummaryResponse, error)
	GetTotalSummary(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, error)
}

type profitRepository struct {
	db *gorm.DB
}

func NewProfitRepository(db *gorm.DB) ProfitRepository {
	return &profitRepository{db: db}
}

// GetDailySummary mengembalikan ringkasan profit per hari per source_type
func (r *profitRepository) GetDailySummary(start, end time.Time, sourceType string) ([]dto.PlatformProfitDailySummaryResponse, error) {
	var results []dto.PlatformProfitDailySummaryResponse

	query := r.db.Model(&models.PlatformProfit{}).
		Select(
			"DATE(profit_date) AS date",
			"source_type",
			"SUM(gross_profit) AS total_gross_profit",
			"SUM(gateway_fee) AS total_gateway_fee",
			"SUM(net_profit) AS total_net_profit",
		).
		Where("profit_date BETWEEN ? AND ?", start, end)

	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}

	err := query.
		Group("DATE(profit_date), source_type").
		Order("DATE(profit_date) ASC, source_type ASC").
		Scan(&results).Error

	return results, err
}

// GetTotalSummary mengembalikan ringkasan total untuk rentang tanggal
func (r *profitRepository) GetTotalSummary(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, error) {
	var result dto.PlatformProfitTotalSummaryResponse

	query := r.db.Model(&models.PlatformProfit{}).
		Select(
			"COALESCE(SUM(gross_profit), 0) AS total_gross_profit",
			"COALESCE(SUM(gateway_fee), 0)  AS total_gateway_fee",
			"COALESCE(SUM(net_profit), 0)   AS total_net_profit",
		).
		Where("profit_date BETWEEN ? AND ?", start, end)

	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}

	err := query.Scan(&result).Error

	return &result, err
}
