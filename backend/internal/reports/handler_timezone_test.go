package reports

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/accounting"
)

func TestParseReportDateRangeUsesInclusiveStoreCalendarDates(t *testing.T) {
	original := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(original) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/?start_date=2026-09-16&end_date=2026-09-16", nil)

	start, end, err := parseReportDateRange(context)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := start.Format(time.RFC3339), "2026-09-16T00:00:00-04:00"; got != want {
		t.Errorf("start boundary = %s, want %s", got, want)
	}
	if got, want := end.Format(time.RFC3339), "2026-09-17T00:00:00-04:00"; got != want {
		t.Errorf("exclusive end boundary = %s, want %s", got, want)
	}
	if duration := end.Sub(start); duration != 24*time.Hour {
		t.Errorf("ordinary store day duration = %s, want 24h", duration)
	}
}

func TestParseReportDateRangeUsesStoreDatesAcrossDST(t *testing.T) {
	original := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(original) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/?start_date=2026-03-08&end_date=2026-03-08", nil)

	start, end, err := parseReportDateRange(context)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := start.UTC().Format(time.RFC3339), "2026-03-08T05:00:00Z"; got != want {
		t.Errorf("DST day start = %s, want %s", got, want)
	}
	if got, want := end.UTC().Format(time.RFC3339), "2026-03-09T04:00:00Z"; got != want {
		t.Errorf("DST day end = %s, want %s", got, want)
	}
	if duration := end.Sub(start); duration != 23*time.Hour {
		t.Errorf("DST transition day duration = %s, want 23h", duration)
	}
}

func TestParseReportDateRangeMapsRFC3339InstantsToStoreDate(t *testing.T) {
	original := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(original) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/?start_date=2026-09-17T02%3A00%3A00Z&end_date=2026-09-17T02%3A00%3A00Z", nil)

	start, end, err := parseReportDateRange(context)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := start.Format(time.RFC3339), "2026-09-16T00:00:00-04:00"; got != want {
		t.Errorf("RFC3339 instant's store-day start = %s, want %s", got, want)
	}
	if got, want := end.Format(time.RFC3339), "2026-09-17T00:00:00-04:00"; got != want {
		t.Errorf("RFC3339 instant's exclusive store-day end = %s, want %s", got, want)
	}
}

func TestParseReportListCalendarEndIncludesWholeStoreDay(t *testing.T) {
	original := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(original) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	end, err := parseReportListBound("2026-09-16", true)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := end.Format(time.RFC3339), "2026-09-17T00:00:00-04:00"; got != want {
		t.Fatalf("report list exclusive end = %s, want %s", got, want)
	}

	instantEnd, err := parseReportListBound("2026-09-16T15:00:00Z", true)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := instantEnd.UTC().Format(time.RFC3339Nano), "2026-09-16T15:00:00.000000001Z"; got != want {
		t.Fatalf("RFC3339 inclusive end conversion = %s, want %s", got, want)
	}
}

func TestReportDateScannerPreservesStoreCalendarDate(t *testing.T) {
	original := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(original) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	var date reportDate
	if err := date.Scan("2026-09-01"); err != nil {
		t.Fatalf("scan SQL date: %v", err)
	}
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	if got := date.In(location).Format("2006-01-02 15:04:05 -07:00"); got != "2026-09-01 00:00:00 -04:00" {
		t.Fatalf("SQL DATE was shifted from its store calendar date: got %s", got)
	}
}
