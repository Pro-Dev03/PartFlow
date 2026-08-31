package suppliers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles supplier data operations
type Repository struct {
	db *sqlx.DB
}

type localSupplierRow struct {
	ID             string         `db:"id"`
	Code           string         `db:"code"`
	Name           string         `db:"name"`
	Email          sql.NullString `db:"email"`
	Phone          sql.NullString `db:"phone"`
	Address        sql.NullString `db:"address"`
	City           sql.NullString `db:"city"`
	Country        sql.NullString `db:"country"`
	TaxID          sql.NullString `db:"tax_id"`
	PaymentTerms   sql.NullString `db:"payment_terms"`
	CreditLimit    float64        `db:"credit_limit"`
	CurrentBalance float64        `db:"current_balance"`
	TotalPurchases float64        `db:"total_purchases"`
	PaidAmount     float64        `db:"paid_amount"`
	Outstanding    float64        `db:"outstanding"`
	Notes          sql.NullString `db:"notes"`
	IsActive       int            `db:"is_active"`
	CreatedAt      string         `db:"created_at"`
	UpdatedAt      string         `db:"updated_at"`
}

func localSupplierFromRow(row localSupplierRow) (Supplier, error) {
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return Supplier{}, err
	}
	created, err := dbutil.ParseTimestamp(row.CreatedAt)
	if err != nil {
		return Supplier{}, err
	}
	updated, err := dbutil.ParseTimestamp(row.UpdatedAt)
	if err != nil {
		return Supplier{}, err
	}
	s := Supplier{ID: id, Code: row.Code, Name: row.Name, CreditLimit: row.CreditLimit, CurrentBalance: row.CurrentBalance, TotalPurchases: row.TotalPurchases, PaidAmount: row.PaidAmount, Outstanding: row.Outstanding, IsActive: row.IsActive != 0, CreatedAt: created, UpdatedAt: updated}
	for value, target := range map[*sql.NullString]**string{&row.Email: &s.Email, &row.Phone: &s.Phone, &row.Address: &s.Address, &row.City: &s.City, &row.Country: &s.Country, &row.TaxID: &s.TaxID, &row.PaymentTerms: &s.PaymentTerms, &row.Notes: &s.Notes} {
		if value.Valid && value.String != "" {
			v := value.String
			*target = &v
		}
	}
	return s, nil
}

func localSupplierQuery() string {
	return `SELECT s.id, s.code, s.name, s.email, s.phone, s.address, s.city, s.country, s.tax_id, s.payment_terms, s.credit_limit, s.current_balance, COALESCE((SELECT SUM(total_amount) FROM purchases p WHERE p.supplier_id = s.id AND p.status NOT IN ('cancelled','reversed')), 0) AS total_purchases, COALESCE((SELECT SUM(COALESCE(paid_amount,0)) FROM purchases p WHERE p.supplier_id = s.id AND p.status NOT IN ('cancelled','reversed')), 0) AS paid_amount, COALESCE((SELECT SUM(CASE WHEN total_amount - COALESCE(paid_amount,0) > 0 THEN total_amount - COALESCE(paid_amount,0) ELSE 0 END) FROM purchases p WHERE p.supplier_id = s.id AND p.status NOT IN ('cancelled','reversed')), 0) AS outstanding, s.notes, s.is_active, s.created_at, s.updated_at FROM suppliers s`
}

// NewRepository creates a new supplier repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func scanSupplier(row interface{ Scan(...any) error }) (Supplier, error) {
	var supplier Supplier
	var email, phone, address, city, country, taxID, paymentTerms, notes sql.NullString
	var createdAt, updatedAt string
	if err := row.Scan(
		&supplier.ID, &supplier.Code, &supplier.Name, &email, &phone, &address, &city,
		&country, &taxID, &paymentTerms, &supplier.CreditLimit, &supplier.CurrentBalance,
		&supplier.TotalPurchases, &supplier.PaidAmount, &supplier.Outstanding, &notes,
		&supplier.IsActive, &createdAt, &updatedAt,
	); err != nil {
		return Supplier{}, err
	}
	for value, target := range map[*sql.NullString]**string{
		&email: &supplier.Email, &phone: &supplier.Phone, &address: &supplier.Address,
		&city: &supplier.City, &country: &supplier.Country, &taxID: &supplier.TaxID,
		&paymentTerms: &supplier.PaymentTerms, &notes: &supplier.Notes,
	} {
		if value.Valid && strings.TrimSpace(value.String) != "" {
			copied := value.String
			*target = &copied
		}
	}
	var err error
	supplier.CreatedAt, err = parseSQLiteTimestamp(createdAt)
	if err != nil {
		return Supplier{}, fmt.Errorf("parse created_at: %w", err)
	}
	supplier.UpdatedAt, err = parseSQLiteTimestamp(updatedAt)
	if err != nil {
		return Supplier{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return supplier, nil
}

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported SQLite timestamp format: %q", raw)
}

