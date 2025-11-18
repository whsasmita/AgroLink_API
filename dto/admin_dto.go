package dto

// ...

// Input untuk endpoint MarkPayoutAsCompleted
type MarkPayoutCompletedInput struct {
	TransferProofURL string `json:"transfer_proof_url" binding:"required,url"`
}

type ReviewVerificationInput struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Notes  string `json:"notes"` // Catatan dari admin (misal: "Foto KTP buram")
}