package service_test

import (
	"context"
	"errors"
	"testing"

	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTalentServiceCreatesOneCertificatePerTalentAndAllowsSamePersonForAnotherCertificate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Talent{}, &domain.CertificateCatalog{}, &domain.Certificate{}, &domain.AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	serviceUnderTest := service.NewTalentService(repository.NewTalentRepository(db))
	input := service.TalentInput{Name: "张三", IDCardNumber: "110101199001011234", Phone: "13800138000", Status: domain.TalentStatusActive, Certificates: []service.CertificateInput{{Name: "一级建造师", Category: "注册类", RegistrationStatus: "active"}}}
	talent, err := serviceUnderTest.Create(context.Background(), input, "admin-1")
	if err != nil {
		t.Fatalf("create talent: %v", err)
	}
	if len(talent.Certificates) != 1 || talent.Certificates[0].CertificateNameSnapshot != "一级建造师" {
		t.Fatalf("certificate was not created: %+v", talent.Certificates)
	}
	secondInput := input
	secondInput.Certificates = []service.CertificateInput{{Name: "二级建造师", RegistrationStatus: "active"}}
	if _, err := serviceUnderTest.Create(context.Background(), secondInput, "admin-1"); err != nil {
		t.Fatalf("same person should be allowed for another certificate: %v", err)
	}
	input.Certificates = append(input.Certificates, service.CertificateInput{Name: "二级建造师", RegistrationStatus: "active"})
	if _, err := serviceUnderTest.Create(context.Background(), input, "admin-1"); !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected one-certificate validation error, got %v", err)
	}
}
