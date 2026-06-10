package main

import (
	"log"

	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/database"
	"github.com/todo-app/backend/routes"
)

func main() {
	log.Println("Starting Todo Application Backend...")

	// 1. Load Configurations
	config.LoadConfig()

	// 2. Connect and Migrate Database
	database.ConnectDatabase()

	// 3. Setup Routes
	router := routes.SetupRouter()

	// 4. Start Server
	port := config.AppConfig.Port
	log.Printf("Server is running on port %s", port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
