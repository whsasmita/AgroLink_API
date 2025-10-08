package services

import (
	"errors"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
	"gorm.io/gorm"
)

// UserService mendefinisikan tugas-tugas yang terkait dengan manajemen pengguna.
type UserService interface {
	CreateUser(input dto.RegisterRequest) (*models.User, error)
	GetUserProfile(userID string) (*models.User, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService adalah constructor untuk userService.
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
		
	}
}

// CreateUser menangani logika bisnis untuk pendaftaran pengguna baru.
func (s *userService) CreateUser(input dto.RegisterRequest) (*models.User, error) {
	// 1. Cek apakah email sudah terdaftar
	existingUser, err := s.userRepo.FindByEmail(input.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // Error database selain "tidak ditemukan"
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// 2. Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// 3. Buat objek User utama
	newUser := &models.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    hashedPassword,
		Role:        input.Role,
		PhoneNumber: &input.PhoneNumber,
	}

	// 4. Buat sub-profil berdasarkan peran (Farmer, Worker, atau Driver)
	// GORM akan secara otomatis membuat record di tabel terkait saat menyimpan newUser
	// karena adanya relasi `has one` di model User.
	switch input.Role {
	case "farmer":
		newUser.Farmer = &models.Farmer{}
	case "worker":
		newUser.Worker = &models.Worker{}
	case "driver":
		newUser.Driver = &models.Driver{}
	case "general":
    	break
	default:
		return nil, errors.New("invalid user role specified")
	}

	// 5. Simpan ke database
	if err := s.userRepo.Create(newUser); err != nil {
		return nil, err
	}

	// Kosongkan password sebelum mengembalikan data
	newUser.Password = ""
	return newUser, nil
}

// GetUserProfile mengambil data profil lengkap pengguna berdasarkan ID.
func (s *userService) GetUserProfile(userID string) (*models.User, error) {
	// Gunakan fungsi repository yang melakukan Preload untuk data profil
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Kosongkan password demi keamanan
	user.Password = ""
	return user, nil
}