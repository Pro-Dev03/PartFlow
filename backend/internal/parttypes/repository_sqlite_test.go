package parttypes

import (
	"context"
	"path/filepath"
	"testing"

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
