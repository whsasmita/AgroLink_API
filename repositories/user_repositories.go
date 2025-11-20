package repositories

import (
	"errors"
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
    FindByEmail(email string) (*models.User, error)
    FindByID(id string) (*models.User, error)
    Create(user *models.User) error
    UpdateProfile(user *models.User) error
    CreateOrUpdateFarmer(farmer *models.Farmer) error
	CreateOrUpdateWorker(worker *models.Worker) error
	CreateOrUpdateDriver(driver *models.Driver) error
	CountNewUsers(since time.Time) (int64, error)
    GetDailyUserTrend(since time.Time) ([]dto.DailyDataPoint, error)
	FindAllUsers(page, limit int, search string, roleFilter string) ([]models.User, int64, error)
	GetUserRoleStats() (*dto.UserRoleStatsResponse, error)
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
    var user models.User
    err := r.db.Preload("Farmer").
        Preload("Worker").
        Preload("Driver").
        Where("id = ?", id).First(&user).Error
    return &user, err
}

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

// UpdateProfile memperbarui kolom spesifik dari profil pengguna.
func (r *userRepository) UpdateProfile(user *models.User) error {
	// Menggunakan Model(&user) untuk menargetkan record berdasarkan Primary Key dari objek user.
	// Menggunakan Select() untuk secara eksplisit menentukan kolom mana yang diizinkan untuk diperbarui.
	// Ini adalah praktik keamanan yang baik untuk mencegah pembaruan massal yang tidak diinginkan.
	// Menggunakan Updates() untuk menerapkan perubahan dari objek user ke database.
	return r.db.Model(&user).
		Select("Name", "PhoneNumber", "ProfilePicture").
		Updates(user).Error
}

// CreateOrUpdateFarmer akan menyimpan data Farmer. Jika sudah ada, akan diperbarui.
func (r *userRepository) CreateOrUpdateFarmer(farmer *models.Farmer) error {
	// `OnConflict` akan melakukan UPDATE pada semua kolom jika ada konflik di primary key (UserID).
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(farmer).Error
}

// CreateOrUpdateWorker akan menyimpan data Worker.
func (r *userRepository) CreateOrUpdateWorker(worker *models.Worker) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(worker).Error
}

// CreateOrUpdateDriver akan menyimpan data Driver.
func (r *userRepository) CreateOrUpdateDriver(driver *models.Driver) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(driver).Error
}

func (r *userRepository) CountNewUsers(since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("created_at > ?", since).
		Count(&count).Error
	return count, err
}

// [FUNGSI BARU]
// GetDailyUserTrend menghitung jumlah pengguna baru per hari untuk data grafik.
func (r *userRepository) GetDailyUserTrend(since time.Time) ([]dto.DailyDataPoint, error) {
	var results []dto.DailyDataPoint
	
	err := r.db.Model(&models.User{}).
		Select("DATE(created_at) as date, COUNT(*) as value"). // Mengelompokkan berdasarkan tanggal
		Where("created_at > ?", since).
		Group("DATE(created_at)").
		Order("date ASC"). // Penting untuk urutan grafik
		Scan(&results).Error // Scan hasil query ke struct DTO

	return results, err
}

func (r *userRepository) FindAllUsers(page, limit int, search string, roleFilter string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	offset := (page - 1) * limit

	// Mulai query: Ambil semua user KECUALI admin
	query := r.db.Model(&models.User{}).Where("role != ?", "admin")

	// Terapkan filter pencarian (Nama atau Email)
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", searchTerm, searchTerm)
	}

	// Terapkan filter role jika ada (misal: hanya ingin lihat 'farmer')
	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}

	// Hitung total data (untuk pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Ambil data dengan limit dan offset
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) GetUserRoleStats() (*dto.UserRoleStatsResponse, error) {
	var total int64
	if err := r.db.
		Model(&models.User{}).
		Where("role != ?", "admin").
		Count(&total).Error; err != nil {
		return nil, err
	}

	// hitung masing-masing role
	var totalGeneral int64
	if err := r.db.
		Model(&models.User{}).
		Where("role = ?", "general").
		Count(&totalGeneral).Error; err != nil {
		return nil, err
	}

	var totalFarmer int64
	if err := r.db.
		Model(&models.User{}).
		Where("role = ?", "farmer").
		Count(&totalFarmer).Error; err != nil {
		return nil, err
	}

	var totalWorker int64
	if err := r.db.
		Model(&models.User{}).
		Where("role = ?", "worker").
		Count(&totalWorker).Error; err != nil {
		return nil, err
	}

	return &dto.UserRoleStatsResponse{
		TotalUsers:   total,
		TotalGeneral: totalGeneral,
		TotalFarmer:  totalFarmer,
		TotalWorker:  totalWorker,
	}, nil
}
