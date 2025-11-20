// file: services/profit_service.go
package services

import (
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/repositories"
)

// ProfitService menangani logika laporan keuntungan platform
type ProfitService interface {
	// sourceType: "" | "utama" | "ecommerce"
	GetPlatformProfitReport(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, []dto.PlatformProfitDailySummaryResponse, error)
}

type profitService struct {
	profitRepo repositories.ProfitRepository
}

func NewProfitService(profitRepo repositories.ProfitRepository) ProfitService {
	return &profitService{profitRepo: profitRepo}
}

func (s *profitService) GetPlatformProfitReport(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, []dto.PlatformProfitDailySummaryResponse, error) {
	// 1. Ambil ringkasan harian
	daily, err := s.profitRepo.GetDailySummary(start, end, sourceType)
	if err != nil {
		return nil, nil, err
	}

	// 2. Ambil ringkasan total
	total, err := s.profitRepo.GetTotalSummary(start, end, sourceType)
	if err != nil {
		return nil, nil, err
	}

	return total, daily, nil
}
