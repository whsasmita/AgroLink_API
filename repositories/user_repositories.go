package repositories

import (
	"errors"

	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type UserRepository interface {
    FindByEmail(email string) (*models.User, error)
    FindByID(id string) (*models.User, error)
    Create(user *models.User) error
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
        Preload("Expedition").
        Where("id = ?", id).First(&user).Error
    return &user, err
}

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}
