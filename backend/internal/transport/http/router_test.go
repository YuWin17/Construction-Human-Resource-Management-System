package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"construction-hrms/backend/internal/config"
	"construction-hrms/backend/internal/database"
	"construction-hrms/backend/internal/domain"
	"construction-hrms/backend/internal/repository"
	"construction-hrms/backend/internal/service"
	httpapi "construction-hrms/backend/internal/transport/http"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
)

func TestRouterHealthAndProtectedRoute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&domain.Admin{}); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	repo := repository.NewAdminRepository(db)
	auth := service.NewAuthService(repo, "test-secret", time.Hour)
	if err := auth.EnsureInitialAdmin(t.Context(), "admin", "password"); err != nil {
		t.Fatalf("initialize admin: %v", err)
	}

	router := httpapi.NewRouter(config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}}, slog.Default(), auth)

	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", health.Code)
	}

	protected := httptest.NewRecorder()
	router.ServeHTTP(protected, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil))
	if protected.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", protected.Code)
	}
}

func TestDailyReminderMessageRouteRequiresConfiguredToken(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	auth := service.NewAuthService(repository.NewAdminRepository(db), "test-secret", time.Hour)
	reminders := service.NewReminderService(db)
	noTokenRouter := httpapi.NewRouter(config.Config{}, slog.Default(), auth, reminders)
	noToken := httptest.NewRecorder()
	noTokenRouter.ServeHTTP(noToken, httptest.NewRequest(http.MethodGet, "/api/v1/integrations/wecom/daily-reminder", nil))
	if noToken.Code != http.StatusNotFound {
		t.Fatalf("expected unconfigured route to return 404, got %d", noToken.Code)
	}

	router := httpapi.NewRouter(config.Config{DailyReminderToken: "scheduler-token"}, slog.Default(), auth, reminders)
	invalid := httptest.NewRecorder()
	router.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, "/api/v1/integrations/wecom/daily-reminder?token=wrong", nil))
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid token status 401, got %d", invalid.Code)
	}

	valid := httptest.NewRecorder()
	router.ServeHTTP(valid, httptest.NewRequest(http.MethodGet, "/api/v1/integrations/wecom/daily-reminder?token=scheduler-token", nil))
	if valid.Code != http.StatusOK {
		t.Fatalf("expected valid token status 200, got %d", valid.Code)
	}
	if contentType := valid.Header().Get("Content-Type"); contentType != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", contentType)
	}
	if body := valid.Body.String(); body == "" || body[0:3] != "【" {
		t.Fatalf("unexpected reminder body: %q", body)
	}
}
