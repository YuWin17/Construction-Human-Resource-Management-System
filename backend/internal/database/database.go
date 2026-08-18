// Package database 管理数据库连接和初始表结构迁移。
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

// OpenSQLite 在打开配置的数据库前创建其父目录。
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

// Migrate 执行当前业务表结构迁移。
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
	// 旧安装版本没有可读的人才编号。建立唯一索引前先回填，确保历史记录可继续使用。
	if err := db.Exec("UPDATE talents SET code = 'RC' || substr(replace(id, '-', ''), 1, 14) WHERE code IS NULL OR code = ''").Error; err != nil {
		return fmt.Errorf("backfill talent codes: %w", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_talents_code ON talents(code)").Error; err != nil {
		return fmt.Errorf("create talent code index: %w", err)
	}
	// 同一人员可因不同证书保留多条记录。将原身份证号唯一索引改为普通查询索引，不修改历史档案。
	if err := db.Exec("DROP INDEX IF EXISTS idx_talents_id_card_number").Error; err != nil {
		return fmt.Errorf("drop unique talent identity index: %w", err)
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_talents_id_card_number ON talents(id_card_number)").Error; err != nil {
		return fmt.Errorf("create talent identity index: %w", err)
	}
	return nil
}
