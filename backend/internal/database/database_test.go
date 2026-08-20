package database

import (
	"path/filepath"
	"testing"

	"construction-hrms/backend/internal/domain"
)

func TestMigratePreservesExistingTalentRecords(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "hrms.db")
	db, err := OpenSQLite(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	want := domain.Talent{
		ID:           "existing-talent",
		Code:         "RCEXISTING",
		Name:         "Existing Record",
		IDCardNumber: "110101199001010000",
		Phone:        "13800000000",
		Status:       domain.TalentStatusActive,
	}
	if err := db.Create(&want).Error; err != nil {
		t.Fatalf("create existing talent: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}

	var got domain.Talent
	if err := db.First(&got, "id = ?", want.ID).Error; err != nil {
		t.Fatalf("load existing talent after migration: %v", err)
	}
	if got.Name != want.Name || got.Code != want.Code || got.IDCardNumber != want.IDCardNumber {
		t.Fatalf("migration changed existing record: got %+v", got)
	}
}
