package ledgers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	db         *sqlx.DB
	repository *Repository
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		db:         db,
		repository: NewRepository(db),
	}
}

// CreateLedgerEntry creates a new ledger entry with automatic balance calculation
func (s *Service) CreateLedgerEntry(ctx context.Context, req *LedgerEntryRequest, userID uuid.UUID) (*LedgerEntry, error) {
	// Get current balance using repository
	currentBalance, err := s.repository.GetCurrentBalance(ctx, req.LedgerType, req.EntityID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	// Calculate new balance
	newBalance := currentBalance + req.Amount

	// Create ledger entry
	entry := &LedgerEntry{
		ID:             uuid.New(),
		LedgerType:     req.LedgerType,
		EntityID:       req.EntityID,
		TransactionType: req.TransactionType,
		ReferenceID:    req.ReferenceID,
		ReferenceType:  req.ReferenceType,
		Amount:         req.Amount,
		Balance:        newBalance,
		PreviousBalance: currentBalance,
		Description:    req.Description,
		Metadata:       req.Metadata,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	// Insert using repository (will trigger automatic balance updates)
	err = s.repository.CreateLedgerEntry(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger entry: %w", err)
	}

	return entry, nil
}

// GetCustomerLedgerSummary retrieves customer ledger summary
func (s *Service) GetCustomerLedgerSummary(ctx context.Context, customerID uuid.UUID) (*CustomerLedger, error) {
	return s.repository.GetCustomerLedgerSummary(ctx, customerID)
}

// GetSupplierLedgerSummary retrieves supplier ledger summary
func (s *Service) GetSupplierLedgerSummary(ctx context.Context, supplierID uuid.UUID) (*SupplierLedger, error) {
	return s.repository.GetSupplierLedgerSummary(ctx, supplierID)
}

// GetInventoryLedgerSummary retrieves inventory ledger summary
func (s *Service) GetInventoryLedgerSummary(ctx context.Context, productID uuid.UUID) (*InventoryLedger, error) {
	return s.repository.GetInventoryLedgerSummary(ctx, productID)
}

// GetLedgerEntries retrieves ledger entries for an entity with pagination
func (s *Service) GetLedgerEntries(ctx context.Context, ledgerType LedgerType, entityID uuid.UUID, page, perPage int) ([]*LedgerEntry, int64, error) {
	return s.repository.GetLedgerEntries(ctx, ledgerType, entityID, page, perPage)
}

// GetOverdueEntities retrieves entities with overdue balances
func (s *Service) GetOverdueEntities(ctx context.Context, ledgerType LedgerType) ([]uuid.UUID, error) {
	return s.repository.GetOverdueEntities(ctx, ledgerType)
}

// CreateSaleLedgerEntry creates a ledger entry for a sale
func (s *Service) CreateSaleLedgerEntry(ctx context.Context, customerID uuid.UUID, saleID uuid.UUID, amount float64, userID uuid.UUID) error {
	refType := "sale"
	req := &LedgerEntryRequest{
		LedgerType:      LedgerTypeCustomer,
		EntityID:        customerID,
		TransactionType: TransactionSale,
		ReferenceID:     &saleID,
		ReferenceType:   &refType,
		Amount:          amount, // positive for debit (customer owes money)
		Description:     fmt.Sprintf("Sale #%s", saleID),
	}
	_, err := s.CreateLedgerEntry(ctx, req, userID)
	return err
}

// CreatePaymentLedgerEntry creates a ledger entry for a payment
func (s *Service) CreatePaymentLedgerEntry(ctx context.Context, customerID uuid.UUID, paymentID uuid.UUID, amount float64, userID uuid.UUID) error {
	refType := "payment"
	req := &LedgerEntryRequest{
		LedgerType:      LedgerTypeCustomer,
		EntityID:        customerID,
		TransactionType: TransactionPayment,
		ReferenceID:     &paymentID,
		ReferenceType:   &refType,
		Amount:          -amount, // negative for credit (payment reduces debt)
		Description:     fmt.Sprintf("Payment #%s", paymentID),
	}
	_, err := s.CreateLedgerEntry(ctx, req, userID)
	return err
}

// CreatePurchaseLedgerEntry creates a ledger entry for a purchase
func (s *Service) CreatePurchaseLedgerEntry(ctx context.Context, supplierID uuid.UUID, purchaseID uuid.UUID, amount float64, userID uuid.UUID) error {
	refType := "purchase"
	req := &LedgerEntryRequest{
		LedgerType:      LedgerTypeSupplier,
		EntityID:        supplierID,
		TransactionType: TransactionPurchase,
		ReferenceID:     &purchaseID,
		ReferenceType:   &refType,
		Amount:          amount, // positive for debit (we owe supplier)
		Description:     fmt.Sprintf("Purchase #%s", purchaseID),
	}
	_, err := s.CreateLedgerEntry(ctx, req, userID)
	return err
}

// CreateStockInLedgerEntry creates a ledger entry for stock received
func (s *Service) CreateStockInLedgerEntry(ctx context.Context, productID uuid.UUID, quantity float64, referenceID uuid.UUID, userID uuid.UUID) error {
	refType := "purchase"
	req := &LedgerEntryRequest{
		LedgerType:      LedgerTypeInventory,
		EntityID:        productID,
		TransactionType: TransactionStockIn,
		ReferenceID:     &referenceID,
		ReferenceType:   &refType,
		Amount:          quantity, // positive for stock in
		Description:     fmt.Sprintf("Stock received via purchase #%s", referenceID),
		Metadata:        map[string]interface{}{"quantity": quantity},
	}
	_, err := s.CreateLedgerEntry(ctx, req, userID)
	return err
}

// CreateStockOutLedgerEntry creates a ledger entry for stock sold
func (s *Service) CreateStockOutLedgerEntry(ctx context.Context, productID uuid.UUID, quantity float64, referenceID uuid.UUID, userID uuid.UUID) error {
	refType := "sale"
	req := &LedgerEntryRequest{
		LedgerType:      LedgerTypeInventory,
		EntityID:        productID,
		TransactionType: TransactionStockOut,
		ReferenceID:     &referenceID,
		ReferenceType:   &refType,
		Amount:          -quantity, // negative for stock out
		Description:     fmt.Sprintf("Stock sold via sale #%s", referenceID),
		Metadata:        map[string]interface{}{"quantity": quantity},
	}
	_, err := s.CreateLedgerEntry(ctx, req, userID)
	return err
}
