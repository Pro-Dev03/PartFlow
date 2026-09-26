package products

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestDeleteCategoryUpdatesProductsAtomicallySQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE categories (id TEXT PRIMARY KEY);
		CREATE TABLE products (id TEXT PRIMARY KEY, category_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	categoryID, productID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO categories (id) VALUES (?)`, categoryID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id,category_id) VALUES (?,?)`, productID.String(), categoryID.String()); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	if _, err := db.Exec(`CREATE TRIGGER reject_category_delete BEFORE DELETE ON categories BEGIN SELECT RAISE(ABORT,'delete rejected'); END`); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteCategory(context.Background(), categoryID); err == nil {
		t.Fatal("category deletion should fail while the rejection trigger is installed")
	}
	var categoryRef string
	if err := db.Get(&categoryRef, `SELECT category_id FROM products WHERE id=?`, productID.String()); err != nil {
		t.Fatal(err)
	}
	if categoryRef != categoryID.String() {
		t.Fatalf("failed delete detached product category to %q; want rollback to %q", categoryRef, categoryID)
	}
	if _, err := db.Exec(`DROP TRIGGER reject_category_delete`); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteCategory(context.Background(), categoryID); err != nil {
		t.Fatalf("delete category and detach product: %v", err)
	}
	var categoryCount int
	var nullCategory bool
	if err := db.Get(&categoryCount, `SELECT COUNT(*) FROM categories WHERE id=?`, categoryID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&nullCategory, `SELECT category_id IS NULL FROM products WHERE id=?`, productID.String()); err != nil {
		t.Fatal(err)
	}
	if categoryCount != 0 || !nullCategory {
		t.Fatalf("category cleanup left category=%d product_category_null=%t; want 0/true", categoryCount, nullCategory)
	}
}

func TestDeleteBrandDetachesProductsAndDeletesBrandAtomicallySQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE brands (id TEXT PRIMARY KEY);
		CREATE TABLE products (id TEXT PRIMARY KEY, brand_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	linkedBrandID, unusedBrandID, productID := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{linkedBrandID, unusedBrandID} {
		if _, err := db.Exec(`INSERT INTO brands (id) VALUES (?)`, id.String()); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO products (id,brand_id) VALUES (?,?)`, productID.String(), linkedBrandID.String()); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	if err := repo.DeleteBrand(context.Background(), linkedBrandID); err != nil {
		t.Fatalf("delete linked brand and detach product: %v", err)
	}
	if err := repo.DeleteBrand(context.Background(), unusedBrandID); err != nil {
		t.Fatalf("delete unused brand: %v", err)
	}
	var linkedBrandCount, unusedBrandCount int
	if err := db.Get(&linkedBrandCount, `SELECT COUNT(*) FROM brands WHERE id=?`, linkedBrandID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&unusedBrandCount, `SELECT COUNT(*) FROM brands WHERE id=?`, unusedBrandID.String()); err != nil {
		t.Fatal(err)
	}
	var nullBrand bool
	if err := db.Get(&nullBrand, `SELECT brand_id IS NULL FROM products WHERE id=?`, productID.String()); err != nil {
		t.Fatal(err)
	}
	if linkedBrandCount != 0 || unusedBrandCount != 0 || !nullBrand {
		t.Fatalf("brand cleanup left linked=%d unused=%d product_brand_null=%t; want 0/0/true", linkedBrandCount, unusedBrandCount, nullBrand)
	}
}
