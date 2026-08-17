// Package config loads application settings from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contains runtime settings used by the API service.
type Config struct {
	AppEnv               string
	HTTPAddr             string
	DatabaseDSN          string
	JWTSecret            string
	JWTTTL               time.Duration
	CORSAllowedOrigins   []string
	InitialAdminUsername string
	InitialAdminPassword string
	Timezone             string
}

// Load reads .env when present, then reads environment variables.
// Real environment variables take precedence over values in .env.
func Load() (Config, error) {
	_ = godotenv.Load()

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "8"))
	if err != nil || ttlHours <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL_HOURS must be a positive integer")
	}

	cfg := Config{
		AppEnv:               getEnv("APP_ENV", "development"),
		HTTPAddr:             getEnv("HTTP_ADDR", ":8080"),
		DatabaseDSN:          getEnv("DATABASE_DSN", "./data/hrms.db"),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTTTL:               time.Duration(ttlHours) * time.Hour,
		CORSAllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		InitialAdminUsername: getEnv("INITIAL_ADMIN_USERNAME", ""),
		InitialAdminPassword: getEnv("INITIAL_ADMIN_PASSWORD", ""),
		Timezone:             getEnv("TIMEZONE", "Asia/Shanghai"),
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return Config{}, fmt.Errorf("invalid TIMEZONE: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