// Create creates a new supplier
func (r *Repository) Create(ctx context.Context, supplier *Supplier) error {
	query := `
		INSERT INTO suppliers (id, code, name, email, phone, address, city, country, tax_id,
			payment_terms, credit_limit, current_balance, notes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		supplier.ID, supplier.Code, supplier.Name, supplier.Email, supplier.Phone, supplier.Address, supplier.City,
		supplier.Country, supplier.TaxID, supplier.PaymentTerms, supplier.CreditLimit,
		supplier.CurrentBalance, supplier.Notes, supplier.IsActive, supplier.CreatedAt, supplier.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}
	return nil
}

// GetByID retrieves a supplier by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Supplier, error) {
	if dbutil.IsSQLite(r.db) {
		var row localSupplierRow
		if err := r.db.GetContext(ctx, &row, localSupplierQuery()+` WHERE s.id = $1`, id); err != nil {
			return nil, ErrSupplierNotFound
		}
		s, err := localSupplierFromRow(row)
		if err != nil {
			return nil, err
		}
		return &s, nil
	}
	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id,
			payment_terms, credit_limit, current_balance,
			COALESCE((SELECT SUM(total_amount) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS total_purchases,
			COALESCE((SELECT SUM(COALESCE(paid_amount, 0)) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS paid_amount,
			COALESCE((SELECT SUM(CASE WHEN total_amount - COALESCE(paid_amount, 0) > 0 THEN total_amount - COALESCE(paid_amount, 0) ELSE 0 END) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS outstanding,
			notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE id = $1
	`
	var supplier Supplier
	err := r.db.GetContext(ctx, &supplier, query, id)
	if err != nil {
		return nil, ErrSupplierNotFound
	}
	return &supplier, nil
}

