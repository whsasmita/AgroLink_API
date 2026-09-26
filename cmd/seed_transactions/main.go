package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/seeders"
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
	log.Println("🚀 Menjalankan Clean & Seed Transaksi AgroLink...")
	seeders.SeedNewTransactions(db)
	log.Println("🎉 Selesai seeding transaksi!")
}
