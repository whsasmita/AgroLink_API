package dto

// PaginationRequest menampung parameter query untuk pagination, sorting, dan filtering.
type PaginationRequest struct {
	Page    int    `form:"page,default=1"`
	Limit   int    `form:"limit,default=10"`
	Sort    string `form:"sort,default=created_at desc"` // Contoh: "start_date asc", "payment_rate desc"
	// Nanti kita bisa tambahkan filter di sini
}

// PaginationResponse adalah wrapper standar untuk response list.
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}