package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	DatabaseURL    string
	JWTSecret      string
	JWTExpiration  time.Duration
	SMSProvider    SMSConfig
	EmailProvider  EmailConfig
	Environment    string
	AllowedOrigins []string
	TrustedProxies []string
	CookieDomain   string
	CookieSecure   bool
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

	// Load and validate JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		env := os.Getenv("ENVIRONMENT")
		if env == "production" {
			return nil, fmt.Errorf("CRITICAL: JWT_SECRET environment variable must be set in production")
		}
		// Development: generate secure random value
		fmt.Println("⚠️  WARNING: JWT_SECRET not set. Generating cryptographically secure random value for development.")
		jwtSecret = generateSecureRandom(32)
		fmt.Printf("Generated JWT_SECRET: %s\n", jwtSecret)
	}

	// Validate JWT secret strength
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("CRITICAL: JWT_SECRET must be at least 32 characters (current: %d). Use 'openssl rand -base64 32' to generate", len(jwtSecret))
	}

	jwtExpiration, _ := time.ParseDuration(getEnv("JWT_EXPIRATION", "8h"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	environment := getEnv("ENVIRONMENT", "development")

	// Load and validate database URL — same no-insecure-default-in-prod pattern as JWT_SECRET.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		if environment == "production" {
			return nil, fmt.Errorf("CRITICAL: DATABASE_URL environment variable must be set in production")
		}
		fmt.Println("⚠️  WARNING: DATABASE_URL not set. Using local development default.")
		databaseURL = "postgres://postgres:password@localhost:5432/tenantly?sslmode=disable"
	}

	// CORS allowed origins — required in production (no wildcard fallback), a
	// sane localhost default in development.
	var allowedOrigins []string
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			if o := strings.TrimSpace(origin); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}
	if len(allowedOrigins) == 0 {
		if environment == "production" {
			return nil, fmt.Errorf("CRITICAL: CORS_ALLOWED_ORIGINS environment variable must be set in production")
		}
		fmt.Println("⚠️  WARNING: CORS_ALLOWED_ORIGINS not set. Defaulting to http://localhost:4200 for development.")
		allowedOrigins = []string{"http://localhost:4200"}
	}

	// Trusted proxies for Gin's ClientIP() resolution — empty/unset means Gin
	// ignores X-Forwarded-For entirely and uses the real socket address (safe
	// default). Set explicitly only if a real reverse proxy forwards client IPs.
	var trustedProxies []string
	if raw := os.Getenv("TRUSTED_PROXIES"); raw != "" {
		for _, proxy := range strings.Split(raw, ",") {
			if p := strings.TrimSpace(proxy); p != "" {
				trustedProxies = append(trustedProxies, p)
			}
		}
	}

	cookieSecure := environment == "production"
	if raw := os.Getenv("COOKIE_SECURE"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			cookieSecure = parsed
		}
	}

	config := &Config{
		DatabaseURL:   databaseURL,
		JWTSecret:     jwtSecret,
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
		Environment:    environment,
		AllowedOrigins: allowedOrigins,
		TrustedProxies: trustedProxies,
		CookieDomain:   getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:   cookieSecure,
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

// generateSecureRandom creates a cryptographically secure random string
func generateSecureRandom(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	return base64.StdEncoding.EncodeToString(b)
}
