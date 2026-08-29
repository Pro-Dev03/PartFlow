package repositories

import "context"

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product interface{}) (string, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetBySKU(ctx context.Context, sku string) (interface{}, error)
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
	Update(ctx context.Context, id string, product interface{}) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error)
	Count(ctx context.Context) (int, error)
}

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	Create(ctx context.Context, customer interface{}) (string, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByCode(ctx context.Context, code string) (interface{}, error)
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
	Update(ctx context.Context, id string, customer interface{}) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error)
	Count(ctx context.Context) (int, error)
}

// SupplierRepository defines the interface for supplier data access
type SupplierRepository interface {
	Create(ctx context.Context, supplier interface{}) (string, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByCode(ctx context.Context, code string) (interface{}, error)
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
	Update(ctx context.Context, id string, supplier interface{}) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error)
	Count(ctx context.Context) (int, error)
}

// InventoryRepository defines the interface for inventory data access
type InventoryRepository interface {
	Create(ctx context.Context, item interface{}) (string, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByBarcode(ctx context.Context, barcode string) (interface{}, error)
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
	ListByProduct(ctx context.Context, productID string, limit, offset int) ([]interface{}, error)
	Update(ctx context.Context, id string, item interface{}) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error)
	Count(ctx context.Context) (int, error)
}

// SalesRepository defines the interface for sales data access
type SalesRepository interface {
	Create(ctx context.Context, sale interface{}) (string, error)
	GetByID(ctx context.Context, id string) (interface{}, error)
	List(ctx context.Context, limit, offset int) ([]interface{}, error)
	ListByCustomer(ctx context.Context, customerID string, limit, offset int) ([]interface{}, error)
	Update(ctx context.Context, id string, sale interface{}) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
	GetTotal(ctx context.Context) (float64, error)
}

// RepositoryFactory creates repositories based on operating mode
type RepositoryFactory interface {
	GetProductRepository() ProductRepository
	GetCustomerRepository() CustomerRepository
	GetSupplierRepository() SupplierRepository
	GetInventoryRepository() InventoryRepository
	GetSalesRepository() SalesRepository
}
