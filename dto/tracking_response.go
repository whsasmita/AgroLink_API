package dto

type UpdateLocationRequest struct {
	Lat float64 `json:"latitude" binding:"required"`
	Lng float64 `json:"longitude" binding:"required"`
}