package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"construction-hrms/backend/internal/cloudbasepg"
	"construction-hrms/backend/internal/config"
	"construction-hrms/backend/internal/database"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	httpapi "construction-hrms/backend/internal/transport/http"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cloudBasePG := cfg.DatabaseDriver == "cloudbase_pg"
	var db *gorm.DB
	if cloudBasePG {
		// The container keeps only a short-lived working set. CloudBase PG is
		// loaded before serving and receives every successful API mutation.
		db, err = database.OpenSQLite(":memory:")
	} else {
		db, err = database.Open(cfg.DatabaseDriver, cfg.DatabaseDSN)
	}
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("get database connection", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()
	if cloudBasePG {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := database.Migrate(db); err != nil {
		logger.Error("migrate database", "error", err)
		os.Exit(1)
	}
	var synchronizer *cloudbasepg.Synchronizer
	if cloudBasePG {
		client, clientErr := cloudbasepg.New(cfg.CloudBaseEnvID, cfg.CloudBaseAPIKey, cfg.CloudBaseAPIBaseURL)
		if clientErr != nil {
			logger.Error("configure CloudBase PostgreSQL", "error", clientErr)
			os.Exit(1)
		}
		if err := client.Load(context.Background(), db); err != nil {
			logger.Error("load CloudBase PostgreSQL data", "error", err)
			os.Exit(1)
		}
		synchronizer = cloudbasepg.NewSynchronizer(client, db, logger)
	}

	adminRepository := repository.NewAdminRepository(db)
	talentRepository := repository.NewTalentRepository(db)
	contractRepository := repository.NewContractRepository(db)
	authService := service.NewAuthService(adminRepository, cfg.JWTSecret, cfg.JWTTTL)
	talentService := service.NewTalentService(talentRepository)
	contractService := service.NewContractService(contractRepository)
	reminderService := service.NewReminderService(db)
	beforeInitialAdmin, err := cloudbasepg.TakeSnapshot(db)
	if err != nil {
		logger.Error("snapshot initial database state", "error", err)
		os.Exit(1)
	}
	if err := authService.EnsureInitialAdmin(context.Background(), cfg.InitialAdminUsername, cfg.InitialAdminPassword); err != nil {
		logger.Error("initialize administrator", "error", err)
		os.Exit(1)
	}
	if cloudBasePG {
		afterInitialAdmin, snapshotErr := cloudbasepg.TakeSnapshot(db)
		if snapshotErr != nil {
			logger.Error("snapshot initialized administrator", "error", snapshotErr)
			os.Exit(1)
		}
		if err := synchronizer.Apply(context.Background(), beforeInitialAdmin, afterInitialAdmin); err != nil {
			logger.Error("persist initialized administrator", "error", err)
			os.Exit(1)
		}
	}

	services := []any{talentService, contractService, reminderService}
	if synchronizer != nil {
		services = append(services, synchronizer.Middleware())
	}
	router := httpapi.NewRouter(cfg, logger, authService, services...)
	logger.Info("api server starting", "address", cfg.HTTPAddr, "environment", cfg.AppEnv)
	if err := router.Run(cfg.HTTPAddr); err != nil {
		logger.Error("api server stopped", "error", err)
		os.Exit(1)
	}
}
