package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"construction-hrms/backend/internal/config"
	"construction-hrms/backend/internal/database"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	httpapi "construction-hrms/backend/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	db, err := database.OpenSQLite(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		logger.Error("migrate database", "error", err)
		os.Exit(1)
	}

	adminRepository := repository.NewAdminRepository(db)
	talentRepository := repository.NewTalentRepository(db)
	contractRepository := repository.NewContractRepository(db)
	authService := service.NewAuthService(adminRepository, cfg.JWTSecret, cfg.JWTTTL)
	talentService := service.NewTalentService(talentRepository)
	contractService := service.NewContractService(contractRepository)
	reminderService := service.NewReminderService(db)
	if err := authService.EnsureInitialAdmin(context.Background(), cfg.InitialAdminUsername, cfg.InitialAdminPassword); err != nil {
		logger.Error("initialize administrator", "error", err)
		os.Exit(1)
	}

	router := httpapi.NewRouter(cfg, logger, authService, talentService, contractService, reminderService)
	logger.Info("api server starting", "address", cfg.HTTPAddr, "environment", cfg.AppEnv)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		logger.Error("api server stopped", "error", err)
		os.Exit(1)
	}
}
