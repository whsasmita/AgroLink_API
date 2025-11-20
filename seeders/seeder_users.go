package seeders

type SeedUserRow struct {
	Nama   string `json:"Nama"`
	NoHP   string `json:"No HP"`
	Email  string `json:"Email"`
	Alamat string `json:"Alamat"`
	Role   string `json:"Role"`

	Password string `json:"Password"`

	// Worker-specific
	Skills            []string `json:"Skills,omitempty"`
	DailyRate         *float64 `json:"DailyRate,omitempty"`
	NationalID        string   `json:"NationalID,omitempty"`
	BankName          string   `json:"BankName,omitempty"`
	BankAccountNumber string   `json:"BankAccountNumber,omitempty"`
	BankAccountHolder string   `json:"BankAccountHolder,omitempty"`

	// Driver-specific
	PricingScheme map[string]float64 `json:"PricingScheme,omitempty"`
	VehicleTypes  []string           `json:"VehicleTypes,omitempty"`
	CurrentLat    *float64           `json:"CurrentLat,omitempty"`
	CurrentLng    *float64           `json:"CurrentLng,omitempty"`
}