// GetByCode retrieves a supplier by code
func (r *Repository) GetByCode(ctx context.Context, code string) (*Supplier, error) {
	if dbutil.IsSQLite(r.db) {
		var row localSupplierRow
		if err := r.db.GetContext(ctx, &row, localSupplierQuery()+` WHERE s.code = $1`, code); err != nil {
			return nil, ErrSupplierNotFound
		}
		s, err := localSupplierFromRow(row)
		if err != nil {
			return nil, err
		}
		return &s, nil
	}
	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id,
			payment_terms, credit_limit, current_balance,
			COALESCE((SELECT SUM(total_amount) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS total_purchases,
			COALESCE((SELECT SUM(COALESCE(paid_amount, 0)) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS paid_amount,
			COALESCE((SELECT SUM(CASE WHEN total_amount - COALESCE(paid_amount, 0) > 0 THEN total_amount - COALESCE(paid_amount, 0) ELSE 0 END) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS outstanding,
			notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE code = $1
	`
	var supplier Supplier
	err := r.db.GetContext(ctx, &supplier, query, code)
	if err != nil {
		return nil, ErrSupplierNotFound
	}
	return &supplier, nil
}

// List retrieves suppliers with pagination and filters
func (r *Repository) List(ctx context.Context, page, perPage int, search string, isActive *bool) ([]Supplier, int, error) {
	if dbutil.IsSQLite(r.db) {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 || perPage > 100 {
			perPage = 20
		}
		base := localSupplierQuery() + ` WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM suppliers WHERE 1=1`
		args := []interface{}{}
		n := 0
		if search != "" {
			n++
			condition := fmt.Sprintf(` AND (LOWER(s.name) LIKE LOWER($%d) OR LOWER(s.code) LIKE LOWER($%d) OR LOWER(COALESCE(s.email,'')) LIKE LOWER($%d) OR LOWER(COALESCE(s.phone,'')) LIKE LOWER($%d))`, n, n, n, n)
			base += condition
			countQuery += strings.Replace(condition, "s.", "", -1)
			args = append(args, "%"+search+"%")
		}
		if isActive != nil {
			n++
			base += fmt.Sprintf(" AND s.is_active = $%d", n)
			countQuery += fmt.Sprintf(" AND is_active = $%d", n)
			value := 0
			if *isActive {
				value = 1
			}
			args = append(args, value)
		}
		var total int
		if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
			return nil, 0, err
		}
		base += fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", n+1, n+2)
		args = append(args, perPage, (page-1)*perPage)
		var rows []localSupplierRow
		if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
			return nil, 0, err
		}
		out := make([]Supplier, 0, len(rows))
		for _, row := range rows {
			s, err := localSupplierFromRow(row)
			if err != nil {
				return nil, 0, err
			}
			out = append(out, s)
		}
		return out, total, nil
	}
	offset := (page - 1) * perPage

	query := `
		SELECT id, code, name, email, phone, address, city, country, tax_id,
			payment_terms, credit_limit, current_balance,
			COALESCE((SELECT SUM(total_amount) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS total_purchases,
			COALESCE((SELECT SUM(COALESCE(paid_amount, 0)) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS paid_amount,
			COALESCE((SELECT SUM(CASE WHEN total_amount - COALESCE(paid_amount, 0) > 0 THEN total_amount - COALESCE(paid_amount, 0) ELSE 0 END) FROM purchases WHERE supplier_id = suppliers.id AND status NOT IN ('cancelled', 'reversed')), 0) AS outstanding,
			notes, is_active, created_at, updated_at
		FROM suppliers
		WHERE 1=1
	`
	countQuery := `
		SELECT COUNT(*) FROM suppliers WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (LOWER(name) LIKE LOWER($%d) OR LOWER(code) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d) OR LOWER(phone) LIKE LOWER($%d))", argCount, argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(name) LIKE LOWER($%d) OR LOWER(code) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d) OR LOWER(phone) LIKE LOWER($%d))", argCount, argCount, argCount, argCount)
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		argCount += 3
	}

	if isActive != nil {
		argCount++
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		countQuery += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, *isActive)
	}

	// Get total count
	countArgs := []interface{}{}
	countArgCount := 0

	if search != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND (LOWER(name) LIKE LOWER($%d) OR LOWER(code) LIKE LOWER($%d) OR LOWER(email) LIKE LOWER($%d) OR LOWER(phone) LIKE LOWER($%d))", countArgCount, countArgCount, countArgCount, countArgCount)
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern, searchPattern)
		countArgCount += 3
	}

	if isActive != nil {
		countArgCount++
		countQuery += fmt.Sprintf(" AND is_active = $%d", countArgCount)
		countArgs = append(countArgs, *isActive)
	}

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}
	defer rows.Close()
	suppliers := make([]Supplier, 0)
	for rows.Next() {
		supplier, scanErr := scanSupplier(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("failed to scan supplier: %w", scanErr)
		}
		suppliers = append(suppliers, supplier)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate suppliers: %w", err)
	}

	return suppliers, total, nil
}

// Update updates a supplier
func (r *Repository) Update(ctx context.Context, supplier *Supplier) error {
	query := `
		UPDATE suppliers
		SET code = $2, name = $3, email = $4, phone = $5, address = $6, city = $7, country = $8, tax_id = $9, payment_terms = $10, credit_limit = $11, notes = $12, is_active = $13, updated_at = $14
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		supplier.ID, supplier.Code, supplier.Name, supplier.Email, supplier.Phone,
		supplier.Address, supplier.City, supplier.Country, supplier.TaxID, supplier.PaymentTerms,
		supplier.CreditLimit, supplier.Notes, supplier.IsActive, supplier.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSupplierNotFound
	}

	return nil
}

// Delete deletes a supplier
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id = $1)`, id)
	if err != nil {
		return fmt.Errorf("failed to find supplier: %w", err)
	}
	if !exists {
		return ErrSupplierNotFound
	}

	query := fmt.Sprintf(`UPDATE suppliers SET is_active = false, updated_at = %s WHERE id = $1`, dbutil.NowSQL(r.db))
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate supplier: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSupplierNotFound
	}

	return nil
}

// HasActivePurchases checks if supplier has active purchases
func (r *Repository) HasActivePurchases(ctx context.Context, supplierID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM purchases 
			WHERE supplier_id = $1 
			AND status != 'cancelled'
			AND created_at > %s
		)
	`
	if dbutil.IsSQLite(r.db) {
		query = fmt.Sprintf(query, "datetime('now','-1 year')")
	} else {
		query = fmt.Sprintf(query, "NOW() - INTERVAL '1 year'")
	}

	var hasPurchases bool
	err := r.db.GetContext(ctx, &hasPurchases, query, supplierID)
	if err != nil {
		// If table doesn't exist, treat as no purchases
		if err.Error() == `pq: relation "purchases" does not exist` {
			return false, nil
		}
		return false, fmt.Errorf("failed to check active purchases: %w", err)
	}

	return hasPurchases, nil
}

// UpdateBalance updates supplier balance
func (r *Repository) UpdateBalance(ctx context.Context, supplierID uuid.UUID, amount float64) error {
	query := fmt.Sprintf(`
		UPDATE suppliers
		SET current_balance = current_balance + $1, updated_at = %s
		WHERE id = $2
	`, dbutil.NowSQL(r.db))
	result, err := r.db.ExecContext(ctx, query, amount, supplierID)
	if err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrSupplierNotFound
	}

	return nil
}

// GetSupplierLedger retrieves supplier ledger entries
func (r *Repository) GetSupplierLedger(ctx context.Context, supplierID uuid.UUID) ([]LedgerEntry, float64, float64, float64, error) {
	// Get ledger entries
	query := `
		SELECT id, supplier_id, type, amount, balance, description, reference_id, created_at
		FROM supplier_ledger
		WHERE supplier_id = $1
		ORDER BY created_at ASC
	`
	var entries []LedgerEntry
	var err error
	if dbutil.IsSQLite(r.db) {
		var rows []struct {
			ID              string         `db:"id"`
			SupplierID      string         `db:"supplier_id"`
			Type            sql.NullString `db:"type"`
			TransactionType sql.NullString `db:"transaction_type"`
			Amount, Balance float64
			Description     sql.NullString `db:"description"`
			ReferenceID     sql.NullString `db:"reference_id"`
			CreatedAt       string         `db:"created_at"`
		}
		err = r.db.SelectContext(ctx, &rows, `SELECT id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at FROM supplier_ledger WHERE supplier_id = $1 ORDER BY created_at ASC`, supplierID)
		if err == nil {
			entries = make([]LedgerEntry, 0, len(rows))
			for _, row := range rows {
				id, e := uuid.Parse(row.ID)
				if e != nil {
					return nil, 0, 0, 0, e
				}
				sid, e := uuid.Parse(row.SupplierID)
				if e != nil {
					return nil, 0, 0, 0, e
				}
				created, e := dbutil.ParseTimestamp(row.CreatedAt)
				if e != nil {
					return nil, 0, 0, 0, e
				}
				typ := row.Type.String
				if typ == "" {
					typ = strings.ToLower(row.TransactionType.String)
				}
				entry := LedgerEntry{ID: id, SupplierID: sid, Type: typ, Amount: row.Amount, Balance: row.Balance, CreatedAt: created}
				if row.Description.Valid {
					entry.Description = row.Description.String
				}
				if row.ReferenceID.Valid && row.ReferenceID.String != "" {
					if ref, e := uuid.Parse(row.ReferenceID.String); e == nil {
						entry.ReferenceID = &ref
					}
				}
				entries = append(entries, entry)
			}
		}
	} else {
		err = r.db.SelectContext(ctx, &entries, query, supplierID)
	}
	if err != nil {
		// If table doesn't exist, return empty ledger
		if err.Error() == `pq: relation "supplier_ledger" does not exist` {
			return []LedgerEntry{}, 0, 0, 0, nil
		}
		return nil, 0, 0, 0, fmt.Errorf("failed to get supplier ledger: %w", err)
	}

	// Get totals
	var totalPurchases, totalPayments, currentBalance float64
	if dbutil.IsSQLite(r.db) {
		query = `SELECT COALESCE(SUM(CASE WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN type = 'credit' OR transaction_type IN ('PAYMENT','RETURN') THEN amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN type = 'debit' OR transaction_type = 'PURCHASE' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $1`
	} else {
		query = `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE 0 END), 0) as total_purchases,
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0) as total_payments,
			COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) as current_balance
		FROM supplier_ledger
		WHERE supplier_id = $1
		`
	}
	err = r.db.QueryRowContext(ctx, query, supplierID).Scan(&totalPurchases, &totalPayments, &currentBalance)
	if err != nil {
		// If table doesn't exist, return empty totals
		if err.Error() == `pq: relation "supplier_ledger" does not exist` {
			return entries, 0, 0, 0, nil
		}
		return nil, 0, 0, 0, fmt.Errorf("failed to get supplier totals: %w", err)
	}

	return entries, totalPurchases, totalPayments, currentBalance, nil
}

// AddPayment adds a payment to supplier ledger
func (r *Repository) AddPayment(ctx context.Context, payment *PaymentResponse) error {
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `INSERT INTO payments (id, transaction_number, supplier_id, amount, payment_method, reference, notes, payment_date, payment_status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'completed', $9, $9)`, payment.ID, "PAY-"+payment.ID.String()[:8], payment.SupplierID, payment.Amount, payment.Method, payment.Reference, payment.Notes, payment.PaymentDate, payment.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to add local payment: %w", err)
		}
		_, err = r.db.ExecContext(ctx, `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT $1, $2, 'credit', 'PAYMENT', $3, COALESCE((SELECT balance FROM supplier_ledger WHERE supplier_id = $2 ORDER BY created_at DESC LIMIT 1), 0) - $3, $4, $5, $6`, uuid.New(), payment.SupplierID, payment.Amount, "Payment: "+payment.Method, payment.ID, payment.CreatedAt)
		return err
	}
	query := `
		INSERT INTO supplier_payments (id, supplier_id, amount, payment_date, method, reference, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.SupplierID, payment.Amount, payment.PaymentDate,
		payment.Method, payment.Reference, payment.Notes, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add payment: %w", err)
	}

	// Add to ledger
	ledgerQuery := `
		INSERT INTO supplier_ledger (id, supplier_id, type, amount, balance, description, reference_id, created_at)
		SELECT $1, $2, 'credit', $3, 
			(SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $2) - $3,
			$4, $5, $6
	`
	_, err = r.db.ExecContext(ctx, ledgerQuery,
		uuid.New(), payment.SupplierID, payment.Amount,
		"Payment: "+payment.Method, payment.ID, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	return nil
}

