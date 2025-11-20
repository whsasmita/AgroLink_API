// File: services/worker_service.go
package services

import (
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/repositories"
)

// WorkerService mendefinisikan interface untuk logika bisnis terkait worker
type WorkerService interface {
	GetWorkers(search, sortBy, order string, limit, offset int, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate float64) ([]dto.WorkerResponse, int64, error)
	GetWorkerProfile(id string) (dto.WorkerResponse, error)
}

type workerService struct {
	repo repositories.WorkerRepository
}

func NewWorkerService(repo repositories.WorkerRepository) WorkerService {
	return &workerService{repo}
}

// GetWorkers berfungsi sebagai jembatan antara handler dan repository
func (s *workerService) GetWorkers(search, sortBy, order string, limit, offset int, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate float64) ([]dto.WorkerResponse, int64, error) {
    workers, total, err := s.repo.GetWorkers(search, sortBy, order, limit, offset, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate)
    if err != nil {
        return nil, 0, err
    }

    var workerResponses []dto.WorkerResponse
    for _, worker := range workers {
        // Mapping dari models.Worker ke dtos.WorkerResponse
        workerResponses = append(workerResponses, dto.WorkerResponse{
            UserID:               worker.UserID,
            Skills:               worker.Skills,
            HourlyRate:           worker.HourlyRate,
            DailyRate:            worker.DailyRate,
            Address:              worker.Address,
            AvailabilitySchedule: worker.AvailabilitySchedule,
            CurrentLocationLat:   worker.CurrentLocationLat,
            CurrentLocationLng:   worker.CurrentLocationLng,
            Rating:               worker.Rating,
            TotalJobsCompleted:   worker.TotalJobsCompleted,
            CreatedAt:            worker.CreatedAt,
            Name:                 worker.User.Name,
            Email:                worker.User.Email,
            PhoneNumber:          worker.User.PhoneNumber,
            ProfilePicture: worker.User.ProfilePicture,
        })
    }

    return workerResponses, total, nil
}


func (s *workerService) GetWorkerProfile(id string) (dto.WorkerResponse, error) {
    worker, err := s.repo.GetWorkerByID(id)
    if err != nil {
        return dto.WorkerResponse{}, err
    }

    // Mapping dari models.Worker ke dtos.WorkerResponse
    response := dto.WorkerResponse{
        UserID:               worker.UserID,
        Skills:               worker.Skills,
        HourlyRate:           worker.HourlyRate,
        DailyRate:            worker.DailyRate,
        Address:              worker.Address,
        AvailabilitySchedule: worker.AvailabilitySchedule,
        CurrentLocationLat:   worker.CurrentLocationLat,
        CurrentLocationLng:   worker.CurrentLocationLng,
        Rating:               worker.Rating,
        TotalJobsCompleted:   worker.TotalJobsCompleted,
        CreatedAt:            worker.CreatedAt,
        Name:                 worker.User.Name,
        Email:                worker.User.Email,
        PhoneNumber:          worker.User.PhoneNumber,
        ProfilePicture: worker.User.ProfilePicture,
    }

    return response, nil
}
