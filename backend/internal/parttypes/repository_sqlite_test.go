package parttypes

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSQLitePartTypesAndSpecifications(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "parttypes.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	repo := NewRepository(sqlx.NewDb(database.DB, "sqlite"))
	ctx := context.Background()
	partType := &PartType{NameAr: "ذاكرة", NameEn: "Memory", IsActive: true}
	if err := repo.CreatePartType(ctx, partType); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetPartType(ctx, partType.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.NameEn != "Memory" {
		t.Fatalf("part type mismatch: %#v", got)
	}
	spec := &PartSpecification{NameAr: "السعة", NameEn: "Capacity", DataType: "number", Options: []string{"8", "16"}, IsRequired: true}
	if err := repo.CreateSpecification(ctx, spec); err != nil {
		t.Fatal(err)
	}
	if err := repo.LinkSpecification(ctx, &TypeSpecification{PartTypeID: partType.ID, SpecificationID: spec.ID}); err != nil {
		t.Fatal(err)
	}
	specs, err := repo.GetTypeSpecifications(ctx, partType.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || len(specs[0].Options) != 2 {
		t.Fatalf("specifications mismatch: %#v", specs)
	}
}

func TestParsePartTypeMapHandlesNumericAndByteValues(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	record := map[string]any{
		"id":         []byte(id.String()),
		"name_ar":    []byte("ذاكرة"),
		"name_en":    "Memory",
		"icon":       "memory",
		"color":      "#123456",
		"is_active":  1,
		"sort_order": int64(7),
		"created_at": "2024-01-02T03:04:05Z",
		"updated_at": "2024-01-03T03:04:05Z",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("parsePartTypeMap panicked: %v", r)
		}
	}()

	partType, err := parsePartTypeMap(record)
	if err != nil {
		t.Fatalf("parsePartTypeMap returned error: %v", err)
	}
	if partType.ID != id {
		t.Fatalf("ID mismatch: got %s, want %s", partType.ID, id)
	}
	if !partType.IsActive {
		t.Fatal("IsActive should be true")
	}
	if partType.SortOrder != 7 {
		t.Fatalf("SortOrder = %d, want 7", partType.SortOrder)
	}
}
