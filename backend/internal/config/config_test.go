package config

import (
	"strings"
	"testing"
)

func TestLoadRejectsSQLiteInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_DRIVER", "sqlite")
	t.Setenv("DATABASE_DSN", "/app/data/hrms.db")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_DRIVER=cloudbase_pg") {
		t.Fatalf("expected production SQLite configuration to be rejected, got %v", err)
	}
}

func TestLoadRequiresManagedDatabaseDSNInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_DRIVER", "sqlite")
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_DSN is required") {
		t.Fatalf("expected missing production database DSN to be rejected, got %v", err)
	}
}

func TestLoadRequiresCloudBasePGCredentialsInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_DRIVER", "cloudbase_pg")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("CLOUDBASE_ENV_ID", "")
	t.Setenv("CLOUDBASE_API_KEY", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "CLOUDBASE_ENV_ID") {
		t.Fatalf("expected missing CloudBase environment to be rejected, got %v", err)
	}
}
