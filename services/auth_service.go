package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(email, password, role, name, phoneNumber string) (*models.User, error)
	VerifyOTP(email, otpCode string) (*models.User, string, error)
	ResendOTP(email string) error
	Login(email, password string) (*models.User, string, error)
	GetProfile(userID string) (*models.User, error)
}

type authService struct {
	UserRepo     repositories.UserRepository
	OTPRepo      repositories.OTPRepository
	EmailService EmailService
}

func NewAuthService(userRepo repositories.UserRepository, otpRepo repositories.OTPRepository, emailService EmailService) AuthService {
	return &authService{
		UserRepo:     userRepo,
		OTPRepo:      otpRepo,
		EmailService: emailService,
	}
}

func generateOTPCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "123456"
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func (s *authService) Register(email, password, role, name, phoneNumber string) (*models.User, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	existingUser, err := s.UserRepo.FindByEmail(cleanEmail)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var user *models.User

	if existingUser != nil {
		if existingUser.EmailVerified {
			return nil, errors.New("email sudah terdaftar dan terverifikasi, silakan login")
		}
		// Jika email belum diverifikasi, perbarui data akun
		existingUser.Name = name
		existingUser.Password = hashedPassword
		existingUser.Role = role
		existingUser.PhoneNumber = &phoneNumber
		existingUser.UpdatedAt = time.Now()

		if err := s.UserRepo.UpdateProfile(existingUser); err != nil {
			return nil, err
		}
		user = existingUser
	} else {
		newUser := &models.User{
			Name:          name,
			Email:         cleanEmail,
			Password:      hashedPassword,
			Role:          role,
			PhoneNumber:   &phoneNumber,
			IsActive:      true,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := s.UserRepo.Create(newUser); err != nil {
			return nil, err
		}
		user = newUser
	}

	// 1. Nonaktifkan OTP lama
	_ = s.OTPRepo.InvalidateExistingOTPs(cleanEmail)

	// 2. Buat kode OTP 6 digit baru
	otpCode := generateOTPCode()
	otp := &models.EmailOTP{
		Email:     cleanEmail,
		OTPCode:   otpCode,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.OTPRepo.CreateOTP(otp); err != nil {
		return nil, fmt.Errorf("gagal membuat kode verifikasi: %w", err)
	}

	// 3. Kirim email OTP
	go func() {
		_ = s.EmailService.SendOTPEmail(cleanEmail, name, otpCode)
	}()

	user.Password = ""
	return user, nil
}

func (s *authService) VerifyOTP(email, otpCode string) (*models.User, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanCode := strings.TrimSpace(otpCode)

	if cleanEmail == "" || cleanCode == "" {
		return nil, "", errors.New("email dan kode OTP wajib diisi")
	}

	// 1. Validasi OTP
	otp, err := s.OTPRepo.FindValidOTP(cleanEmail, cleanCode)
	if err != nil {
		return nil, "", err
	}

	// 2. Ambil user
	user, err := s.UserRepo.FindByEmail(cleanEmail)
	if err != nil || user == nil {
		return nil, "", errors.New("pengguna tidak ditemukan")
	}

	// 3. Verifikasi akun
	user.EmailVerified = true
	user.IsActive = true
	if err := s.UserRepo.UpdateProfile(user); err != nil {
		return nil, "", err
	}

	// 4. Tandai OTP sudah digunakan
	_ = s.OTPRepo.MarkAsUsed(otp.ID)

	// 5. Generate JWT Token
	token, err := config.GenerateToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat token autentikasi: %w", err)
	}

	user.Password = ""
	return user, token, nil
}

func (s *authService) ResendOTP(email string) error {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return errors.New("email wajib diisi")
	}

	user, err := s.UserRepo.FindByEmail(cleanEmail)
	if err != nil || user == nil {
		return errors.New("email tidak terdaftar")
	}

	if user.EmailVerified {
		return errors.New("email sudah terverifikasi, silakan langsung login")
	}

	// Invalidate OTP lama
	_ = s.OTPRepo.InvalidateExistingOTPs(cleanEmail)

	// Buat OTP baru
	otpCode := generateOTPCode()
	otp := &models.EmailOTP{
		Email:     cleanEmail,
		OTPCode:   otpCode,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}

	if err := s.OTPRepo.CreateOTP(otp); err != nil {
		return fmt.Errorf("gagal membuat kode verifikasi baru: %w", err)
	}

	// Kirim email
	go func() {
		_ = s.EmailService.SendOTPEmail(cleanEmail, user.Name, otpCode)
	}()

	return nil
}

func (s *authService) Login(email, password string) (*models.User, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	user, err := s.UserRepo.FindByEmail(cleanEmail)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := config.GenerateToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return nil, "", err
	}

	user.Password = ""
	return user, token, nil
}

func (s *authService) GetProfile(userID string) (*models.User, error) {
	return s.UserRepo.FindByID(userID)
}
