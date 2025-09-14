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
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type ContractService interface {
	SignContract(contractID string, workerID uuid.UUID) (*models.Contract, error)
	GenerateContractPDF(contractID string) (*bytes.Buffer, error)
}

type contractService struct {
	contractRepo   repositories.ContractRepository
	projectService ProjectService
}

func NewContractService(repo repositories.ContractRepository, projectService ProjectService) ContractService {
	return &contractService{
		contractRepo:   repo,
		projectService: projectService,
	}
}

func (s *contractService) SignContract(contractID string, workerID uuid.UUID) (*models.Contract, error) {
	contract, err := s.contractRepo.FindByID(contractID)
	if err != nil {
		return nil, fmt.Errorf("contract not found")
	}
	if contract.WorkerID != workerID {
		return nil, fmt.Errorf("forbidden: you are not authorized to sign this contract")
	}
	if contract.Status != "pending_signature" {
		return nil, fmt.Errorf("contract is no longer pending signature")
	}

	contract.SignedByWorker = true
	contract.Status = "active"
	now := time.Now()
	contract.SignedAt = &now

	if err := s.contractRepo.Update(contract); err != nil {
		return nil, fmt.Errorf("failed to update contract status: %w", err)
	}

	go s.projectService.CheckAndFinalizeProject(contract.ProjectID)
	return contract, nil
}

func (s *contractService) GenerateContractPDF(contractID string) (*bytes.Buffer, error) {
	contract, err := s.contractRepo.FindByIDWithDetails(contractID)
	if err != nil {
		return nil, errors.New("contract details not found")
	}

	data := gin.H{
		"Contract":         contract,
		"FormattedContent": template.HTML(contract.Content),
		"TanggalPembuatan":   contract.CreatedAt.Format("dddd, 2 January 2006"),
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