// AddLedgerEntry adds a ledger entry
func (r *Repository) AddLedgerEntry(ctx context.Context, supplierID uuid.UUID, entryType string, amount float64, description string, referenceID uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) SELECT $1, $2, $3, CASE WHEN $3 = 'debit' THEN 'PURCHASE' ELSE 'ADJUSTMENT' END, $4, COALESCE((SELECT balance FROM supplier_ledger WHERE supplier_id = $2 ORDER BY created_at DESC LIMIT 1), 0) + CASE WHEN $3 = 'debit' THEN $4 ELSE -$4 END, $5, $6, CURRENT_TIMESTAMP`, uuid.New(), supplierID, entryType, amount, description, referenceID)
		return err
	}
	ledgerQuery := `
		INSERT INTO supplier_ledger (id, supplier_id, type, amount, balance, description, reference_id, created_at)
		SELECT $1, $2, $3, $4, 
			(SELECT COALESCE(SUM(CASE WHEN type = 'debit' THEN amount ELSE -amount END), 0) FROM supplier_ledger WHERE supplier_id = $2) + 
			CASE WHEN $3 = 'debit' THEN $4 ELSE -$4 END,
			$5, $6, $7
	`
	_, err := r.db.ExecContext(ctx, ledgerQuery,
		uuid.New(), supplierID, entryType, amount, description, referenceID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	return nil
}

// CreateDebtEntry creates a new debt entry
func (r *Repository) CreateDebtEntry(ctx context.Context, debt *DebtEntry) error {
	if dbutil.IsSQLite(r.db) {
		if _, err := r.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS supplier_debts (id TEXT PRIMARY KEY, supplier_id TEXT NOT NULL, amount REAL NOT NULL, reference_id TEXT, reference_type TEXT, due_date TEXT, is_paid INTEGER DEFAULT 0, paid_amount REAL DEFAULT 0, created_at TEXT NOT NULL)`); err != nil {
			return err
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO supplier_debts (id, supplier_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, debt.ID, debt.SupplierID, debt.Amount, debt.ReferenceID, debt.ReferenceType, debt.DueDate, debt.IsPaid, debt.PaidAmount, debt.CreatedAt)
		return err
	}
	query := `
		INSERT INTO supplier_debts (id, supplier_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		debt.ID, debt.SupplierID, debt.Amount, debt.ReferenceID, debt.ReferenceType,
		debt.DueDate, debt.IsPaid, debt.PaidAmount, debt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create debt entry: %w", err)
	}
	return nil
}

// GetDebtEntries retrieves debt entries for a supplier
func (r *Repository) GetDebtEntries(ctx context.Context, supplierID uuid.UUID) ([]DebtEntry, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []struct {
			ID                                             string `db:"id"`
			SupplierID                                     string `db:"supplier_id"`
			Amount, Paid                                   float64
			ReferenceID, ReferenceType, DueDate, CreatedAt string
			IsPaid                                         int
		}
		if err := r.db.SelectContext(ctx, &rows, `SELECT id, supplier_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at FROM supplier_debts WHERE supplier_id = $1 ORDER BY due_date ASC`, supplierID); err != nil {
			return nil, err
		}
		out := make([]DebtEntry, 0, len(rows))
		for _, row := range rows {
			id, e := uuid.Parse(row.ID)
			if e != nil {
				return nil, e
			}
			sid, e := uuid.Parse(row.SupplierID)
			if e != nil {
				return nil, e
			}
			ref, _ := uuid.Parse(row.ReferenceID)
			due, e := dbutil.ParseTimestamp(row.DueDate)
			if e != nil {
				return nil, e
			}
			created, e := dbutil.ParseTimestamp(row.CreatedAt)
			if e != nil {
				return nil, e
			}
			out = append(out, DebtEntry{ID: id, SupplierID: sid, Amount: row.Amount, PaidAmount: row.Paid, ReferenceID: ref, ReferenceType: row.ReferenceType, DueDate: due, IsPaid: row.IsPaid != 0, CreatedAt: created})
		}
		return out, nil
	}
	query := `
		SELECT id, supplier_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at
		FROM supplier_debts
		WHERE supplier_id = $1
		ORDER BY due_date ASC
	`
	var debts []DebtEntry
	err := r.db.SelectContext(ctx, &debts, query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt entries: %w", err)
	}
	return debts, nil
}

// UpdateDebtPayment updates payment for a debt entry
func (r *Repository) UpdateDebtPayment(ctx context.Context, debtID uuid.UUID, paymentAmount float64) error {
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, `UPDATE supplier_debts SET paid_amount = MIN(amount, paid_amount + $1), is_paid = CASE WHEN paid_amount + $1 >= amount THEN 1 ELSE 0 END WHERE id = $2`, paymentAmount, debtID)
		return err
	}
	query := `
		UPDATE supplier_debts
		SET paid_amount = paid_amount + $1,
			is_paid = (paid_amount + $1) >= amount
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, paymentAmount, debtID)
	if err != nil {
		return fmt.Errorf("failed to update debt payment: %w", err)
	}
	return nil
}

