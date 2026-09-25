package config

import (
	"strings"
	"testing"
)

func setProductionConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SERVER_MODE", "")
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_LOCAL", "sqlite://partflow-test.db")
	t.Setenv("DATABASE_URL_CLOUD", "")
	t.Setenv("DB_LOCAL_URL", "")
	t.Setenv("DB_CLOUD_URL", "")
	t.Setenv("LOCAL_DATABASE_URL", "")
	t.Setenv("CLOUD_DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "production-test-secret-that-is-not-the-default")
	t.Setenv("DISABLE_AUTH", "false")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "")
	t.Setenv("PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
}

func TestLoadTreatsAPPENVProductionAsProduction(t *testing.T) {
	setProductionConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServerMode != "production" {
		t.Fatalf("ServerMode = %q, want normalized production", cfg.ServerMode)
	}
}

func TestLoadRejectsUnsafeProductionAuthenticationConfig(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{"default jwt secret", "JWT_SECRET", "change-this-secret-in-production", "JWT_SECRET must be changed"},
		{"short jwt secret", "JWT_SECRET", "short-secret", "at least 32 bytes"},
		{"disabled auth", "DISABLE_AUTH", "true", "DISABLE_AUTH must be false"},
		{"cloud auth disabled", "PARTFLOW_REQUIRE_CLOUD_AUTH", "false", "PARTFLOW_REQUIRE_CLOUD_AUTH must be true"},
		{"local auth bypass enabled", "PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS", "true", "PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS must be false"},
		{"credentialed wildcard CORS", "CORS_ALLOWED_ORIGINS", "*", "CORS_ALLOWED_ORIGINS must be an explicit origin allowlist"},
		{"opaque null origin CORS", "CORS_ALLOWED_ORIGINS", "https://partflow-hpv7.onrender.com,null", "must not include the opaque null origin"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setProductionConfigEnv(t)
			t.Setenv(test.key, test.value)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want error containing %q", err, test.want)
			}
		})
	}
}
