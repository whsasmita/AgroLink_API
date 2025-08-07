package utils

import (
	"log"
)

func SendVerificationEmail(email string) error {
	log.Println("Sending verification email to:", email)
	// Simulasikan delay pengiriman email, jika ingin
	// time.Sleep(2 * time.Second)
	return nil
}
