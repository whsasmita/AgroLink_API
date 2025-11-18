package services

import (
	"errors"
	"time"

	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(email, password, role, name, phoneNumber string) (*models.User, string, error)
	Login(email, password string) (*models.User, string, error)
	GetProfile(userID string) (*models.User, error)
}

type authService struct {
	UserRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &authService{
		UserRepo: userRepo,
	}
}

func (s *authService) Register(email, password, role, name, phoneNumber string) (*models.User,string, error) {
	existingUser, err := s.UserRepo.FindByEmail(email)

	// Hanya error selain 'record not found' yang harus ditangani sebagai error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	// Jika user ditemukan, maka email sudah digunakan
	if existingUser != nil && err == nil {
		return nil,"", errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil,"", err
	}

	newUser := &models.User{
		Name:        name,
		Email:       email,
		Password:    hashedPassword,
		Role:        role,
		PhoneNumber: &phoneNumber,
		CreatedAt:   time.Now(),
	}

	// Simpan ke DB
	if err := s.UserRepo.Create(newUser); err != nil {
		return nil, "", err
	}

	// Jangan kirim password ke luar service
	newUser.Password = ""
	token, err := config.GenerateToken(newUser.ID.String(), newUser.Email, newUser.Role)
	if err != nil {
		return nil,"",err
	}
	return newUser, token, nil
}

func (s *authService) Login(email, password string) (*models.User, string, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil || user == nil {
		return user, "", errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return user, "", errors.New("invalid email or password")
	}

	token, err := config.GenerateToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return user, "", err
	}

	return user,token, nil
}

func (s *authService) GetProfile(userID string) (*models.User, error) {
	return s.UserRepo.FindByID(userID)
}
