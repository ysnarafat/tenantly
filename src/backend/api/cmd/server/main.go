package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/config"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env", "src/backend/api/.env"); err != nil {
		// Try loading from the api directory if not found in current directory
		if err := godotenv.Load("src/backend/api/.env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	// Load configuration
	cfg, err := config.Load()

	if err != nil {
		log.Fatal("Failed to load configs:", err)
	}

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}

	// Encrypt any legacy plaintext tenant NIDs left over from before at-rest
	// protection was introduced (idempotent, safe to run on every startup).
	if err := repositories.BackfillTenantNID(db, cfg.NIDProtector); err != nil {
		log.Fatal("Failed to back-fill tenant NID: ", err)
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize and start server
	srv := server.New(cfg, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Tenantly API server on port %s", port)
	if err := srv.Start(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
