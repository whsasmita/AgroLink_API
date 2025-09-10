package services

import (
	"github.com/whsasmita/AgroLink_API/repositories"
)

type PayoutService interface {
	// GetPendingPayouts() ([]dto.PayoutDetailResponse, error)
}

type payoutService struct {
	payoutRepo repositories.PayoutRepository
}

func NewPayoutService(payoutRepo repositories.PayoutRepository) PayoutService {
	return &payoutService{payoutRepo: payoutRepo}
}

// func (s *payoutService) GetPendingPayouts() ([]dto.PayoutDetailResponse, error) {
// 	payouts, err := s.payoutRepo.FindPendingPayouts()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var responseDTOs []dto.PayoutDetailResponse
// 	for _, p := range payouts {
// 		dto := dto.PayoutDetailResponse{
// 			PayoutID:   p.ID,
// 			WorkerID:   p.WorkerID,
// 			Amount:     p.Amount,
// 			ReleasedAt: p.ReleasedAt,
// 			Status:     p.Status,
// 		}
// 		if p.Worker.User.Name != "" {
// 			dto.WorkerName = p.Worker.User.Name
// 		}
// 		if p.Transaction.Invoice.Project.Title != "" {
// 			dto.ProjectTitle = p.Transaction.Invoice.Project.Title
// 		}
// 		responseDTOs = append(responseDTOs, dto)
// 	}

// 	return responseDTOs, nil
// }