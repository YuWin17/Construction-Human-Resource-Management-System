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
	DatabaseDSN          string
	JWTSecret            string
	JWTTTL               time.Duration
	CORSAllowedOrigins   []string
	InitialAdminUsername string
	InitialAdminPassword string
	DailyReminderToken   string
	Timezone             string
}

// Load 在 .env 存在时先读取该文件，再读取环境变量。
// 环境变量优先于 .env 中的同名配置。
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
		DailyReminderToken:   getEnv("DAILY_REMINDER_TOKEN", ""),
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
