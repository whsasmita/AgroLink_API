package config

import (
	"errors"
	"log" // Untuk menampilkan pesan saat fallback
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getAndValidateEnv("JWT_SECRET"))

// Custom claims
type JWTClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

// GenerateToken creates a signed JWT token
func GenerateToken(userID, email, role string) (string, error) {
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 jam masa berlaku
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		Email: email,
		Role:  role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken parses and verifies a JWT token
func ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// PENTING: Validasi bahwa metode signing adalah yang Anda harapkan (misal: HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// Helper untuk mendapatkan env. Akan panic jika kosong.
func getAndValidateEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		// Di lingkungan pengembangan, kita bisa beri fallback agar tidak crash.
		// Di produksi, ini seharusnya menyebabkan panic.
		if os.Getenv("APP_ENV") != "production" {
			log.Printf("PERINGATAN: JWT_SECRET tidak diatur. Menggunakan secret key default yang tidak aman.")
			return "my-super-secret-for-development" // Secret default HANYA untuk development
		}
		// Di produksi, aplikasi harus crash jika secret tidak ada.
		panic("FATAL: JWT_SECRET environment variable is not set.")
	}
	return val
}