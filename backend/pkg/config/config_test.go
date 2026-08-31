package config

import "testing"

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
