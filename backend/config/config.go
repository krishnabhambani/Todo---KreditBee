package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config defines the interface for application configuration.
type Config interface {
	DB() DatabaseConfig
	JWT() JWTConfig
	Server() ServerConfig
}

// ─────────────────────────────────────────────────────────────
// Database Configuration
// ─────────────────────────────────────────────────────────────

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// ─────────────────────────────────────────────────────────────
// JWT Configuration
// ─────────────────────────────────────────────────────────────

type JWTConfig struct {
	Secret string
}

const jwtDefaultSecret = "default_secret_key"

// JWTDefaultSecret returns the default (insecure) JWT secret value.
func JWTDefaultSecret() string { return jwtDefaultSecret }

// ─────────────────────────────────────────────────────────────
// Server Configuration
// ─────────────────────────────────────────────────────────────

type ServerConfig struct {
	Port string
}

// ─────────────────────────────────────────────────────────────
// Config Implementation
// ─────────────────────────────────────────────────────────────

type appConfig struct {
	db     DatabaseConfig
	jwt    JWTConfig
	server ServerConfig
}

func (c *appConfig) DB() DatabaseConfig   { return c.db }
func (c *appConfig) JWT() JWTConfig       { return c.jwt }
func (c *appConfig) Server() ServerConfig { return c.server }

// LoadConfig reads environment variables and returns a Config interface implementation.
func LoadConfig() Config {
	_ = godotenv.Load()

	return &appConfig{
		db: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "tododb"),
		},
		jwt: JWTConfig{
			Secret: getEnv("JWT_SECRET", jwtDefaultSecret),
		},
		server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
