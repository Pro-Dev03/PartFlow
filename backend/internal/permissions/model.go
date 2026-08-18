package permissions

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`        // e.g., "products.read"
	Description string    `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"` // e.g., "products"
	Action      string    `json:"action" db:"action"`     // e.g., "read"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// PermissionRequest represents permission creation request
type PermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
}

// Define standard permissions
const (
	// Products
	ProductRead       = "products.read"
	ProductCreate     = "products.create"
	ProductUpdate     = "products.update"
	ProductDelete     = "products.delete"
	ProductArchive    = "products.archive"

	// Inventory
	InventoryRead      = "inventory.read"
	InventoryAdjust    = "inventory.adjust"
	InventoryTransfer  = "inventory.transfer"
	InventoryInspect   = "inventory.inspect"

	// Sales
	SaleRead      = "sales.read"
	SaleCreate    = "sales.create"
	SaleCancel    = "sales.cancel"
	SaleRefund    = "sales.refund"

	// Customers
	CustomerRead   = "customers.read"
	CustomerCreate = "customers.create"
	CustomerUpdate = "customers.update"
	CustomerDelete = "customers.delete"

	// Debts
	DebtRead   = "debts.read"
	DebtManage = "debts.manage"

	// Suppliers
	SupplierRead   = "suppliers.read"
	SupplierCreate = "suppliers.create"
	SupplierUpdate = "suppliers.update"
	SupplierDelete = "suppliers.delete"

	// Purchases
	PurchaseRead    = "purchases.read"
	PurchaseCreate  = "purchases.create"
	PurchaseReceive = "purchases.receive"

	// Expenses
	ExpenseRead   = "expenses.read"
	ExpenseCreate = "expenses.create"
	ExpenseUpdate = "expenses.update"
	ExpenseDelete = "expenses.delete"

	// Returns
	ReturnRead    = "returns.read"
	ReturnCreate  = "returns.create"
	ReturnApprove = "returns.approve"

	// Warranty
	WarrantyRead  = "warranties.read"
	WarrantyClaim = "warranties.claim"

	// Reports
	ReportRead   = "reports.read"
	ReportExport = "reports.export"

	// Users
	UserRead   = "users.read"
	UserCreate = "users.create"
	UserUpdate = "users.update"
	UserDelete = "users.delete"

	// Settings
	SettingsManage = "settings.manage"

	// Audit
	AuditRead = "audit.read"
)

// GetStandardPermissions returns all standard permissions
func GetStandardPermissions() []Permission {
	return []Permission{
		// Products
		{Name: ProductRead, Description: "Read products", Resource: "products", Action: "read"},
		{Name: ProductCreate, Description: "Create products", Resource: "products", Action: "create"},
		{Name: ProductUpdate, Description: "Update products", Resource: "products", Action: "update"},
		{Name: ProductDelete, Description: "Delete products", Resource: "products", Action: "delete"},
		{Name: ProductArchive, Description: "Archive products", Resource: "products", Action: "archive"},

		// Inventory
		{Name: InventoryRead, Description: "Read inventory", Resource: "inventory", Action: "read"},
		{Name: InventoryAdjust, Description: "Adjust inventory", Resource: "inventory", Action: "adjust"},
		{Name: InventoryTransfer, Description: "Transfer inventory", Resource: "inventory", Action: "transfer"},
		{Name: InventoryInspect, Description: "Inspect inventory items", Resource: "inventory", Action: "inspect"},

		// Sales
		{Name: SaleRead, Description: "Read sales", Resource: "sales", Action: "read"},
		{Name: SaleCreate, Description: "Create sales", Resource: "sales", Action: "create"},
		{Name: SaleCancel, Description: "Cancel sales", Resource: "sales", Action: "cancel"},
		{Name: SaleRefund, Description: "Refund sales", Resource: "sales", Action: "refund"},

		// Customers
		{Name: CustomerRead, Description: "Read customers", Resource: "customers", Action: "read"},
		{Name: CustomerCreate, Description: "Create customers", Resource: "customers", Action: "create"},
		{Name: CustomerUpdate, Description: "Update customers", Resource: "customers", Action: "update"},
		{Name: CustomerDelete, Description: "Delete customers", Resource: "customers", Action: "delete"},

		// Debts
		{Name: DebtRead, Description: "Read debts", Resource: "debts", Action: "read"},
		{Name: DebtManage, Description: "Manage debts", Resource: "debts", Action: "manage"},

		// Suppliers
		{Name: SupplierRead, Description: "Read suppliers", Resource: "suppliers", Action: "read"},
		{Name: SupplierCreate, Description: "Create suppliers", Resource: "suppliers", Action: "create"},
		{Name: SupplierUpdate, Description: "Update suppliers", Resource: "suppliers", Action: "update"},
		{Name: SupplierDelete, Description: "Delete suppliers", Resource: "suppliers", Action: "delete"},

		// Purchases
		{Name: PurchaseRead, Description: "Read purchases", Resource: "purchases", Action: "read"},
		{Name: PurchaseCreate, Description: "Create purchases", Resource: "purchases", Action: "create"},
		{Name: PurchaseReceive, Description: "Receive purchases", Resource: "purchases", Action: "receive"},

		// Expenses
		{Name: ExpenseRead, Description: "Read expenses", Resource: "expenses", Action: "read"},
		{Name: ExpenseCreate, Description: "Create expenses", Resource: "expenses", Action: "create"},
		{Name: ExpenseUpdate, Description: "Update expenses", Resource: "expenses", Action: "update"},
		{Name: ExpenseDelete, Description: "Delete expenses", Resource: "expenses", Action: "delete"},

		// Returns
		{Name: ReturnRead, Description: "Read returns", Resource: "returns", Action: "read"},
		{Name: ReturnCreate, Description: "Create returns", Resource: "returns", Action: "create"},
		{Name: ReturnApprove, Description: "Approve returns", Resource: "returns", Action: "approve"},

		// Warranty
		{Name: WarrantyRead, Description: "Read warranties", Resource: "warranties", Action: "read"},
		{Name: WarrantyClaim, Description: "Claim warranties", Resource: "warranties", Action: "claim"},

		// Reports
		{Name: ReportRead, Description: "Read reports", Resource: "reports", Action: "read"},
		{Name: ReportExport, Description: "Export reports", Resource: "reports", Action: "export"},

		// Users
		{Name: UserRead, Description: "Read users", Resource: "users", Action: "read"},
		{Name: UserCreate, Description: "Create users", Resource: "users", Action: "create"},
		{Name: UserUpdate, Description: "Update users", Resource: "users", Action: "update"},
		{Name: UserDelete, Description: "Delete users", Resource: "users", Action: "delete"},

		// Settings
		{Name: SettingsManage, Description: "Manage settings", Resource: "settings", Action: "manage"},

		// Audit
		{Name: AuditRead, Description: "Read audit logs", Resource: "audit", Action: "read"},
	}
}
