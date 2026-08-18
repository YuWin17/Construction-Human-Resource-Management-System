package service_test

import (
	"context"
	"testing"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTalentServiceCreatesAndUpdatesTalentCertificate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Talent{}, &domain.CertificateCatalog{}, &domain.Certificate{}, &domain.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	serviceUnderTest := service.NewTalentService(repository.NewTalentRepository(db))
	input := service.TalentInput{Name: "张三", IDCardNumber: "110101199001011234", Phone: "13800138000", Status: domain.TalentStatusActive, Certificate: &service.CertificateInput{Name: "一级建造师", Category: "注册类", RegistrationStatus: "active"}}
	talent, err := serviceUnderTest.Create(context.Background(), input, "admin-1")
	if err != nil {
		t.Fatalf("create talent: %v", err)
	}
	if talent.Certificate == nil || talent.Certificate.CertificateNameSnapshot != "一级建造师" {
		t.Fatalf("certificate was not created: %+v", talent.Certificate)
	}
	certificateID := talent.Certificate.ID
	input.Certificate = &service.CertificateInput{Name: "二级建造师", Category: "注册类", Specialty: "建筑工程", RegistrationStatus: "registered"}
	updated, err := serviceUnderTest.Update(context.Background(), talent.ID, input, "admin-1")
	if err != nil {
		t.Fatalf("update talent certificate: %v", err)
	}
	if updated.Certificate == nil || updated.Certificate.ID != certificateID || updated.Certificate.CertificateNameSnapshot != "二级建造师" || updated.Certificate.Specialty != "建筑工程" {
		t.Fatalf("certificate was not updated: %+v", updated.Certificate)
	}
}

func TestTalentServiceCreatesCertificateWithExpiryDate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Talent{}, &domain.CertificateCatalog{}, &domain.Certificate{}, &domain.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	serviceUnderTest := service.NewTalentService(repository.NewTalentRepository(db))
	talent, err := serviceUnderTest.Create(context.Background(), service.TalentInput{
		Name:         "测试人员3",
		IDCardNumber: "444444444444444444",
		Phone:        "13231332322",
		Certificate: &service.CertificateInput{
			Name:      "测试证书4",
			ExpiresOn: "2026-08-28",
		},
	}, "admin-1")
	if err != nil {
		t.Fatalf("create talent with certificate expiry date: %v", err)
	}
	if talent.Certificate == nil || talent.Certificate.ExpiresOn != "2026-08-28" {
		t.Fatalf("certificate expiry date was not saved: %+v", talent.Certificate)
	}
}