// CreateDebtCollection creates a new debt collection action
func (r *Repository) CreateDebtCollection(ctx context.Context, collection *DebtCollection) error {
	query := `
		INSERT INTO supplier_debt_collections (id, supplier_id, type, status, notes, scheduled_date, completed_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		collection.ID, collection.SupplierID, collection.Type, collection.Status,
		collection.Notes, collection.ScheduledDate, collection.CompletedDate, collection.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create debt collection: %w", err)
	}
	return nil
}

// GetDebtCollections retrieves debt collection actions for a supplier
func (r *Repository) GetDebtCollections(ctx context.Context, supplierID uuid.UUID) ([]DebtCollection, error) {
	query := `
		SELECT id, supplier_id, type, status, notes, scheduled_date, completed_date, created_at
		FROM supplier_debt_collections
		WHERE supplier_id = $1
		ORDER BY scheduled_date DESC
	`
	var collections []DebtCollection
	err := r.db.SelectContext(ctx, &collections, query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get debt collections: %w", err)
	}
	return collections, nil
}

// GetPendingDebtCollections retrieves pending debt collection actions
func (r *Repository) GetPendingDebtCollections(ctx context.Context) ([]DebtCollection, error) {
	query := `
		SELECT dc.id, dc.supplier_id, dc.type, dc.status, dc.notes, dc.scheduled_date, dc.completed_date, dc.created_at
		FROM supplier_debt_collections dc
		JOIN suppliers s ON dc.supplier_id = s.id
		WHERE dc.status = 'pending'
		AND dc.scheduled_date <= NOW()
		ORDER BY dc.scheduled_date ASC
	`
	var collections []DebtCollection
	err := r.db.SelectContext(ctx, &collections, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending debt collections: %w", err)
	}
	return collections, nil
}
