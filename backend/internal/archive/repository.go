package archive

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) tableExists(ctx context.Context, table string) bool {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`
	if dbutil.IsSQLite(r.db) {
		query = `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`
	}
	if err := r.db.GetContext(ctx, &count, query, table); err != nil {
		return false
	}
	return count > 0
}

func (r *Repository) legacyLedgerSources(ctx context.Context) string {
	parts := make([]string, 0, 2)
	if r.tableExists(ctx, "customer_ledger") {
		parts = append(parts, `
			SELECT 'customer_ledger:' || CAST(cl.id AS TEXT) AS event_id,
				COALESCE(cl.transaction_type, cl.type, 'ledger_entry') AS event_type,
				'customers' AS section, 'customer' AS entity_type,
				CAST(cl.customer_id AS TEXT) AS entity_id,
				COALESCE(cl.reference_type, '') AS reference_type,
				COALESCE(CAST(cl.reference_id AS TEXT), '') AS reference_id,
				COALESCE(CAST(cl.created_by AS TEXT), '') AS user_id, '' AS user_name,
				'completed' AS status,
				COALESCE(cl.description, cl.transaction_type, cl.type, 'معاملة عميل') AS description,
				'' AS details, CAST(cl.amount AS REAL) AS quantity, NULL AS before_value,
				CAST(cl.balance AS REAL) AS after_value, cl.created_at
			FROM customer_ledger cl`)
	}
	if r.tableExists(ctx, "supplier_ledger") {
		parts = append(parts, `
			SELECT 'supplier_ledger:' || CAST(sl.id AS TEXT) AS event_id,
				COALESCE(sl.transaction_type, sl.type, 'ledger_entry') AS event_type,
				'suppliers' AS section, 'supplier' AS entity_type,
				CAST(sl.supplier_id AS TEXT) AS entity_id,
				COALESCE(sl.reference_type, '') AS reference_type,
				COALESCE(CAST(sl.reference_id AS TEXT), '') AS reference_id,
				COALESCE(CAST(sl.created_by AS TEXT), '') AS user_id, '' AS user_name,
				'completed' AS status,
				COALESCE(sl.description, sl.transaction_type, sl.type, 'معاملة مورد') AS description,
				'' AS details, CAST(sl.amount AS REAL) AS quantity, NULL AS before_value,
				CAST(sl.balance AS REAL) AS after_value, sl.created_at
			FROM supplier_ledger sl`)
	}
	if len(parts) == 0 {
		return ""
	}
	return "\n\n\t\t\tUNION ALL" + strings.Join(parts, "\n\n\t\t\tUNION ALL")
}

func (r *Repository) auditDetailsExpression(ctx context.Context) string {
	if dbutil.IsSQLite(r.db) {
		var hasChanges int
		if err := r.db.GetContext(ctx, &hasChanges, `SELECT COUNT(*) FROM pragma_table_info('audit_logs') WHERE name = 'changes'`); err == nil && hasChanges > 0 {
			return "COALESCE(COALESCE(al.new_values, al.changes), '')"
		}
		var hasNewValues int
		if err := r.db.GetContext(ctx, &hasNewValues, `SELECT COUNT(*) FROM pragma_table_info('audit_logs') WHERE name = 'new_values'`); err == nil && hasNewValues > 0 {
			return "COALESCE(al.new_values, '')"
		}
		return "''"
	}

	var hasChanges int
	if err := r.db.GetContext(ctx, &hasChanges, `SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'audit_logs' AND column_name = 'changes'`); err == nil && hasChanges > 0 {
		return "COALESCE(COALESCE(al.new_values, al.changes), '')"
	}
	var hasNewValues int
	if err := r.db.GetContext(ctx, &hasNewValues, `SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'audit_logs' AND column_name = 'new_values'`); err == nil && hasNewValues > 0 {
		return "COALESCE(al.new_values, '')"
	}
	return "''"
}

func (r *Repository) List(ctx context.Context, req ListRequest) ([]Event, int, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 || req.PerPage > 100 {
		req.PerPage = 25
	}

	likeOperator := "ILIKE"
	if dbutil.IsSQLite(r.db) {
		likeOperator = "LIKE"
	}

	auditDetailsSQL := r.auditDetailsExpression(ctx)

	base := fmt.Sprintf(`
		SELECT event_id, event_type, section, entity_type, entity_id, reference_type, reference_id,
		       user_id, user_name, status, description, details, quantity, before_value, after_value, created_at
		FROM (
			SELECT
				'audit:' || CAST(al.id AS TEXT) AS event_id,
				al.action AS event_type,
				CASE
					WHEN LOWER(al.entity_type) IN ('sale', 'sales') THEN 'sales'
					WHEN LOWER(al.entity_type) IN ('purchase', 'purchases') THEN 'purchases'
					WHEN LOWER(al.entity_type) IN ('inventory', 'inventory_item', 'stock') THEN 'inventory'
					WHEN LOWER(al.entity_type) IN ('return', 'returns') THEN 'returns'
					WHEN LOWER(al.entity_type) IN ('customer', 'supplier', 'payment') THEN LOWER(al.entity_type)
					WHEN LOWER(al.entity_type) IN ('customer_ledger', 'customerledger') THEN 'customers'
					WHEN LOWER(al.entity_type) IN ('supplier_ledger', 'supplierledger') THEN 'suppliers'
					ELSE 'audit'
				END AS section,
				al.entity_type, CAST(al.entity_id AS TEXT) AS entity_id,
				'' AS reference_type, '' AS reference_id,
				COALESCE(CAST(al.user_id AS TEXT), '') AS user_id,
				COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), u.email, CAST(al.user_id AS TEXT), '') AS user_name,
				COALESCE(al.status, 'success') AS status,
				COALESCE(al.description, al.action, 'عملية مسجلة') AS description,
				%s AS details,
				NULL AS quantity, NULL AS before_value, NULL AS after_value, al.created_at
			FROM audit_logs al
			LEFT JOIN users u ON CAST(u.id AS TEXT) = CAST(al.user_id AS TEXT)

			UNION ALL

			SELECT
				'movement:' || CAST(im.id AS TEXT) AS event_id,
				im.movement_type AS event_type,
				'inventory' AS section,
				CASE WHEN im.item_id IS NOT NULL AND CAST(im.item_id AS TEXT) <> '' THEN 'inventory_item' ELSE 'product' END AS entity_type,
				COALESCE(NULLIF(CAST(im.item_id AS TEXT), ''), CAST(im.product_id AS TEXT), '') AS entity_id,
				COALESCE(im.reference_type, '') AS reference_type,
				COALESCE(CAST(im.reference_id AS TEXT), '') AS reference_id,
				COALESCE(CAST(im.created_by AS TEXT), '') AS user_id,
				COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), u.email, CAST(im.created_by AS TEXT), '') AS user_name,
				CASE WHEN LOWER(CAST(COALESCE(im.is_reversed, FALSE) AS TEXT)) IN ('1', 't', 'true') THEN 'reversed' ELSE 'completed' END AS status,
				COALESCE(im.reason, im.movement_type, 'حركة مخزون') AS description,
				'' AS details,
				CAST(im.quantity AS REAL) AS quantity,
				CAST(im.before_quantity AS REAL) AS before_value,
				CAST(im.after_quantity AS REAL) AS after_value,
				im.created_at
			FROM inventory_movements im
			LEFT JOIN users u ON CAST(u.id AS TEXT) = CAST(im.created_by AS TEXT)

			UNION ALL

			SELECT
				'item_history:' || CAST(ih.id AS TEXT) AS event_id,
				COALESCE(ih.event_type, 'inventory_history') AS event_type,
				'inventory' AS section,
				CASE
					WHEN LOWER(COALESCE(ih.reference_type, '')) IN ('sale', 'purchase', 'return', 'inspection', 'repair', 'acquisition') THEN LOWER(ih.reference_type)
					ELSE 'inventory_item'
				END AS entity_type,
				COALESCE(CAST(ih.inventory_item_id AS TEXT), '') AS entity_id,
				COALESCE(ih.reference_type, '') AS reference_type,
				COALESCE(CAST(ih.reference_id AS TEXT), '') AS reference_id,
				COALESCE(CAST(ih.created_by AS TEXT), '') AS user_id,
				COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), u.email, CAST(ih.created_by AS TEXT), '') AS user_name,
				'completed' AS status,
				COALESCE(ih.description, ih.event_type, 'تاريخ مخزون') AS description,
				COALESCE(ih.metadata, '') AS details,
				NULL AS quantity, NULL AS before_value, NULL AS after_value, ih.created_at
			FROM item_history ih
			LEFT JOIN users u ON CAST(u.id AS TEXT) = CAST(ih.created_by AS TEXT)

			UNION ALL

			SELECT
				'ledger:' || CAST(le.id AS TEXT) AS event_id,
				COALESCE(le.transaction_type, 'ledger_entry') AS event_type,
				CASE
					WHEN LOWER(COALESCE(le.ledger_type, '')) = 'customer' THEN 'customers'
					WHEN LOWER(COALESCE(le.ledger_type, '')) = 'supplier' THEN 'suppliers'
					ELSE 'inventory'
				END AS section,
				CASE
					WHEN LOWER(COALESCE(le.ledger_type, '')) = 'customer' THEN 'customer'
					WHEN LOWER(COALESCE(le.ledger_type, '')) = 'supplier' THEN 'supplier'
					ELSE 'product'
				END AS entity_type,
				COALESCE(CAST(le.entity_id AS TEXT), '') AS entity_id,
				COALESCE(le.reference_type, '') AS reference_type,
				COALESCE(CAST(le.reference_id AS TEXT), '') AS reference_id,
				COALESCE(CAST(le.created_by AS TEXT), '') AS user_id,
				COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), u.email, CAST(le.created_by AS TEXT), '') AS user_name,
				'completed' AS status,
				COALESCE(le.description, le.transaction_type, 'معاملة مالية') AS description,
				COALESCE(le.metadata, '') AS details,
				CAST(le.amount AS REAL) AS quantity,
				CAST(le.previous_balance AS REAL) AS before_value,
				CAST(le.balance AS REAL) AS after_value,
				le.created_at
			FROM ledger_entries le
			LEFT JOIN users u ON CAST(u.id AS TEXT) = CAST(le.created_by AS TEXT)

			%s
		) events
		WHERE 1 = 1`, auditDetailsSQL, r.legacyLedgerSources(ctx))

	args := make([]any, 0, 16)
	add := func(condition string, value any) { base += " AND " + condition; args = append(args, value) }
	if req.Search != "" {
		pattern := "%" + req.Search + "%"
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		searchFields := []string{"event_id", "event_type", "section", "entity_type", "entity_id", "reference_id", "user_name", "description", "details"}
		conditions := make([]string, 0, len(searchFields))
		for _, field := range searchFields {
			conditions = append(conditions, field+" "+likeOperator+" "+placeholder)
		}
		add("("+strings.Join(conditions, " OR ")+")", pattern)
	}
	if req.Section != "" {
		add(fmt.Sprintf("section = $%d", len(args)+1), req.Section)
	}
	if req.EventType != "" {
		add(fmt.Sprintf("LOWER(event_type) = LOWER($%d)", len(args)+1), req.EventType)
	}
	if req.UserID != "" {
		add(fmt.Sprintf("user_id = $%d", len(args)+1), req.UserID)
	}
	if req.Status != "" {
		add(fmt.Sprintf("LOWER(status) = LOWER($%d)", len(args)+1), req.Status)
	}
	if req.EntityType != "" {
		add(fmt.Sprintf("LOWER(entity_type) = LOWER($%d)", len(args)+1), req.EntityType)
	}
	if req.EntityID != "" {
		add(fmt.Sprintf("entity_id = $%d", len(args)+1), req.EntityID)
	}
	if req.Reference != "" {
		add(fmt.Sprintf("(reference_id = $%d OR details %s $%d)", len(args)+1, likeOperator, len(args)+1), req.Reference)
	}
	if req.StartDate != "" {
		add(fmt.Sprintf("created_at >= $%d", len(args)+1), req.StartDate)
	}
	if req.EndDate != "" {
		add(fmt.Sprintf("created_at <= $%d", len(args)+1), req.EndDate)
	}

	orderBy := "created_at"
	if req.SortBy == "event_type" {
		orderBy = "event_type"
	}
	if req.SortBy == "section" {
		orderBy = "section"
	}
	if req.SortBy == "status" {
		orderBy = "status"
	}
	order := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		order = "ASC"
	}

	countSQL := "SELECT COUNT(*) FROM (" + base + ") counted"
	var total int
	if err := r.db.GetContext(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, fmt.Errorf("count archive events: %w", err)
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	query := fmt.Sprintf("SELECT event_id AS id, event_type, section, entity_type, entity_id, reference_type, reference_id, user_id, user_name, status, description, details, quantity, before_value, after_value, created_at FROM (%s) events WHERE 1 = 1 ORDER BY %s %s LIMIT $%d OFFSET $%d", base, orderBy, order, limitIndex, offsetIndex)
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
	var events []Event
	if err := r.db.SelectContext(ctx, &events, query, args...); err != nil {
		return nil, 0, fmt.Errorf("list archive events: %w", err)
	}
	return events, total, nil
}
