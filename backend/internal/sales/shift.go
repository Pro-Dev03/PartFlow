package sales

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/response"
)

type PosShift struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Status      string     `json:"status" db:"status"`
	OpenedAt    time.Time  `json:"opened_at" db:"opened_at"`
	OpeningCash float64    `json:"opening_cash" db:"opening_cash"`
	ClosedAt    *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	ClosingCash *float64   `json:"closing_cash,omitempty" db:"closing_cash"`
	SalesTotal  float64    `json:"sales_total" db:"sales_total"`
	SaleCount   int        `json:"sale_count" db:"sale_count"`
}

type shiftAmountRequest struct {
	Amount float64 `json:"amount"`
}

func ensurePosShiftTable(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS pos_shifts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'open',
			opened_at TEXT NOT NULL,
			opening_cash REAL NOT NULL DEFAULT 0,
			closed_at TEXT,
			closing_cash REAL,
			sales_total REAL NOT NULL DEFAULT 0,
			sale_count INTEGER NOT NULL DEFAULT 0
		)`)
	if err != nil {
		return err
	}

	// Older deployments may already have pos_shifts but lack fields added by
	// later versions. CREATE TABLE IF NOT EXISTS does not update that schema,
	// which made even GET /shifts/current fail with an undefined-column error.
	columns := []struct {
		name    string
		typeSQL string
	}{
		{name: "opening_cash", typeSQL: "NUMERIC(15,2) NOT NULL DEFAULT 0"},
		{name: "closed_at", typeSQL: "TIMESTAMPTZ"},
		{name: "closing_cash", typeSQL: "NUMERIC(15,2)"},
		{name: "sales_total", typeSQL: "NUMERIC(15,2) NOT NULL DEFAULT 0"},
		{name: "sale_count", typeSQL: "BIGINT NOT NULL DEFAULT 0"},
	}

	if dbutil.IsSQLite(db) {
		rows, err := db.Queryx(`PRAGMA table_info(pos_shifts)`)
		if err != nil {
			return fmt.Errorf("inspect POS shift columns: %w", err)
		}
		existing := make(map[string]bool, len(columns)+4)
		for rows.Next() {
			var cid, notNull, primaryKey int
			var name, columnType string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
				_ = rows.Close()
				return fmt.Errorf("read POS shift column: %w", err)
			}
			existing[name] = true
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("iterate POS shift columns: %w", err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close POS shift schema query: %w", err)
		}
		for _, column := range columns {
			if existing[column.name] {
				continue
			}
			if _, err := db.Exec(`ALTER TABLE pos_shifts ADD COLUMN ` + column.name + ` ` + column.typeSQL); err != nil {
				return fmt.Errorf("add POS shift column %s: %w", column.name, err)
			}
		}
		return nil
	}

	for _, column := range columns {
		if _, err := db.Exec(`ALTER TABLE pos_shifts ADD COLUMN IF NOT EXISTS ` + column.name + ` ` + column.typeSQL); err != nil {
			return fmt.Errorf("add POS shift column %s: %w", column.name, err)
		}
	}
	return nil
}

func currentShift(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (*PosShift, error) {
	return currentShiftFrom(ctx, db, userID)
}

func currentShiftFrom(ctx context.Context, exec sqlx.ExtContext, userID uuid.UUID) (*PosShift, error) {
	var row struct {
		ID          string  `db:"id"`
		UserID      string  `db:"user_id"`
		Status      string  `db:"status"`
		OpenedAt    any     `db:"opened_at"`
		OpeningCash float64 `db:"opening_cash"`
		ClosedAt    any     `db:"closed_at"`
		ClosingCash any     `db:"closing_cash"`
		SalesTotal  float64 `db:"sales_total"`
		SaleCount   int     `db:"sale_count"`
	}

	err := sqlx.GetContext(ctx, exec, &row, `
		SELECT id, user_id, status, opened_at, opening_cash, closed_at, closing_cash, sales_total, sale_count
		FROM pos_shifts WHERE CAST(user_id AS TEXT) = $1 AND status = 'open' ORDER BY opened_at DESC LIMIT 1`, userID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	openedAt, err := dbutil.ParseTimestamp(row.OpenedAt)
	if err != nil {
		return nil, err
	}

	shift := &PosShift{
		ID:          mustParseUUID(row.ID),
		UserID:      mustParseUUID(row.UserID),
		Status:      row.Status,
		OpenedAt:    openedAt,
		OpeningCash: row.OpeningCash,
		SalesTotal:  row.SalesTotal,
		SaleCount:   row.SaleCount,
	}
	if row.ClosedAt != nil {
		closedAt, err := dbutil.ParseTimestamp(row.ClosedAt)
		if err != nil {
			return nil, err
		}
		shift.ClosedAt = &closedAt
	}
	if row.ClosingCash != nil {
		switch value := row.ClosingCash.(type) {
		case float64:
			shift.ClosingCash = &value
		case float32:
			converted := float64(value)
			shift.ClosingCash = &converted
		case int:
			converted := float64(value)
			shift.ClosingCash = &converted
		case int64:
			converted := float64(value)
			shift.ClosingCash = &converted
		case []byte:
			parsed, err := strconv.ParseFloat(string(value), 64)
			if err != nil {
				return nil, err
			}
			shift.ClosingCash = &parsed
		}
	}
	return shift, nil
}

func mustParseUUID(value string) uuid.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func shiftUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}

func getCurrentShiftHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := shiftUserID(c)
		if !ok {
			errors.HandleError(c, errors.NewUnauthorizedError("User not authenticated", nil))
			return
		}
		shift, err := currentShift(c.Request.Context(), db, userID)
		if err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to retrieve current shift"))
			return
		}
		response.OK(c, shift, "Current shift retrieved successfully")
	}
}

func openShiftHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := shiftUserID(c)
		if !ok {
			errors.HandleError(c, errors.NewUnauthorizedError("User not authenticated", nil))
			return
		}
		var req shiftAmountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.HandleError(c, errors.ValidateRequest(err))
			return
		}
		if req.Amount < 0 {
			errors.HandleError(c, errors.NewValidationError("opening cash cannot be negative", nil))
			return
		}
		if existing, err := currentShift(c.Request.Context(), db, userID); err != nil {
			errors.HandleError(c, err)
			return
		} else if existing != nil {
			errors.HandleError(c, errors.NewBusinessError("an open shift already exists", nil))
			return
		}
		now := time.Now().UTC()
		shift := PosShift{ID: uuid.New(), UserID: userID, Status: "open", OpenedAt: now, OpeningCash: req.Amount}
		_, err := db.ExecContext(c.Request.Context(), `INSERT INTO pos_shifts (id, user_id, status, opened_at, opening_cash) VALUES ($1, $2, 'open', $3, $4)`, shift.ID.String(), userID.String(), now, req.Amount)
		if err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to open shift"))
			return
		}
		response.Created(c, shift, "Shift opened successfully")
	}
}

func closeShiftHandler(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := shiftUserID(c)
		if !ok {
			errors.HandleError(c, errors.NewUnauthorizedError("User not authenticated", nil))
			return
		}
		var req shiftAmountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errors.HandleError(c, errors.ValidateRequest(err))
			return
		}
		if req.Amount < 0 {
			errors.HandleError(c, errors.NewValidationError("closing cash cannot be negative", nil))
			return
		}
		tx, err := db.BeginTxx(c.Request.Context(), nil)
		if err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to begin shift close"))
			return
		}
		committed := false
		defer func() {
			if !committed {
				_ = tx.Rollback()
			}
		}()
		shift, err := currentShiftFrom(c.Request.Context(), tx, userID)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		if shift == nil {
			errors.HandleError(c, errors.NewBusinessError("no open shift exists", nil))
			return
		}
		now := time.Now().UTC()
		if err = refreshShiftSummary(c.Request.Context(), tx, shift.ID.String(), shift.UserID.String(), shift.OpenedAt, &now); err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to reconcile shift sales"))
			return
		}
		var reconciled struct {
			SalesTotal float64 `db:"sales_total"`
			SaleCount  int     `db:"sale_count"`
		}
		if err = tx.GetContext(c.Request.Context(), &reconciled, `SELECT sales_total, sale_count FROM pos_shifts WHERE id = $1`, shift.ID.String()); err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to read reconciled shift"))
			return
		}
		result, err := tx.ExecContext(c.Request.Context(), `UPDATE pos_shifts SET status = 'closed', closed_at = $1, closing_cash = $2 WHERE id = $3 AND status = 'open'`, now, req.Amount, shift.ID.String())
		if err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to close shift"))
			return
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			errors.HandleError(c, errors.NewBusinessError("shift is no longer open", nil))
			return
		}
		if err = tx.Commit(); err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to commit shift close"))
			return
		}
		committed = true
		shift.Status, shift.ClosedAt, shift.ClosingCash = "closed", &now, &req.Amount
		shift.SalesTotal, shift.SaleCount = reconciled.SalesTotal, reconciled.SaleCount
		response.OK(c, shift, "Shift closed successfully")
	}
}

func refreshShiftSummary(ctx context.Context, exec sqlx.ExtContext, shiftID, userID string, openedAt time.Time, closedAt *time.Time) error {
	query := `SELECT COALESCE(SUM(total_amount), 0) AS sales_total, COUNT(*) AS sale_count FROM sales WHERE CAST(user_id AS TEXT) = $1 AND LOWER(COALESCE(status, 'completed')) = 'completed' AND created_at >= $2`
	args := []interface{}{userID, openedAt.UTC().Format(time.RFC3339Nano)}
	if closedAt != nil {
		query += ` AND created_at < $3`
		args = append(args, closedAt.UTC().Format(time.RFC3339Nano))
	}
	var totals struct {
		SalesTotal float64 `db:"sales_total"`
		SaleCount  int     `db:"sale_count"`
	}
	if err := sqlx.GetContext(ctx, exec, &totals, query, args...); err != nil {
		return err
	}
	_, err := exec.ExecContext(ctx, `UPDATE pos_shifts SET sales_total = $1, sale_count = $2 WHERE id = $3`, totals.SalesTotal, totals.SaleCount, shiftID)
	return err
}

func registerShiftRoutes(router *gin.RouterGroup, db *sqlx.DB) error {
	if err := ensurePosShiftTable(db); err != nil {
		return fmt.Errorf("initialize POS shifts: %w", err)
	}
	shifts := router.Group("/sales/shifts")
	shifts.GET("/current", getCurrentShiftHandler(db))
	// Shift operations are store workflows for every authenticated subscriber.
	// The account-level subscription middleware protects this entire route group.
	shifts.POST("/open", openShiftHandler(db))
	shifts.POST("/close", closeShiftHandler(db))
	return nil
}
