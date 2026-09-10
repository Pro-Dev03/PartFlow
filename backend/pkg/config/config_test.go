package config

import (
	"testing"
)

func TestLoad_RejectsDefaultJWTSecretInReleaseMode(t *testing.T) {
	t.Setenv("SERVER_MODE", "release")
	t.Setenv("DATABASE_URL", "sqlite://test.db")
	t.Setenv("JWT_SECRET", "change-this-secret-in-production")
	t.Setenv("DISABLE_AUTH", "false")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when JWT_SECRET is left at the default production value")
	}
}

func TestLoad_RejectsDisabledAuthInReleaseMode(t *testing.T) {
	t.Setenv("SERVER_MODE", "release")
	t.Setenv("DATABASE_URL", "sqlite://test.db")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("DISABLE_AUTH", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when DISABLE_AUTH is enabled in release mode")
	}
}

func TestLoad_RejectsCloudAuthDisabledInReleaseMode(t *testing.T) {
	t.Setenv("SERVER_MODE", "release")
	t.Setenv("DATABASE_URL", "sqlite://test.db")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("DISABLE_AUTH", "false")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "false")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when PARTFLOW_REQUIRE_CLOUD_AUTH is disabled in release mode")
	}
}

func TestLoad_RejectsLocalAuthBypassInReleaseMode(t *testing.T) {
	t.Setenv("SERVER_MODE", "release")
	t.Setenv("DATABASE_URL", "sqlite://test.db")
	t.Setenv("JWT_SECRET", "production-secret")
	t.Setenv("DISABLE_AUTH", "false")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS", "true")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS is enabled in release mode")
	}
}

func TestResolveDatabaseURL_LocalModeFallsBackToSQLite(t *testing.T) {
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_LOCAL", "")
	t.Setenv("DATABASE_URL_CLOUD", "")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", "C:/tmp/partflow-local.db")

	url, source, err := ResolveDatabaseURL()
	if err != nil {
		t.Fatalf("ResolveDatabaseURL() returned error for local mode: %v", err)
	}
	if source != "local" {
		t.Fatalf("ResolveDatabaseURL() source = %q, want local", source)
	}
	if url == "" || url[:len("sqlite://")] != "sqlite://" {
		t.Fatalf("ResolveDatabaseURL() url = %q, want sqlite:// fallback", url)
	}
}

func TestResolveDatabaseURL_CloudModeRequiresConfiguredCloudURL(t *testing.T) {
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_LOCAL", "")
	t.Setenv("DATABASE_URL_CLOUD", "")

	_, _, err := ResolveDatabaseURL()
	if err == nil {
		t.Fatal("ResolveDatabaseURL() expected error when cloud mode has no configured URL")
	}
}
