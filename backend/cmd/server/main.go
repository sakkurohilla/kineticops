package main

import (
	"log"

	"github.com/kineticops/backend/internal/config"
	"github.com/kineticops/backend/internal/database"
	"github.com/kineticops/backend/internal/server"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Println("✅ Configuration loaded")

	// Initialize database
	database.ConnectDB(cfg)
	log.Println("✅ Connected to PostgreSQL successfully")
	defer database.CloseDB()

	// Start server
	server.StartServer(cfg)
}
