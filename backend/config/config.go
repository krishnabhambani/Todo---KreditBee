package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration organized by domain.
type Config struct {
	db     *databaseConfig
	jwt    *jwtConfig
	server *serverConfig
}

// ─────────────────────────────────────────────────────────────
// Database Configuration
// ─────────────────────────────────────────────────────────────

type databaseConfig struct {
	host     string
	port     string
	user     string
	password string
	name     string
}

func (c *Config) GetDBHost() string     { return c.db.host }
func (c *Config) GetDBPort() string     { return c.db.port }
func (c *Config) GetDBUser() string     { return c.db.user }
func (c *Config) GetDBPassword() string { return c.db.password }
func (c *Config) GetDBName() string     { return c.db.name }

// ─────────────────────────────────────────────────────────────
// JWT Configuration
// ─────────────────────────────────────────────────────────────

type jwtConfig struct {
	secret string
}

const jwtDefaultSecret = "default_secret_key"

func (c *Config) GetJWTSecret() string { return c.jwt.secret }

// JWTDefaultSecret returns the default (insecure) JWT secret value.
// Used by main.go to detect when a real secret has not been configured.
func JWTDefaultSecret() string { return jwtDefaultSecret }

// ─────────────────────────────────────────────────────────────
// Server Configuration
// ─────────────────────────────────────────────────────────────

type serverConfig struct {
	port string
}

func (c *Config) GetPort() string { return c.server.port }

// ─────────────────────────────────────────────────────────────
// Config Initialization
// ─────────────────────────────────────────────────────────────

// LoadConfig reads environment variables (from .env file if present) and
// returns a Config value with all domain-specific configuration loaded.
func LoadConfig() *Config {
	// Attempt to load .env — ignore error (env may be injected directly in containers)
	_ = godotenv.Load()

	return &Config{
		db: &databaseConfig{
			host:     getEnv("DB_HOST", "localhost"),
			port:     getEnv("DB_PORT", "3306"),
			user:     getEnv("DB_USER", "root"),
			password: getEnv("DB_PASSWORD", ""),
			name:     getEnv("DB_NAME", "tododb"),
		},
		jwt: &jwtConfig{
			secret: getEnv("JWT_SECRET", jwtDefaultSecret),
		},
		server: &serverConfig{
			port: getEnv("PORT", "8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
