package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSQLiteResetRejectsCloudTarget(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/settings/database?target=online&confirmation_token=DELETE%20ALL%20DATA", nil)
	NewDatabaseHandler(db).DeleteAllData(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("SQLite reset with online target status=%d body=%s; want 400", recorder.Code, recorder.Body.String())
	}
}
