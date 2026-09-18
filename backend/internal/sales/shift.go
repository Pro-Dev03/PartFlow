package sales

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	return err
}

func currentShift(ctx context.Context, db *sqlx.DB, userID uuid.UUID) (*PosShift, error) {
	var shift PosShift
	err := db.GetContext(ctx, &shift, `
		SELECT id, user_id, status, opened_at, opening_cash, closed_at, closing_cash, sales_total, sale_count
		FROM pos_shifts WHERE user_id = $1 AND status = 'open' ORDER BY opened_at DESC LIMIT 1`, userID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &shift, nil
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
		shift, err := currentShift(c.Request.Context(), db, userID)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		if shift == nil {
			errors.HandleError(c, errors.NewBusinessError("no open shift exists", nil))
			return
		}
		now := time.Now().UTC()
		_, err = db.ExecContext(c.Request.Context(), `UPDATE pos_shifts SET status = 'closed', closed_at = $1, closing_cash = $2 WHERE id = $3`, now, req.Amount, shift.ID.String())
		if err != nil {
			errors.HandleError(c, errors.WrapError(err, "Failed to close shift"))
			return
		}
		shift.Status, shift.ClosedAt, shift.ClosingCash = "closed", &now, &req.Amount
		response.OK(c, shift, "Shift closed successfully")
	}
}

func registerShiftRoutes(router *gin.RouterGroup, db *sqlx.DB) error {
	if err := ensurePosShiftTable(db); err != nil {
		return fmt.Errorf("initialize POS shifts: %w", err)
	}
	shifts := router.Group("/sales/shifts")
	shifts.GET("/current", getCurrentShiftHandler(db))
	shifts.POST("/open", openShiftHandler(db))
	shifts.POST("/close", closeShiftHandler(db))
	return nil
}
