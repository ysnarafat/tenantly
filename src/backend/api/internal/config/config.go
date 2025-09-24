package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// loadEnv attempts to load environment variables from .env files
func loadEnv() error {
	envFiles := []string{
		".env", // Current directory

		"../../../.env", // Project root
		"../../.env",    // One level up
		filepath.Join(os.Getenv("HOME"), ".tenantly.env"), // Home directory
	}

	var loaded bool
	var lastErr error
	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			loaded = true
			fmt.Printf("Loaded environment from: %s\n", file)
			break
		} else {
			lastErr = err
		}
	}

	if !loaded && os.Getenv("ENVIRONMENT") != "production" {
		fmt.Println("Warning: .env file not found, using default values")
		return lastErr
	}
	return nil
}

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	JWTExpiration time.Duration
	SMSProvider   SMSConfig
	EmailProvider EmailConfig
	Environment   string
}

type SMSConfig struct {
	Provider string
	APIKey   string
	APIUrl   string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

func Load() (*Config, error) {
	if err := loadEnv(); err != nil && os.Getenv("ENVIRONMENT") != "production" {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	jwtExpiration, _ := time.ParseDuration(getEnv("JWT_EXPIRATION", "8h"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	config := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/tenantly?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration: jwtExpiration,
		SMSProvider: SMSConfig{
			Provider: getEnv("SMS_PROVIDER", "ssl_wireless"),
			APIKey:   getEnv("SMS_API_KEY", ""),
			APIUrl:   getEnv("SMS_API_URL", ""),
		},
		EmailProvider: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     smtpPort,
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@tenantly.com"),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	// Get from environment or .env file (already loaded by godotenv)
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Log when falling back to default in non-production
	if os.Getenv("ENVIRONMENT") != "production" {
		fmt.Printf("Config: using default value for %s\n", key)
	}

	return defaultValue
}
