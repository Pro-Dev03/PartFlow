package localdb

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateExistingFileAcceptsPartFlowDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partflow.sqlite")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)
	local, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	if err := local.DB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExistingFile(path); err != nil {
		t.Fatalf("valid PartFlow database rejected: %v", err)
	}
}

func TestValidateExistingFileRejectsInvalidAndForeignSQLiteFiles(t *testing.T) {
	invalid := filepath.Join(t.TempDir(), "invalid.db")
	if err := os.WriteFile(invalid, []byte("not a sqlite database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExistingFile(invalid); err == nil {
		t.Fatal("invalid file was accepted")
	}

	foreign := filepath.Join(t.TempDir(), "foreign.sqlite")
	db, err := sql.Open("sqlite", foreign)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE unrelated (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExistingFile(foreign); err == nil {
		t.Fatal("foreign SQLite database was accepted as PartFlow data")
	}
}

func TestCreateConsistentBackupIncludesCommittedWalData(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", source)
	local, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := local.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, "backup-product", "BACKUP-SKU", "WAL product", now, now); err != nil {
		t.Fatal(err)
	}
	walInfo, err := os.Stat(source + "-wal")
	if err != nil {
		t.Fatalf("source WAL should exist while the database remains open: %v", err)
	}
	if walInfo.Size() <= 32 {
		t.Fatalf("source WAL is unexpectedly empty: %d bytes", walInfo.Size())
	}

	destination := filepath.Join(t.TempDir(), "backup.db")
	if err := CreateConsistentBackup(source, destination); err != nil {
		t.Fatalf("create consistent backup: %v", err)
	}
	if err := ValidateExistingFile(destination); err != nil {
		t.Fatalf("validate standalone backup: %v", err)
	}
	backup, err := sql.Open("sqlite", destination)
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Close()
	var productName string
	if err := backup.QueryRow(`SELECT name FROM products WHERE id = ?`, "backup-product").Scan(&productName); err != nil {
		t.Fatalf("read committed source row from backup: %v", err)
	}
	if productName != "WAL product" {
		t.Fatalf("backup product name=%q", productName)
	}
	if err := CreateConsistentBackup(source, destination); err == nil {
		t.Fatal("existing backup destination should be rejected")
	}
}
