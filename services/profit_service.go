package services

import (
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type ProfitService interface {
	GetPlatformProfitReport(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, []dto.PlatformProfitDailySummaryResponse, error)
}

type profitService struct {
	profitRepo repositories.ProfitRepository
}

func NewProfitService(profitRepo repositories.ProfitRepository) ProfitService {
	return &profitService{profitRepo: profitRepo}
}

// GetPlatformProfitReport mengembalikan ringkasan total + detail per hari
func (s *profitService) GetPlatformProfitReport(start, end time.Time, sourceType string) (*dto.PlatformProfitTotalSummaryResponse, []dto.PlatformProfitDailySummaryResponse, error) {
	daily, err := s.profitRepo.GetDailySummary(start, end, sourceType)
	if err != nil {
		return nil, nil, err
	}

	total, err := s.profitRepo.GetTotalSummary(start, end, sourceType)
	if err != nil {
		return nil, nil, err
	}

	return total, daily, nil
}
