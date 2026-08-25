package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService interface {
	SendEmail(toEmail, toName, subject, htmlContent string) error
	SendOTPEmail(toEmail, toName, otpCode string) error
}

type emailService struct {
	sendgridKey  string
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewEmailService() EmailService {
	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "noreply@agrolink.com"
	}

	return &emailService{
		sendgridKey:  os.Getenv("SENDGRID_API_KEY"),
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    from,
	}
}

func (s *emailService) SendEmail(toEmail, toName, subject, htmlContent string) error {
	// Selalu log ke terminal agar developer/reviewer bisa melihat aktivitas pengiriman email
	log.Printf("📧 [EMAIL DISPATCH] To: %s <%s> | Subject: %s", toName, toEmail, subject)

	// 1. Coba kirim via SendGrid jika API Key tersedia
	if s.sendgridKey != "" {
		from := mail.NewEmail("AgroLink Platform", s.fromEmail)
		to := mail.NewEmail(toName, toEmail)
		message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
		client := sendgrid.NewSendClient(s.sendgridKey)
		res, err := client.Send(message)
		if err == nil && res.StatusCode < 400 {
			log.Printf("✅ [SENDGRID SUCCESS] Email sent to %s", toEmail)
			return nil
		}
		log.Printf("⚠️ [SENDGRID WARN] Failed to send via SendGrid: %v", err)
	}

	// 2. Coba kirim via SMTP jika konfigurasi host dan password tersedia
	if s.smtpHost != "" && s.smtpUsername != "" && s.smtpPassword != "" && !strings.Contains(s.smtpUsername, "your-email") {
		auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
		addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

		headers := make(map[string]string)
		headers["From"] = fmt.Sprintf("AgroLink Platform <%s>", s.fromEmail)
		headers["To"] = toEmail
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		var msg strings.Builder
		for k, v := range headers {
			msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		msg.WriteString("\r\n" + htmlContent)

		err := smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, []byte(msg.String()))
		if err != nil {
			log.Printf("⚠️ [SMTP WARN] Failed to send via SMTP (%s): %v", addr, err)
			return err
		}
		log.Printf("✅ [SMTP SUCCESS] Email sent to %s", toEmail)
		return nil
	}

	log.Printf("ℹ️ [EMAIL MOCK] Email simulated (SMTP/SendGrid credentials not configured). Content ready for %s.", toEmail)
	return nil
}

func (s *emailService) SendOTPEmail(toEmail, toName, otpCode string) error {
	subject := fmt.Sprintf("Kode Verifikasi Pendaftaran AgroLink: %s", otpCode)
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verifikasi Email AgroLink</title>
  <style>
    body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f7f6; margin: 0; padding: 20px; color: #333; }
    .container { max-width: 540px; margin: 0 auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.08); }
    .header { background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px 20px; text-align: center; color: #ffffff; }
    .header h1 { margin: 0; font-size: 26px; font-weight: 700; letter-spacing: 0.5px; }
    .header p { margin: 8px 0 0; font-size: 14px; opacity: 0.9; }
    .content { padding: 32px 28px; text-align: center; }
    .greeting { font-size: 16px; margin-bottom: 20px; text-align: left; color: #4b5563; }
    .message { font-size: 15px; line-height: 1.6; color: #4b5563; margin-bottom: 28px; text-align: left; }
    .otp-box { background: #ecfdf5; border: 2px dashed #10b981; border-radius: 10px; padding: 20px; margin: 0 auto 28px; display: inline-block; }
    .otp-code { font-size: 36px; font-weight: 800; color: #047857; letter-spacing: 8px; font-family: 'Courier New', Courier, monospace; }
    .expiry { font-size: 13px; color: #6b7280; margin-top: 10px; }
    .alert-box { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 12px 16px; margin: 20px 0; border-radius: 4px; text-align: left; font-size: 13px; color: #92400e; }
    .footer { background: #f9fafb; padding: 20px; text-align: center; font-size: 12px; color: #9ca3af; border-top: 1px solid #e5e7eb; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>🌱 AgroLink Platform</h1>
      <p>Solusi Digital Pertanian & Perkebunan Terintegrasi</p>
    </div>
    <div class="content">
      <div class="greeting">Halo <strong>%s</strong>,</div>
      <div class="message">
        Terima kasih telah mendaftar di <strong>AgroLink</strong>. Untuk mengaktifkan akun Anda dan memverifikasi bahwa email ini valid, silakan gunakan kode OTP di bawah ini:
      </div>
      
      <div class="otp-box">
        <div class="otp-code">%s</div>
        <div class="expiry">⏰ Berlaku selama <strong>10 menit</strong></div>
      </div>

      <div class="alert-box">
        🔒 <strong>Penting:</strong> Jangan berikan kode verifikasi ini kepada siapa pun, termasuk pihak yang mengatasnamakan AgroLink.
      </div>
    </div>
    <div class="footer">
      &copy; 2026 AgroLink Platform. Seluruh hak cipta dilindungi undang-undang.<br>
      Email ini dikirimkan secara otomatis, mohon untuk tidak membalas email ini.
    </div>
  </div>
</body>
</html>`, toName, otpCode)

	log.Printf("🔑 [EMAIL OTP GENERATED] To: %s | OTP Code: %s", toEmail, otpCode)

	return s.SendEmail(toEmail, toName, subject, html)
}