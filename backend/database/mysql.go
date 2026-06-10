package database

import (
	"fmt"
	"log"

	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	cfg := config.AppConfig
	
	// Create DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL database: %v", err)
	}

	// 1. Clean up old tables and columns from the previous categorization layout if they exist
	_ = database.Exec("ALTER TABLE todos DROP FOREIGN KEY fk_todos_groups").Error
	_ = database.Exec("ALTER TABLE todos DROP COLUMN group_id").Error
	_ = database.Exec("DROP TABLE IF EXISTS todo_groups").Error

	// 2. Auto migrate tables
	err = database.AutoMigrate(&models.User{}, &models.Todo{}, &models.GroupShare{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	DB = database
	log.Println("MySQL Database connected and auto-migrated successfully")
}
