// models/driver_route.go
package models

import "github.com/google/uuid"

type DriverRoute struct {
    ID            uuid.UUID `gorm:"type:char(36);primary_key"`
    DriverID      uuid.UUID `gorm:"type:char(36);not null"`
    Origin        string    // Kota/area asal
    Destination   string    // Kota/area tujuan
    DaysAvailable string    // Contoh: "Senin, Rabu, Jumat"
}