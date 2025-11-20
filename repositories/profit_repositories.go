// file: repositories/profit_repository.go
package repositories

import (
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ProfitRepository interface {
	// sourceType: "" | "utama" | "ecommerce"
	GetDailySummary(start, end time.Time, sourceType string) ([]dto.PlatformProfitDailySummaryResponse, error)
	GetTotalSummary(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, error)
}

type profitRepository struct {
	db *gorm.DB
}

func NewProfitRepository(db *gorm.DB) ProfitRepository {
	return &profitRepository{db: db}
}

// GetTotalSummary → agregat total gross/gateway/net dalam periode
func (r *profitRepository) GetTotalSummary(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, error) {
	var result dto.PlatformProfitTotalSummaryResponse

	q := r.db.Model(&models.PlatformProfit{}).
		Where("profit_date BETWEEN ? AND ?", start, end)

	// filter source_type kalau diisi
	if sourceType != "" {
		q = q.Where("source_type = ?", sourceType)
	}

	// SUM gross, gateway, net, count(*) sebagai total transaksi
	if err := q.Select(`
		COALESCE(SUM(gross_profit), 0)      AS total_gross_profit,
		COALESCE(SUM(gateway_fee), 0)       AS total_gateway_fee,
		COALESCE(SUM(net_profit), 0)        AS total_net_profit,
		COUNT(*)                            AS total_transactions
	`).Scan(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDailySummary → agregat per hari dalam periode
func (r *profitRepository) GetDailySummary(start, end time.Time, sourceType string) ([]dto.PlatformProfitDailySummaryResponse, error) {
	var rows []dto.PlatformProfitDailySummaryResponse

	q := r.db.Model(&models.PlatformProfit{}).
		Where("profit_date BETWEEN ? AND ?", start, end)

	if sourceType != "" {
		q = q.Where("source_type = ?", sourceType)
	}

	// Group per tanggal profit_date
	if err := q.Select(`
		DATE(profit_date)                            AS date,
		COALESCE(SUM(gross_profit), 0)              AS gross_profit,
		COALESCE(SUM(gateway_fee), 0)               AS gateway_fee,
		COALESCE(SUM(net_profit), 0)                AS net_profit,
		COUNT(*)                                    AS transaction_count
	`).
		Group("DATE(profit_date)").
		Order("DATE(profit_date)").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Format tanggal ke "YYYY-MM-DD" kalau perlu (kalau Date sudah string, ini bisa di-skip)
	for i := range rows {
		if len(rows[i].Date) > 10 {
			rows[i].Date = rows[i].Date[:10]
		}
	}

	return rows, nil
}
