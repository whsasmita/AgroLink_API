package models

import (
	"github.com/google/uuid"
)

// Common constants for the models package
const (
	// User roles
	RoleFarmer     = "farmer"
	RoleWorker     = "worker"
	RoleExpedition = "expedition"
	RoleAdmin      = "admin"
	RoleCS         = "cs"

	// Project types
	ProjectTypePlanting     = "planting"
	ProjectTypeMaintenance  = "maintenance"
	ProjectTypeHarvesting   = "harvesting"
	ProjectTypeIrrigation   = "irrigation"
	ProjectTypePestControl  = "pest_control"

	// Project status
	ProjectStatusDraft      = "draft"
	ProjectStatusOpen       = "open"
	ProjectStatusInProgress = "in_progress"
	ProjectStatusCompleted  = "completed"
	ProjectStatusCancelled  = "cancelled"

	// Application status
	ApplicationStatusPending   = "pending"
	ApplicationStatusAccepted  = "accepted"
	ApplicationStatusRejected  = "rejected"
	ApplicationStatusWithdrawn = "withdrawn"

	// Assignment status
	AssignmentStatusAssigned   = "assigned"
	AssignmentStatusStarted    = "started"
	AssignmentStatusPaused     = "paused"
	AssignmentStatusCompleted  = "completed"
	AssignmentStatusTerminated = "terminated"

	// Delivery status
	DeliveryStatusScheduled        = "scheduled"
	DeliveryStatusPickupPending    = "pickup_pending"
	DeliveryStatusPickedUp         = "picked_up"
	DeliveryStatusInTransit        = "in_transit"
	DeliveryStatusOutForDelivery   = "out_for_delivery"
	DeliveryStatusDelivered        = "delivered"
	DeliveryStatusFailed           = "failed"
	DeliveryStatusCancelled        = "cancelled"

	// Transaction status
	TransactionStatusPending   = "pending"
	TransactionStatusHold      = "hold"
	TransactionStatusReleased  = "released"
	TransactionStatusCancelled = "cancelled"
	TransactionStatusRefunded  = "refunded"

	// Transaction types
	TransactionTypeWorkPayment     = "work_payment"
	TransactionTypeDeliveryPayment = "delivery_payment"
	TransactionTypePlatformFee     = "platform_fee"
	TransactionTypeRefund          = "refund"

	// Contract types
	ContractTypeWork        = "work"
	ContractTypeDelivery    = "delivery"
	ContractTypeMaintenance = "maintenance"

	// Contract status
	ContractStatusDraft            = "draft"
	ContractStatusPendingSignature = "pending_signature"
	ContractStatusActive           = "active"
	ContractStatusCompleted        = "completed"
	ContractStatusTerminated       = "terminated"
	ContractStatusExpired          = "expired"

	// Notification types
	NotificationTypeInfo    = "info"
	NotificationTypeWarning = "warning"
	NotificationTypeSuccess = "success"
	NotificationTypeError   = "error"

	// Notification categories
	NotificationCategoryJob              = "job"
	NotificationCategoryPayment          = "payment"
	NotificationCategoryDelivery         = "delivery"
	NotificationCategorySystem           = "system"
	NotificationCategoryAIRecommendation = "ai_recommendation"
	NotificationCategorySchedule         = "schedule"

	// Priority levels
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"

	// Review types
	ReviewTypeWorkQuality     = "work_quality"
	ReviewTypeDeliveryService = "delivery_service"
	ReviewTypeCommunication   = "communication"
	ReviewTypePunctuality     = "punctuality"
	ReviewTypeOverall         = "overall"

	// Support ticket categories
	SupportCategoryTechnical = "technical"
	SupportCategoryPayment   = "payment"
	SupportCategoryDelivery  = "delivery"
	SupportCategoryAccount   = "account"
	SupportCategoryDispute   = "dispute"
	SupportCategoryOther     = "other"

	// Support ticket status
	SupportStatusOpen            = "open"
	SupportStatusInProgress      = "in_progress"
	SupportStatusWaitingCustomer = "waiting_customer"
	SupportStatusResolved        = "resolved"
	SupportStatusClosed          = "closed"

	// Dispute types
	DisputeTypePaymentIssue     = "payment_issue"
	DisputeTypeWorkQuality      = "work_quality"
	DisputeTypeDeliveryProblem  = "delivery_problem"
	DisputeTypeContractBreach   = "contract_breach"
	DisputeTypeOther            = "other"

	// Dispute status
	DisputeStatusOpen        = "open"
	DisputeStatusUnderReview = "under_review"
	DisputeStatusMediation   = "mediation"
	DisputeStatusResolved    = "resolved"
	DisputeStatusClosed      = "closed"
	DisputeStatusEscalated   = "escalated"
)

// Helper functions for model operations

// IsValidRole checks if the given role is valid
func IsValidRole(role string) bool {
	validRoles := []string{RoleFarmer, RoleWorker, RoleExpedition, RoleAdmin, RoleCS}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsValidProjectType checks if the given project type is valid
func IsValidProjectType(projectType string) bool {
	validTypes := []string{
		ProjectTypePlanting,
		ProjectTypeMaintenance,
		ProjectTypeHarvesting,
		ProjectTypeIrrigation,
		ProjectTypePestControl,
	}
	for _, validType := range validTypes {
		if projectType == validType {
			return true
		}
	}
	return false
}

// IsValidUUIDString checks if string is a valid UUID
func IsValidUUIDString(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// Common validation functions for ratings
func IsValidRating(rating int) bool {
	return rating >= 1 && rating <= 5
}

// Common response structures for API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Pagination helper
type Pagination struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

// GetOffset calculates offset for database queries
func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit returns limit with default value
func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 || p.Limit > 100 {
		return 20 // Default limit
	}
	return p.Limit
}

// Search filters for common queries
type ProjectFilter struct {
	Pagination
	Status       string  `json:"status" form:"status"`
	ProjectType  string  `json:"project_type" form:"project_type"`
	MinBudget    float64 `json:"min_budget" form:"min_budget"`
	MaxBudget    float64 `json:"max_budget" form:"max_budget"`
	Skills       string  `json:"skills" form:"skills"`
	UrgencyLevel string  `json:"urgency_level" form:"urgency_level"`
	Location     string  `json:"location" form:"location"`
	Radius       int     `json:"radius" form:"radius"`
}

type WorkerFilter struct {
	Pagination
	Skills     string  `json:"skills" form:"skills"`
	MinRating  float64 `json:"min_rating" form:"min_rating"`
	MaxRate    float64 `json:"max_rate" form:"max_rate"`
	Location   string  `json:"location" form:"location"`
	Radius     int     `json:"radius" form:"radius"`
	Available  bool    `json:"available" form:"available"`
}

type ExpeditionFilter struct {
	Pagination
	ServiceAreas string  `json:"service_areas" form:"service_areas"`
	VehicleTypes string  `json:"vehicle_types" form:"vehicle_types"`
	MinRating    float64 `json:"min_rating" form:"min_rating"`
	MaxPrice     float64 `json:"max_price" form:"max_price"`
}