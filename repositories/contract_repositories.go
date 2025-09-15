package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// ContractRepository mendefinisikan interface untuk operasi database kontrak.
type ContractRepository interface {
	// Create menggunakan 'tx *gorm.DB' agar bisa menjadi bagian dari sebuah transaksi.
	Create(tx *gorm.DB, contract *models.Contract) error
	FindByID(id string) (*models.Contract, error)
	Update(contract *models.Contract) error
	FindByIDWithDetails(id string) (*models.Contract, error)
}

type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository membuat instance baru dari ContractRepository.
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &contractRepository{db: db}
}

// Create menyimpan record kontrak baru.
// Fungsi ini menerima 'tx' agar bisa dijalankan di dalam sebuah transaksi database
// yang diinisialisasi oleh service.
func (r *contractRepository) Create(tx *gorm.DB, contract *models.Contract) error {
	return tx.Create(contract).Error
}

// FindByID mencari satu kontrak berdasarkan ID-nya.
// Menggunakan Preload untuk mengambil data proyek yang berelasi.
func (r *contractRepository) FindByID(id string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.Preload("Project").Where("id = ?", id).First(&contract).Error
	return &contract, err
}

// Update memperbarui data kontrak di database.
// Menggunakan Save() akan memperbarui semua kolom.
func (r *contractRepository) Update(contract *models.Contract) error {
	return r.db.Save(contract).Error
}

func (r *contractRepository) FindByIDWithDetails(id string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.
		Preload("Project").
		Preload("Farmer.User").
		Preload("Worker.User").
		Where("id = ?", id).
		First(&contract).Error
	return &contract, err
}
