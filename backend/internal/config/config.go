// Package config 从环境变量加载应用配置。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config 保存 API 服务运行所需的配置。
type Config struct {
	AppEnv               string
	HTTPAddr             string
	DatabaseDriver       string
	DatabaseDSN          string
	JWTSecret            string
	JWTTTL               time.Duration
	CORSAllowedOrigins   []string
	InitialAdminUsername string
	InitialAdminPassword string
	DailyReminderToken   string
	Timezone             string
	CloudBaseEnvID       string
	CloudBaseAPIKey      string
	CloudBaseAPIBaseURL  string
}

// Load 在 .env 存在时先读取该文件，再读取环境变量。
// 环境变量优先于 .env 中的同名配置。
func Load() (Config, error) {
	_ = godotenv.Load()
	appEnv := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "development")))
	databaseDriverDefault := "sqlite"
	databaseDSNDefault := "./data/hrms.db"
	if appEnv == "production" {
		// Production must fail closed instead of silently creating a database in
		// the Cloud Run container's ephemeral filesystem.
		databaseDriverDefault = "cloudbase_pg"
		databaseDSNDefault = ""
	}

	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "8"))
	if err != nil || ttlHours <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL_HOURS must be a positive integer")
	}

	cfg := Config{
		AppEnv:               appEnv,
		HTTPAddr:             getEnv("HTTP_ADDR", ":8080"),
		DatabaseDriver:       strings.ToLower(strings.TrimSpace(getEnv("DATABASE_DRIVER", databaseDriverDefault))),
		DatabaseDSN:          strings.TrimSpace(getEnv("DATABASE_DSN", databaseDSNDefault)),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTTTL:               time.Duration(ttlHours) * time.Hour,
		CORSAllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		InitialAdminUsername: getEnv("INITIAL_ADMIN_USERNAME", ""),
		InitialAdminPassword: getEnv("INITIAL_ADMIN_PASSWORD", ""),
		DailyReminderToken:   getEnv("DAILY_REMINDER_TOKEN", ""),
		Timezone:             getEnv("TIMEZONE", "Asia/Shanghai"),
		CloudBaseEnvID:       strings.TrimSpace(getEnv("CLOUDBASE_ENV_ID", "")),
		CloudBaseAPIKey:      strings.TrimSpace(getEnv("CLOUDBASE_API_KEY", "")),
		CloudBaseAPIBaseURL:  strings.TrimSpace(getEnv("CLOUDBASE_PG_REST_BASE_URL", "")),
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.DatabaseDriver != "sqlite" && cfg.DatabaseDriver != "mysql" && cfg.DatabaseDriver != "cloudbase_pg" {
		return Config{}, fmt.Errorf("DATABASE_DRIVER must be sqlite, mysql, or cloudbase_pg")
	}
	if cfg.DatabaseDriver != "cloudbase_pg" && cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}
	if cfg.AppEnv == "production" && cfg.DatabaseDriver != "cloudbase_pg" {
		return Config{}, fmt.Errorf("production requires DATABASE_DRIVER=cloudbase_pg to prevent ephemeral local storage")
	}
	if cfg.DatabaseDriver == "cloudbase_pg" {
		if cfg.CloudBaseEnvID == "" {
			return Config{}, fmt.Errorf("CLOUDBASE_ENV_ID is required for CloudBase PostgreSQL")
		}
		if cfg.CloudBaseAPIKey == "" {
			return Config{}, fmt.Errorf("CLOUDBASE_API_KEY is required for CloudBase PostgreSQL")
		}
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
