package inspections

import (
	"time"

	"github.com/google/uuid"
)

// Inspection represents an inspection of a used item
type Inspection struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	SerialNumber   string     `json:"serial_number" db:"serial_number"`
	InspectionDate time.Time  `json:"inspection_date" db:"inspection_date"`
	InspectedBy    uuid.UUID  `json:"inspected_by" db:"inspected_by"`
	Status         string     `json:"status" db:"status"` // pending, passed, failed, needs_repair
	Condition      string     `json:"condition" db:"condition"` // excellent, very_good, good, fair, poor
	Grade          string     `json:"grade" db:"grade"` // A, B, C, D, F
	Notes          string     `json:"notes" db:"notes"`
	Photos         []string   `json:"photos" db:"photos"`
	TestResults    TestResults `json:"test_results" db:"test_results"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TestResults represents test results for an inspection
type TestResults struct {
	PowerTest       bool   `json:"power_test" db:"power_test"`
	TemperatureTest bool   `json:"temperature_test" db:"temperature_test"`
	PerformanceTest bool   `json:"performance_test" db:"performance_test"`
	PortsTest       bool   `json:"ports_test" db:"ports_test"`
	StorageTest     bool   `json:"storage_test" db:"storage_test"`
	VisualTest      bool   `json:"visual_test" db:"visual_test"`
	SerialTest      bool   `json:"serial_test" db:"serial_test"`
	Notes           string `json:"notes" db:"notes"`
}

// InspectionRequest represents inspection creation request
type InspectionRequest struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	SerialNumber   string     `json:"serial_number"`
	InspectionDate time.Time  `json:"inspection_date" binding:"required"`
	Condition      string     `json:"condition" binding:"required,oneof=excellent very_good good fair poor"`
	Grade          string     `json:"grade" binding:"required,oneof=A B C D F"`
	Notes          string     `json:"notes"`
	Photos         []string   `json:"photos"`
	TestResults    TestResults `json:"test_results"`
}

// InspectionUpdateRequest represents inspection update request
type InspectionUpdateRequest struct {
	InspectionDate time.Time  `json:"inspection_date"`
	Status         string     `json:"status" binding:"omitempty,oneof=pending passed failed needs_repair"`
	Condition      string     `json:"condition" binding:"omitempty,oneof=excellent very_good good fair poor"`
	Grade          string     `json:"grade" binding:"omitempty,oneof=A B C D F"`
	Notes          string     `json:"notes"`
	Photos         []string   `json:"photos"`
	TestResults    TestResults `json:"test_results"`
}

// InspectionResponse represents inspection response with related data
type InspectionResponse struct {
	Inspection Inspection   `json:"inspection"`
	Product   *ProductInfo `json:"product,omitempty"`
	Inspector *UserInfo    `json:"inspector,omitempty"`
}

// ProductInfo represents product information
type ProductInfo struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Model  string    `json:"model"`
	SKU    string    `json:"sku"`
	Barcode string   `json:"barcode"`
}

// UserInfo represents user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName string    `json:"last_name"`
	Email     string    `json:"email"`
}

// InspectionListRequest represents inspection list query parameters
type InspectionListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	ProductID    *uuid.UUID `form:"product_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending passed failed needs_repair"`
	Condition    string     `form:"condition" binding:"omitempty,oneof=excellent very_good good fair poor"`
	Grade        string     `form:"grade" binding:"omitempty,oneof=A B C D F"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	InspectedBy  *uuid.UUID `form:"inspected_by"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// InspectionSummary represents inspection summary statistics
type InspectionSummary struct {
	TotalInspections    int              `json:"total_inspections"`
	PassedInspections   int              `json:"passed_inspections"`
	FailedInspections   int              `json:"failed_inspections"`
	PendingInspections  int              `json:"pending_inspections"`
	ByCondition         map[string]int    `json:"by_condition"`
	ByGrade             map[string]int    `json:"by_grade"`
	ThisWeek            int              `json:"this_week"`
	ThisMonth           int              `json:"this_month"`
}
