package notifications

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSQLiteNotificationRoundTrip(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "notifications.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	repo := NewRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	if _, err := database.DB.Exec(`INSERT INTO users (id,email,password_hash,first_name,last_name,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, userID.String(), "notify@example.com", "hash", "Notify", "User", now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	n := &Notification{UserID: userID, Type: "low_stock", Title: "Low stock", Message: "One item", Data: `{"id":"x"}`, Priority: "high", Status: "unread"}
	if err := repo.CreateNotification(ctx, n); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetNotificationByID(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.UserID != userID || loaded.Data == "" {
		t.Fatalf("loaded notification mismatch: %#v", loaded)
	}
	list, count, err := repo.ListNotifications(ctx, userID, NotificationListRequest{Page: 1, PerPage: 20})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(list) != 1 {
		t.Fatalf("list count/items = %d/%d", count, len(list))
	}
	if err := repo.MarkAsRead(ctx, n.ID); err != nil {
		t.Fatal(err)
	}
	summary, err := repo.GetNotificationSummary(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 1 || summary.Unread != 0 || summary.Priority != 1 {
		t.Fatalf("summary mismatch: %#v", summary)
	}
}
