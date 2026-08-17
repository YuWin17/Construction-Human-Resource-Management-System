package service_test

import (
	"context"
	"testing"
	"time"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthServiceInitializesAndAuthenticatesAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Admin{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	repo := repository.NewAdminRepository(db)
	auth := service.NewAuthService(repo, "test-secret", time.Hour)
	ctx := context.Background()

	if err := auth.EnsureInitialAdmin(ctx, "admin", "strong-password"); err != nil {
		t.Fatalf("initialize admin: %v", err)
	}

	token, admin, err := auth.Login(ctx, "admin", "strong-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || admin.Username != "admin" {
		t.Fatalf("unexpected login result: token=%q admin=%+v", token, admin)
	}

	current, err := auth.CurrentAdmin(ctx, token)
	if err != nil {
		t.Fatalf("get current admin: %v", err)
	}
	if current.ID != admin.ID {
		t.Fatalf("expected admin id %q, got %q", admin.ID, current.ID)
	}
}
