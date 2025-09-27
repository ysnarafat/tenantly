package config

import (
	"os"
	"strconv"
	"time"
)

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

func Load() *Config {
	jwtExpiration, _ := time.ParseDuration(getEnv("JWT_EXPIRATION", "8h"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
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
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
