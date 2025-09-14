package services

import (
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService interface {
    SendEmail(toEmail, toName, subject, htmlContent string) error
}

type sendgridService struct {
    apiKey string
}

func NewEmailService() EmailService {
    return &sendgridService{apiKey: os.Getenv("SENDGRID_API_KEY")}
}

func (s *sendgridService) SendEmail(toEmail, toName, subject, htmlContent string) error {
    from := mail.NewEmail("AgroLink Platform", "noreply@agrolink.com") // Ganti dengan email Anda
    to := mail.NewEmail(toName, toEmail)
    message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
    client := sendgrid.NewSendClient(s.apiKey)
    _, err := client.Send(message)
    return err
}