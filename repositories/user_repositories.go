package repositories

import (
	"errors"

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