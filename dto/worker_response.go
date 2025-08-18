// file: dtos/worker_response.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

type WorkerResponse struct {
	UserID               uuid.UUID `json:"user_id"`
	Name                 string    `json:"name"`
	Email                string    `json:"email"`
	ProfilePicture       *string   `json:"profile_picture"`
	PhoneNumber          *string   `json:"phone_number"`
	Skills               string    `json:"skills"`
	HourlyRate           *float64  `json:"hourly_rate"`
	DailyRate            *float64  `json:"daily_rate"`
	Address              *string   `json:"address"`
	AvailabilitySchedule *string   `json:"availability_schedule"`
	CurrentLocationLat   *float64  `json:"current_location_lat"`
	CurrentLocationLng   *float64  `json:"current_location_lng"`
	Rating               float64   `json:"rating"`
	TotalJobsCompleted   int       `json:"total_jobs_completed"`
	CreatedAt            time.Time `json:"created_at"`

	// Field dari model User yang di-flatten

	// Relasi lainnya (jika perlu)
	ProjectApplications interface{} `json:"ProjectApplications"`
	ProjectAssignments  interface{} `json:"ProjectAssignments"`
	WorkerAvailability  interface{} `json:"WorkerAvailability"`
}
