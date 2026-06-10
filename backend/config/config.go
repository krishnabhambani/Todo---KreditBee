package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPassword string
	DBName    string
	JWTSecret string
	Port      string
}

var AppConfig *Config

func LoadConfig() {
	// Attempt to load .env file
	_ = godotenv.Load() // ignore error, might be loaded from environment directly

	AppConfig = &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "tododb"),
		JWTSecret:  getEnv("JWT_SECRET", "default_secret_key"),
		Port:       getEnv("PORT", "8080"),
	}

	if AppConfig.JWTSecret == "default_secret_key" {
		log.Println("WARNING: Using default JWT secret key. Please set JWT_SECRET in production.")
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
