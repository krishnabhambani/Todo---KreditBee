package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config is the application configuration contract.
// Consumers depend on this interface, never on the concrete struct,
// which makes test-time substitution straightforward.
type Config interface {
	GetDBHost() string
	GetDBPort() string
	GetDBUser() string
	GetDBPassword() string
	GetDBName() string
	GetJWTSecret() string
	GetPort() string
}

// appConfig is the production implementation of Config.
// It is unexported — callers receive it only as the Config interface.
type appConfig struct {
	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbName     string
	jwtSecret  string
	port       string
}

func (c *appConfig) GetDBHost() string     { return c.dbHost }
func (c *appConfig) GetDBPort() string     { return c.dbPort }
func (c *appConfig) GetDBUser() string     { return c.dbUser }
func (c *appConfig) GetDBPassword() string { return c.dbPassword }
func (c *appConfig) GetDBName() string     { return c.dbName }
func (c *appConfig) GetJWTSecret() string  { return c.jwtSecret }
func (c *appConfig) GetPort() string       { return c.port }

// LoadConfig reads environment variables (from .env file if present) and
// returns an immutable Config value. The caller owns the returned value —
// there is no package-level global.

func LoadConfig() Config {
	// Attempt to load .env — ignore error (env may be injected directly in containers)
	_ = godotenv.Load()

	return &appConfig{
		dbHost:     getEnv("DB_HOST", "localhost"),
		dbPort:     getEnv("DB_PORT", "3306"),
		dbUser:     getEnv("DB_USER", "root"),
		dbPassword: getEnv("DB_PASSWORD", ""),
		dbName:     getEnv("DB_NAME", "tododb"),
		jwtSecret:  getEnv("JWT_SECRET", defaultJWTSecret),
		port:       getEnv("PORT", "8080"),
	}
}

// defaultJWTSecret is exported as an unexported constant so main.go can
// compare against it to emit the production warning without hard-coding the string.
const defaultJWTSecret = "default_secret_key"

// DefaultJWTSecret returns the default (insecure) JWT secret value.
// main.go uses this to detect when a real secret has not been configured.
func DefaultJWTSecret() string { return defaultJWTSecret }

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
