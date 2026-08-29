package repositories

import (
	"database/sql"
	"fmt"

	sqliterep "github.com/partflow/smart-store/internal/repositories/sqlite"

	"github.com/jmoiron/sqlx"
)

// RepositoryFactoryImpl implements the RepositoryFactory interface
type RepositoryFactoryImpl struct {
	postgresDB *sqlx.DB
	sqliteDB   *sql.DB
	mode       string // "offline" or "online"
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(postgresDB *sqlx.DB, sqliteDB *sql.DB, mode string) *RepositoryFactoryImpl {
	return &RepositoryFactoryImpl{
		postgresDB: postgresDB,
		sqliteDB:   sqliteDB,
		mode:       mode,
	}
}

// GetProductRepository returns the appropriate product repository
func (f *RepositoryFactoryImpl) GetProductRepository() ProductRepository {
	if f.mode == "offline" {
		return sqliterep.NewProductRepository(f.sqliteDB)
	}
	// For online mode, we would use PostgreSQL repository
	// This will be implemented as we migrate the existing code
	return sqliterep.NewProductRepository(f.sqliteDB)
}

// GetCustomerRepository returns the appropriate customer repository
func (f *RepositoryFactoryImpl) GetCustomerRepository() CustomerRepository {
	if f.mode == "offline" {
		return sqliterep.NewCustomerRepository(f.sqliteDB)
	}
	// For online mode, we would use PostgreSQL repository
	return sqliterep.NewCustomerRepository(f.sqliteDB)
}

// GetSupplierRepository returns the appropriate supplier repository
func (f *RepositoryFactoryImpl) GetSupplierRepository() SupplierRepository {
	if f.mode == "offline" {
		return sqliterep.NewSupplierRepository(f.sqliteDB)
	}
	// For online mode, we would use PostgreSQL repository
	return sqliterep.NewSupplierRepository(f.sqliteDB)
}

// GetInventoryRepository returns the appropriate inventory repository
func (f *RepositoryFactoryImpl) GetInventoryRepository() InventoryRepository {
	if f.mode == "offline" {
		return sqliterep.NewInventoryRepository(f.sqliteDB)
	}
	// For online mode, we would use PostgreSQL repository
	return sqliterep.NewInventoryRepository(f.sqliteDB)
}

// GetSalesRepository returns the appropriate sales repository
func (f *RepositoryFactoryImpl) GetSalesRepository() SalesRepository {
	if f.mode == "offline" {
		return sqliterep.NewSalesRepository(f.sqliteDB)
	}
	// For online mode, we would use PostgreSQL repository
	return sqliterep.NewSalesRepository(f.sqliteDB)
}

// SetOperatingMode updates the operating mode
func (f *RepositoryFactoryImpl) SetOperatingMode(mode string) error {
	if mode != "offline" && mode != "online" {
		return fmt.Errorf("invalid operating mode: %s", mode)
	}
	f.mode = mode
	return nil
}

// GetOperatingMode returns the current operating mode
func (f *RepositoryFactoryImpl) GetOperatingMode() string {
	return f.mode
}
