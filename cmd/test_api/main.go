package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/services"
)

func main() {
	_ = godotenv.Load()
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "127.0.0.1")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "3307")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "root")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "agrolink123")
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "db_agrolink")
	}

	db := config.ConnectDatabase()

	payoutRepo := repositories.NewPayoutRepository(db)
	userRepo := repositories.NewUserRepository(db)
	userVerificationRepo := repositories.NewUserVerificationRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	deliveryRepo := repositories.NewDeliveryRepository(db)
	ecommPaymentRepo := repositories.NewECommercePaymentRepository(db)
	orderRepo := repositories.NewOrderRepository(db)

	adminService := services.NewAdminService(
		payoutRepo, userRepo, userVerificationRepo, transactionRepo,
		projectRepo, deliveryRepo, ecommPaymentRepo, orderRepo, db,
	)

	fmt.Println("=== 1. API: GetDashboardStats ===")
	stats, err := adminService.GetDashboardStats()
	if err != nil {
		log.Fatalf("Error GetDashboardStats: %v", err)
	}

	fmt.Printf("Total Transactions: %d (Sukses: %d, Gagal: %d)\n",
		stats.FinancialSummary.TotalTransactions,
		stats.FinancialSummary.SuccessfulTransactions,
		stats.FinancialSummary.FailedTransactions)
	fmt.Printf("Total GMV: Rp %.0f, Net Profit: Rp %.0f, Gross Profit: Rp %.0f, Gateway Fee: Rp %.0f, Mitra Share: Rp %.0f\n",
		stats.FinancialSummary.TotalGMV,
		stats.FinancialSummary.TotalNetProfit,
		stats.FinancialSummary.TotalGrossProfit,
		stats.FinancialSummary.TotalGatewayFee,
		stats.FinancialSummary.TotalMitraShare)
	fmt.Printf("Phase 1: %d tx (Net: Rp %.0f), Phase 2: %d tx (Net: Rp %.0f)\n",
		stats.FinancialSummary.Phase1Transactions,
		stats.FinancialSummary.Phase1NetProfit,
		stats.FinancialSummary.Phase2Transactions,
		stats.FinancialSummary.Phase2NetProfit)

	fmt.Println("\nStatus Breakdown dari API DashboardStats:")
	for _, sb := range stats.StatusBreakdown {
		fmt.Printf(" - Layanan: %-16s | Sektor: %-15s | Sukses: %3d | Gagal: %2d | Total: %3d | Gross: Rp %9.0f | Net: Rp %9.0f\n",
			sb.Layanan, sb.Sektor, sb.Sukses, sb.Gagal, sb.Total, sb.GrossProfit, sb.NetProfit)
	}

	fmt.Println("\n=== 2. API: GetCombinedTransactions (List API) ===")
	txRes, err := adminService.GetCombinedTransactions(1, 10, "", "", "", "", "", "")
	if err != nil {
		log.Fatalf("Error GetCombinedTransactions: %v", err)
	}
	fmt.Printf("Pagination Total Items: %d, Total Pages: %d, Current Page: %d\n", txRes.TotalItems, txRes.TotalPages, txRes.CurrentPage)
}
