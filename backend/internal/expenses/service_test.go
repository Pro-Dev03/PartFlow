package expenses

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	_ "modernc.org/sqlite"
)

func newExpenseServiceTestDB(t *testing.T) (*Service, *sql.DB, uuid.UUID, uuid.UUID) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE expense_categories (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, color TEXT, icon TEXT,
			budget REAL NOT NULL DEFAULT 0, is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE expenses (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, category_id TEXT, amount REAL NOT NULL,
			currency TEXT, reference TEXT, notes TEXT, expense_date TEXT NOT NULL,
			status TEXT NOT NULL, is_recurring INTEGER NOT NULL DEFAULT 0, recurring_period TEXT,
			approved_by TEXT, created_by TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			description TEXT, payment_method TEXT, receipt_url TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	categoryID := uuid.New()
	expenseID := uuid.New()
	stamp := "2026-09-25T12:00:00Z"
	if _, err := db.Exec(`INSERT INTO expense_categories (id, name, budget, is_active, created_at, updated_at) VALUES (?, 'Rent', 500, 1, ?, ?)`, categoryID.String(), stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO expenses (id, title, category_id, amount, currency, expense_date, status, is_recurring, created_by, created_at, updated_at, description, payment_method) VALUES (?, 'Rent', ?, 25, 'ILS', ?, 'pending', 0, ?, ?, ?, 'Rent', 'cash')`, expenseID.String(), categoryID.String(), stamp, uuid.NewString(), stamp, stamp); err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite")
	return NewService(NewRepository(sqlxDB)), db, categoryID, expenseID
}

func TestValidateExpenseRequestAcceptsDecimalAmounts(t *testing.T) {
	request := &ExpenseRequest{
		CategoryID:    uuid.New(),
		Title:         "  Utilities  ",
		Amount:        12.75,
		Currency:      "ILS",
		ExpenseDate:   time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC),
		PaymentMethod: "cash",
	}
	if err := ValidateExpenseRequest(request); err != nil {
		t.Fatalf("valid decimal expense rejected: %v", err)
	}
	if request.Title != "Utilities" {
		t.Fatalf("title was not trimmed: %q", request.Title)
	}
}

func TestUpdateExpenseAcceptsDecimalAmountAndPreservesOmittedFields(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	amount := 49.95
	request := &ExpenseUpdateRequest{Amount: &amount}
	if _, err := service.UpdateExpense(context.Background(), expenseID, request); err != nil {
		t.Fatalf("update with decimal amount failed: %v", err)
	}
	var storedAmount float64
	var recurring int
	if err := db.QueryRow(`SELECT amount, is_recurring FROM expenses WHERE id = ?`, expenseID.String()).Scan(&storedAmount, &recurring); err != nil {
		t.Fatal(err)
	}
	if storedAmount != amount || recurring != 0 {
		t.Fatalf("stored amount/recurring = %v/%d, want %v/0", storedAmount, recurring, amount)
	}
}

func TestUpdateExpenseCannotChangeApprovalStatusOrSetZeroAmount(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	zero := 0.0
	if _, err := service.UpdateExpense(context.Background(), expenseID, &ExpenseUpdateRequest{Amount: &zero}); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("zero amount error = %v, want ErrInvalidAmount", err)
	}
	if _, err := service.UpdateExpense(context.Background(), expenseID, &ExpenseUpdateRequest{Status: "approved"}); !errors.Is(err, ErrInvalidExpenseStatus) {
		t.Fatalf("status update error = %v, want ErrInvalidExpenseStatus", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM expenses WHERE id = ?`, expenseID.String()).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("expense status = %q after rejected update, want pending", status)
	}
}

func TestAccountingExpensesCannotBeEdited(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	for _, status := range []string{"approved", "paid", "completed"} {
		if _, err := db.Exec(`UPDATE expenses SET status = ? WHERE id = ?`, status, expenseID.String()); err != nil {
			t.Fatal(err)
		}
		amount := 75.25
		if _, err := service.UpdateExpense(context.Background(), expenseID, &ExpenseUpdateRequest{Amount: &amount}); !errors.Is(err, ErrExpenseAlreadyApproved) {
			t.Errorf("update %q expense error = %v, want ErrExpenseAlreadyApproved", status, err)
		}
	}
	var amount float64
	if err := db.QueryRow(`SELECT amount FROM expenses WHERE id = ?`, expenseID.String()).Scan(&amount); err != nil {
		t.Fatal(err)
	}
	if amount != 25 {
		t.Fatalf("financially committed expense amount changed to %v, want 25", amount)
	}
}

