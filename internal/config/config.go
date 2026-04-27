package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server         ServerConfig
	Security       SecurityConfig
	Database       DatabaseConfig
	Logging        LoggingConfig
	FaceRecognition FaceRecognitionConfig
	Environment    string // development or production
}

type ServerConfig struct {
	Host string
	Port string
	Mode string // debug, release
}

type SecurityConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	AllowedIPs    []string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

type LoggingConfig struct {
	Level string
	File  string
}

type FaceRecognitionConfig struct {
	Enabled        bool
	ModelsPath     string
	MatchThreshold float64
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Security: SecurityConfig{
			JWTSecret:     getEnv("JWT_SECRET", "change-this-secret-key-in-production"),
			JWTExpiration: 24 * time.Hour,
			AllowedIPs:    parseAllowedIPs(getEnv("ALLOWED_IPS", "127.0.0.1/32,192.168.0.0/16")),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "sqlite"),
			DSN:    getEnv("DB_DSN", "./data/absensi.db"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", "./logs/app.log"),
		},
		FaceRecognition: FaceRecognitionConfig{
			Enabled:        getEnv("FACE_RECOGNITION_ENABLED", "false") == "true",
			ModelsPath:     getEnv("FACE_MODELS_PATH", "./models"),
			MatchThreshold: parseFloat(getEnv("FACE_MATCH_THRESHOLD", "0.6")),
		},
		Environment: getEnv("ENVIRONMENT", "development"), // development or production
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(c.Security.AllowedIPs) == 0 {
		return fmt.Errorf("at least one allowed IP is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseAllowedIPs(ips string) []string {
	if ips == "" {
		return []string{}
	}
	parts := strings.Split(ips, ",")
	result := make([]string, 0, len(parts))
	for _, ip := range parts {
		trimmed := strings.TrimSpace(ip)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
