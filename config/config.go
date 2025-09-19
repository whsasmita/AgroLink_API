package config

import (
	"os"
	"strconv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Upload   UploadConfig
	Platform PlatformConfig
}

type AppConfig struct {
	Env     string
	Port    string
	APP_URL any
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Charset  string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

type UploadConfig struct {
	Path        string
	MaxFileSize int64
}

type PlatformConfig struct {
	FeePercent float64
}

var AppConfig_ *Config

func LoadConfig() *Config {
	AppConfig_ = &Config{
		App: AppConfig{
			Env:     getEnvWithDefault("APP_ENV", "development"),
			Port:    getEnvWithDefault("PORT", "8080"),
			APP_URL: getEnvWithDefault("APP_URL", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvWithDefault("DB_PORT", "3306"),
			User:     getEnvWithDefault("DB_USER", "root"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnvWithDefault("DB_NAME", "agri_platform"),
			Charset:  getEnvWithDefault("DB_CHARSET", "utf8mb4"),
		},
		JWT: JWTConfig{
			Secret:      getEnvWithDefault("JWT_SECRET", "your-secret-key"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		Email: EmailConfig{
			SMTPHost:     getEnvWithDefault("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    getEnvWithDefault("FROM_EMAIL", "noreply@agriplatform.com"),
		},
		Upload: UploadConfig{
			Path:        getEnvWithDefault("UPLOAD_PATH", "./uploads"),
			MaxFileSize: int64(getEnvAsInt("MAX_FILE_SIZE", 5242880)), // 5MB default
		},
		Platform: PlatformConfig{
			FeePercent: getEnvAsFloat("PLATFORM_FEE_PERCENT", 5.0),
		},
	}

	return AppConfig_
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
