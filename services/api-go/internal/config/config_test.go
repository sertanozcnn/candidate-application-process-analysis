package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("CAPA_APP_ENV", "")
	t.Setenv("CAPA_LOG_LEVEL", "")
	t.Setenv("CAPA_API_ADDR", "")
	t.Setenv("CAPA_API_SHUTDOWN_TIMEOUT", "")
	t.Setenv("CAPA_ADMIN_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://capa:capa_local_password@localhost:55432/capa?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want default local database URL", cfg.DatabaseURL)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}
	if len(cfg.AdminOrigins) != 2 || cfg.AdminOrigins[0] != "http://localhost:3001" {
		t.Fatalf("AdminOrigins = %#v, want local admin origins", cfg.AdminOrigins)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("CAPA_APP_ENV", "production")
	t.Setenv("CAPA_LOG_LEVEL", "debug")
	t.Setenv("CAPA_API_ADDR", "127.0.0.1:18080")
	t.Setenv("CAPA_DATABASE_URL", "postgres://user:pass@localhost:15432/db?sslmode=disable")
	t.Setenv("CAPA_API_SHUTDOWN_TIMEOUT", "2s")
	t.Setenv("CAPA_ADMIN_ORIGINS", "https://admin.example, https://admin.internal")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "production" {
		t.Fatalf("AppEnv = %q, want production", cfg.AppEnv)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.HTTPAddr != "127.0.0.1:18080" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:18080", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:15432/db?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want override database URL", cfg.DatabaseURL)
	}
	if cfg.ShutdownTimeout != 2*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 2s", cfg.ShutdownTimeout)
	}
	if len(cfg.AdminOrigins) != 2 || cfg.AdminOrigins[1] != "https://admin.internal" {
		t.Fatalf("AdminOrigins = %#v, want configured origins", cfg.AdminOrigins)
	}
}

func TestLoadReadsDotEnv(t *testing.T) {
	unsetEnvForTest(t, "CAPA_APP_ENV")
	unsetEnvForTest(t, "CAPA_LOG_LEVEL")
	unsetEnvForTest(t, "CAPA_API_ADDR")
	unsetEnvForTest(t, "CAPA_DATABASE_URL")
	unsetEnvForTest(t, "CAPA_API_SHUTDOWN_TIMEOUT")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Chdir(originalDir) error = %v", err)
		}
	})

	envPath := filepath.Join(tempDir, ".env")
	content := []byte("CAPA_APP_ENV=test\nCAPA_LOG_LEVEL=warn\nCAPA_API_ADDR=127.0.0.1:19090\nCAPA_DATABASE_URL=postgres://env:pass@localhost:25432/envdb?sslmode=disable\nCAPA_API_SHUTDOWN_TIMEOUT=3s\n")
	if err := os.WriteFile(envPath, content, 0o600); err != nil {
		t.Fatalf("WriteFile(.env) error = %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "test" {
		t.Fatalf("AppEnv = %q, want test", cfg.AppEnv)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("LogLevel = %q, want warn", cfg.LogLevel)
	}
	if cfg.HTTPAddr != "127.0.0.1:19090" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:19090", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://env:pass@localhost:25432/envdb?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want .env database URL", cfg.DatabaseURL)
	}
	if cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 3s", cfg.ShutdownTimeout)
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	value, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%s) error = %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			if err := os.Setenv(key, value); err != nil {
				t.Fatalf("Setenv(%s) error = %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%s) error = %v", key, err)
		}
	})
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	t.Setenv("CAPA_API_SHUTDOWN_TIMEOUT", "soon")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
