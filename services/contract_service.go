package services

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type ContractService interface {
	SignContract(contractID string, workerID uuid.UUID) (*models.Contract, error)
	GenerateContractPDF(contractID string) (*bytes.Buffer, error)
}

type contractService struct {
	contractRepo repositories.ContractRepository
}

func NewContractService(repo repositories.ContractRepository) ContractService {
	return &contractService{contractRepo: repo}
}

func (s *contractService) SignContract(contractID string, workerID uuid.UUID) (*models.Contract, error) {
	// 1. Ambil kontrak dari database
	contract, err := s.contractRepo.FindByID(contractID)
	if err != nil {
		return nil, errors.New("contract not found")
	}

	// 2. Validasi Keamanan: Pastikan yang mengakses adalah pekerja yang benar.
	if contract.WorkerID != workerID {
		return nil, errors.New("forbidden: you are not authorized to sign this contract")
	}

	// 3. Validasi Status: Pastikan kontrak masih menunggu tanda tangan.
	if contract.Status != "pending_signature" {
		return nil, errors.New("contract is no longer pending signature")
	}

	// 4. Update status kontrak
	contract.SignedByWorker = true
	contract.Status = "active" // Kontrak aktif setelah kedua belah pihak setuju
	now := time.Now()
	contract.SignedAt = &now

	if err := s.contractRepo.Update(contract); err != nil {
		return nil, fmt.Errorf("failed to update contract status: %w", err)
	}

	return contract, nil
}

func (s *contractService) GenerateContractPDF(contractID string) (*bytes.Buffer, error) {
    // 1. Ambil data kontrak lengkap (termasuk relasi project, worker, farmer)
    // Anda mungkin perlu membuat fungsi baru di repository: FindByIDWithDetails(id)
    contract, err := s.contractRepo.FindByIDWithDetails(contractID) // Asumsi fungsi ini ada
    if err != nil {
        return nil, errors.New("contract details not found")
    }

    // 2. Buat objek PDF
    pdf := fpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "B", 16)
    
    // 3. Isi Konten PDF (Contoh Sederhana)
    pdf.Cell(40, 10, "KONTRAK KERJA")
    pdf.Ln(12) // Pindah baris

    pdf.SetFont("Arial", "", 12)
    pdf.Cell(40, 10, fmt.Sprintf("Proyek: %s", contract.Project.Title))
    pdf.Ln(6)
    pdf.Cell(40, 10, fmt.Sprintf("Petani: %s", contract.Farmer.User.Name))
    pdf.Ln(6)
    pdf.Cell(40, 10, fmt.Sprintf("Pekerja: %s", contract.Worker.User.Name))
    pdf.Ln(10)

    pdf.SetFont("Arial", "I", 10)
    pdf.MultiCell(0, 5, contract.Content, "", "", false) // Isi kontrak utama
    pdf.Ln(10)

    // 4. Tanda Tangan Sederhana
    pdf.SetFont("Arial", "", 12)
    if contract.SignedByFarmer {
        pdf.Cell(40, 10, fmt.Sprintf("Disetujui oleh Petani: %s pada %s", contract.Farmer.User.Name, contract.UpdatedAt.Format("02 Jan 2006")))
        pdf.Ln(6)
    }
    if contract.SignedByWorker {
        pdf.Cell(40, 10, fmt.Sprintf("Disetujui oleh Pekerja: %s pada %s", contract.Worker.User.Name, contract.SignedAt.Format("02 Jan 2006")))
        pdf.Ln(6)
    }

    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, fmt.Errorf("failed to generate PDF buffer: %w", err)
    }
    return &buf, nil
}
