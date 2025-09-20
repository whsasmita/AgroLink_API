package services

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

// [PERBAIKAN] Interface sekarang mengembalikan DTO
type ContractService interface {
	SignContract(contractID string, userID uuid.UUID) (*dto.SignContractResponse, error)
	GenerateContractPDF(contractID string) (*bytes.Buffer, error)
	GetMyContracts(userID uuid.UUID) ([]dto.MyContractResponse, error)
}

type contractService struct {
	contractRepo repositories.ContractRepository
	invoiceRepo  repositories.InvoiceRepository
	projectService ProjectService // <-- Gunakan interface
	deliveryRepo repositories.DeliveryRepository
	db           *gorm.DB
}

// [PERBAIKAN] Konstruktor sekarang menggunakan interface ProjectService
func NewContractService(
	contractRepo repositories.ContractRepository,
	projectService ProjectService,
	invoiceRepo repositories.InvoiceRepository,
	deliveryRepo repositories.DeliveryRepository,
	db *gorm.DB,
) ContractService {
	return &contractService{
		contractRepo: contractRepo,
		invoiceRepo:  invoiceRepo,
		projectService: projectService,
		deliveryRepo: deliveryRepo,
		db:           db,
	}
}

func (s *contractService) GetMyContracts(userID uuid.UUID) ([]dto.MyContractResponse, error) {
	contracts, err := s.contractRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responseDTOs []dto.MyContractResponse
	for _, contract := range contracts {
		dto := dto.MyContractResponse{
			ContractID:   contract.ID,
			ContractType: contract.ContractType,
			Status:       contract.Status,
			OfferedAt:    contract.CreatedAt,
		}
        // Tentukan judul berdasarkan tipe kontrak
		if contract.ContractType == "work" && contract.Project != nil {
			dto.Title = contract.Project.Title
		} else if contract.ContractType == "delivery" && contract.Delivery != nil {
			dto.Title = "Pengiriman: " + contract.Delivery.ItemDescription
		}
		responseDTOs = append(responseDTOs, dto)
	}

	return responseDTOs, nil
}


// [PERBAIKAN] Fungsi SignContract sekarang mengembalikan DTO dan memiliki logika yang benar
func (s *contractService) SignContract(contractID string, userID uuid.UUID) (*dto.SignContractResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	contract, err := s.contractRepo.FindByIDWithDetails(contractID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("contract not found")
	}

	if contract.Status != "pending_signature" {
		tx.Rollback()
		return nil, errors.New("contract is not awaiting signature")
	}

	// Tentukan peran dan validasi
	switch contract.ContractType {
	case "work":
		if contract.WorkerID == nil || *contract.WorkerID != userID {
			tx.Rollback()
			return nil, errors.New("forbidden: you are not the designated worker for this contract")
		}
		contract.SignedBySecondParty = true
	case "delivery":
		if contract.DriverID == nil || *contract.DriverID != userID {
			tx.Rollback()
			return nil, errors.New("forbidden: you are not the designated driver for this contract")
		}
		contract.SignedBySecondParty = true // Menggunakan field yang sama
	default:
		tx.Rollback()
		return nil, errors.New("unknown contract type")
	}

	now := time.Now()
	contract.SignedAt = &now
	contract.Status = "active"

	if err := s.contractRepo.Update(tx, contract); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update contract: %w", err)
	}

	// Buat Invoice HANYA jika ini adalah kontrak pengiriman
	if contract.ContractType == "delivery" {
		delivery, err := s.deliveryRepo.FindByID(contract.DeliveryID.String())
		if err != nil {
			tx.Rollback()
			return nil, errors.New("associated delivery not found")
		}

		totalAmount := 150000.0
		platformFee := totalAmount * 0.05

		newInvoice := &models.Invoice{
			DeliveryID:  contract.DeliveryID,
			FarmerID:    contract.FarmerID,
			Amount:      totalAmount - platformFee,
			PlatformFee: platformFee,
			TotalAmount: totalAmount,
			Status:      "pending",
			DueDate:     time.Now().Add(24 * time.Hour),
		}
		if err := s.invoiceRepo.Create(tx, newInvoice); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice: %w", err)
		}

		delivery.Status = "pending_payment"
		if err := s.deliveryRepo.Update(tx, delivery); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update delivery status: %w", err)
		}
	} else if contract.ContractType == "work" {
		// Panggil fungsi pengecekan untuk proyek kerja
		go s.projectService.CheckAndFinalizeProject(*contract.ProjectID)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
    
    // Buat dan kembalikan DTO
    response := &dto.SignContractResponse{
        ContractID: contract.ID,
        Status:     contract.Status,
        SignedAt:   *contract.SignedAt,
        Message:    "Contract signed successfully.",
    }
    if contract.Project != nil {
        response.ProjectTitle = contract.Project.Title
    }
    if contract.DeliveryID != nil {
        response.DeliveryID = contract.DeliveryID
    }	
	return response, nil
}


func (s *contractService) GenerateContractPDF(contractID string) (*bytes.Buffer, error) {
	contract, err := s.contractRepo.FindByIDWithDetails(contractID)
	if err != nil {
		return nil, errors.New("contract details not found")
	}

	durationDays := contract.Project.EndDate.Sub(contract.Project.StartDate).Hours()/24 + 1

	// paymentRate := contract.Project.PaymentRate

	// log(paymentRate)
	var upahPerHari string
	if contract.Project.PaymentRate != nil {
		// Jika data ada, format dengan benar
		upahPerHari = fmt.Sprintf("Rp %.0f", *contract.Project.PaymentRate)
	} else {
		// Jika data tidak ada, berikan nilai default
		upahPerHari = "[JUMLAH BELUM DITETAPKAN]"
	}

	data := gin.H{
		"Contract":         contract,
		"TanggalPembuatan": contract.CreatedAt.Format("2 January 2006"), // Format tanggal sederhana
		"DurasiHari":       fmt.Sprintf("%.0f", durationDays),           // <-- Tambahkan ini
		"UpahPerHari":      upahPerHari,
	}

	tmpl, err := template.ParseFiles("templates/contract_template.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse html template: %w", err)
	}
	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return nil, fmt.Errorf("could not execute html template: %w", err)
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("could not create PDF generator: %w", err)
	}

	pdfg.AddPage(wkhtmltopdf.NewPageReader(&htmlBuffer))
	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("could not create PDF: %w", err)
	}

	return pdfg.Buffer(), nil
}
