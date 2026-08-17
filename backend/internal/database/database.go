// Package database owns database connection and stage-A schema migration.
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"construction-hrms/backend/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenSQLite creates parent directories before opening the configured database.
func OpenSQLite(dsn string) (*gorm.DB, error) {
	if dsn != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dsn), 0o750); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	return db, nil
}

// Migrate applies the small stage-A schema. Versioned SQL migrations will
// replace this temporary bootstrap as business tables are introduced.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domain.Admin{},
		&domain.Talent{},
		&domain.CertificateCatalog{},
		&domain.Certificate{},
		&domain.Contract{},
		&domain.Reminder{},
		&domain.SystemSetting{},
		&domain.Company{}, &domain.CompanyRequirement{}, &domain.DeliveryOrder{}, &domain.DeliveryOrderTalent{},
		&domain.AuditLog{},
	); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	// Existing installations predate the human-readable talent number. Backfill
	// before applying the unique index so historical records remain usable.
	if err := db.Exec("UPDATE talents SET code = 'RC' || substr(replace(id, '-', ''), 1, 14) WHERE code IS NULL OR code = ''").Error; err != nil {
		return fmt.Errorf("backfill talent codes: %w", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_talents_code ON talents(code)").Error; err != nil {
		return fmt.Errorf("create talent code index: %w", err)
	}
	// A person can have several records, one for each certificate. Convert the
	// former unique identity-number index to a normal lookup index without
	// altering historical talent records.
	if err := db.Exec("DROP INDEX IF EXISTS idx_talents_id_card_number").Error; err != nil {
		return fmt.Errorf("drop unique talent identity index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_talents_id_card_number ON talents(id_card_number)").Error; err != nil {
		return fmt.Errorf("create talent identity index: %w", err)
	}
	return nil
}
