package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// AssignmentRepository mendefinisikan interface untuk operasi database penugasan proyek.
type AssignmentRepository interface {
	// Create menggunakan 'tx *gorm.DB' agar bisa menjadi bagian dari sebuah transaksi.
	Create(tx *gorm.DB, assignment *models.ProjectAssignment) (error)
	FindAllByProjectID(projectID string) ([]models.ProjectAssignment, error)
	FindAllByWorkerID(workerID string) ([]models.ProjectAssignment, error)
}

type assignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository membuat instance baru dari AssignmentRepository.
func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

// Create menyimpan record penugasan baru.
// Fungsi ini menerima 'tx' agar bisa dijalankan di dalam sebuah transaksi database
// yang diinisialisasi oleh service.
func (r *assignmentRepository) Create(tx *gorm.DB, assignment *models.ProjectAssignment) error {
	return tx.Create(assignment).Error
}

// FindAllByProjectID mencari semua penugasan yang terkait dengan satu proyek.
// Berguna untuk melihat daftar pekerja di sebuah proyek.
func (r *assignmentRepository) FindAllByProjectID(projectID string) ([]models.ProjectAssignment, error) {
	var assignments []models.ProjectAssignment
	// Preload Worker dan User di dalamnya untuk bisa menampilkan nama pekerja
	err := r.db.Preload("Worker.User").Where("project_id = ?", projectID).Find(&assignments).Error
	return assignments, err
}

// FindAllByWorkerID mencari semua penugasan yang dimiliki oleh seorang pekerja.
// Berguna untuk halaman riwayat pekerjaan seorang pekerja.
func (r *assignmentRepository) FindAllByWorkerID(workerID string) ([]models.ProjectAssignment, error) {
	var assignments []models.ProjectAssignment
	// Preload Project untuk bisa menampilkan judul proyek
	err := r.db.Preload("Project").Where("worker_id = ?", workerID).Find(&assignments).Error
	return assignments, err
}