func TestAccountingExpensesCanBeArchivedWithoutChangingFinancialTotals(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	dbHandle := sqlx.NewDb(db, "sqlite")
	start := time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	for _, status := range []string{"approved", "paid", "completed"} {
		if _, err := db.Exec(`UPDATE expenses SET status = ? WHERE id = ?`, status, expenseID.String()); err != nil {
			t.Fatal(err)
		}
		before, err := accounting.TotalExpensesForPeriod(context.Background(), dbHandle, start, end)
		if err != nil {
			t.Fatal(err)
		}
		if err := service.DeleteExpense(context.Background(), expenseID); err != nil {
			t.Fatalf("archive %q expense: %v", status, err)
		}
		var storedStatus string
		if err := db.QueryRow(`SELECT status FROM expenses WHERE id = ?`, expenseID.String()).Scan(&storedStatus); err != nil {
			t.Fatal(err)
		}
		if storedStatus != "archived" {
			t.Fatalf("expense status after cleanup = %q, want archived", storedStatus)
		}
		after, err := accounting.TotalExpensesForPeriod(context.Background(), dbHandle, start, end)
		if err != nil {
			t.Fatal(err)
		}
		if before != 25 || after != before {
			t.Fatalf("financial total before/after archive = %v/%v, want 25/25", before, after)
		}
		if err := service.DeleteExpense(context.Background(), expenseID); !errors.Is(err, ErrExpenseAlreadyArchived) {
			t.Fatalf("repeat archive error = %v, want ErrExpenseAlreadyArchived", err)
		}
	}
}

func TestArchivedExpensesAreHiddenFromOperationalListButCanBeIncluded(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	if _, err := db.Exec(`UPDATE expenses SET status = 'archived' WHERE id = ?`, expenseID.String()); err != nil {
		t.Fatal(err)
	}
	_, count, err := service.repo.ListExpenses(context.Background(), ExpenseListRequest{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("operational list count = %d, want 0", count)
	}
	rows, count, err := service.repo.ListExpenses(context.Background(), ExpenseListRequest{Page: 1, PerPage: 20, IncludeArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(rows) != 1 || rows[0].Status != "archived" {
		t.Fatalf("archived inclusive list = %d rows/%d total, want 1/1", len(rows), count)
	}
}

func TestUnknownExpenseStatusesCannotBeDeleted(t *testing.T) {
	service, db, _, expenseID := newExpenseServiceTestDB(t)
	if _, err := db.Exec(`UPDATE expenses SET status = '' WHERE id = ?`, expenseID.String()); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteExpense(context.Background(), expenseID); !errors.Is(err, ErrInvalidExpenseStatus) {
		t.Fatalf("delete expense with unknown status error = %v, want ErrInvalidExpenseStatus", err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE id = ?`, expenseID.String()).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("unknown-status expense rows remaining = %d, want 1", remaining)
	}
}

func TestExpenseCategoryDefaultsActiveAndPartialUpdatePreservesBudgetAndStatus(t *testing.T) {
	service, _, categoryID, _ := newExpenseServiceTestDB(t)
	created, err := service.CreateExpenseCategory(context.Background(), &ExpenseCategoryRequest{Name: "Office"})
	if err != nil {
		t.Fatalf("create category without is_active: %v", err)
	}
	if !created.IsActive {
		t.Fatal("category created without is_active should default active")
	}
	updated, err := service.UpdateExpenseCategory(context.Background(), categoryID, &ExpenseCategoryUpdateRequest{Name: "Shop Rent"})
	if err != nil {
		t.Fatalf("partial category update: %v", err)
	}
	if !updated.IsActive || updated.Budget != 500 {
		t.Fatalf("partial update changed category active/budget: %v/%v, want true/500", updated.IsActive, updated.Budget)
	}
}

func TestUpdateExpenseCategoryRejectsBlankNameAndNonFiniteBudget(t *testing.T) {
	service, _, categoryID, _ := newExpenseServiceTestDB(t)
	if _, err := service.UpdateExpenseCategory(context.Background(), categoryID, &ExpenseCategoryUpdateRequest{Name: "   "}); !errors.Is(err, ErrInvalidExpenseCategoryName) {
		t.Fatalf("blank category name error = %v, want ErrInvalidExpenseCategoryName", err)
	}
	notANumber := math.NaN()
	if _, err := service.UpdateExpenseCategory(context.Background(), categoryID, &ExpenseCategoryUpdateRequest{Budget: &notANumber}); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("non-finite category budget error = %v, want ErrInvalidAmount", err)
	}
}

func TestExpenseSortClausesAllowlistSQLIdentifiersAndDirections(t *testing.T) {
	column, direction := expenseSortClause("amount", "asc")
	if column != "amount" || direction != "ASC" {
		t.Fatalf("valid expense sort = %s %s, want amount ASC", column, direction)
	}
	column, direction = expenseSortClause("expense_date; DROP TABLE expenses", "DESC; --")
	if column != "expense_date" || direction != "DESC" {
		t.Fatalf("untrusted expense sort = %s %s, want expense_date DESC", column, direction)
	}
	column, direction = expenseCategorySortClause("budget", "desc")
	if column != "budget" || direction != "DESC" {
		t.Fatalf("valid category sort = %s %s, want budget DESC", column, direction)
	}
	column, direction = expenseCategorySortClause("name;--", "ASC NULLS LAST")
	if column != "name" || direction != "ASC" {
		t.Fatalf("untrusted category sort = %s %s, want name ASC", column, direction)
	}
}
