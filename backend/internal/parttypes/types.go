package parttypes

import (
	"github.com/google/uuid"
	"time"
)

// PartType represents a type of used part (GPU, RAM, etc.)
type PartType struct {
	ID        uuid.UUID `json:"id" db:"id"`
	NameAr    string    `json:"name_ar" db:"name_ar"`
	NameEn    string    `json:"name_en" db:"name_en"`
	Icon      string    `json:"icon" db:"icon"`
	Color     string    `json:"color" db:"color"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PartSpecification represents a specification field (capacity, brand, etc.)
type PartSpecification struct {
	ID         uuid.UUID `json:"id" db:"id"`
	NameAr     string    `json:"name_ar" db:"name_ar"`
	NameEn     string    `json:"name_en" db:"name_en"`
	DataType   string    `json:"data_type" db:"data_type"` // text, number, select, boolean
	Options    []string  `json:"options" db:"options"`
	IsRequired bool      `json:"is_required" db:"is_required"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// TypeSpecification links specifications to part types
type TypeSpecification struct {
	ID              uuid.UUID `json:"id" db:"id"`
	PartTypeID      uuid.UUID `json:"part_type_id" db:"part_type_id"`
	SpecificationID uuid.UUID `json:"specification_id" db:"specification_id"`
	SortOrder       int       `json:"sort_order" db:"sort_order"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// ItemSpecificationValue stores actual spec values for inventory items
type ItemSpecificationValue struct {
	ID              uuid.UUID `json:"id" db:"id"`
	InventoryItemID uuid.UUID `json:"inventory_item_id" db:"inventory_item_id"`
	SpecificationID uuid.UUID `json:"specification_id" db:"specification_id"`
	ValueText       *string   `json:"value_text" db:"value_text"`
	ValueNumber     *float64  `json:"value_number" db:"value_number"`
	ValueBoolean    *bool     `json:"value_boolean" db:"value_boolean"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// PartTypeWithSpecs represents a part type with its specifications
type PartTypeWithSpecs struct {
	PartType
	Specifications []PartSpecification `json:"specifications"`
}

// CreatePartTypeRequest for creating new part types
type CreatePartTypeRequest struct {
	NameAr    string `json:"name_ar" binding:"required"`
	NameEn    string `json:"name_en" binding:"required"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// UpdatePartTypeRequest for updating part types
type UpdatePartTypeRequest struct {
	NameAr    *string `json:"name_ar"`
	NameEn    *string `json:"name_en"`
	Icon      *string `json:"icon"`
	Color     *string `json:"color"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

// CreateSpecificationRequest for creating new specifications
type CreateSpecificationRequest struct {
	NameAr     string   `json:"name_ar" binding:"required"`
	NameEn     string   `json:"name_en" binding:"required"`
	DataType   string   `json:"data_type" binding:"required"`
	Options    []string `json:"options"`
	IsRequired bool     `json:"is_required"`
}

// LinkSpecificationRequest for linking specifications to types
type LinkSpecificationRequest struct {
	PartTypeID      uuid.UUID `json:"part_type_id" binding:"required"`
	SpecificationID uuid.UUID `json:"specification_id" binding:"required"`
	SortOrder       int       `json:"sort_order"`
}

// UpdateItemSpecsRequest for updating item specification values
type UpdateItemSpecsRequest struct {
	Specifications []ItemSpecValue `json:"specifications"`
}

// ItemSpecValue represents a specification value for an item
type ItemSpecValue struct {
	SpecificationID uuid.UUID `json:"specification_id" binding:"required"`
	ValueText       *string   `json:"value_text"`
	ValueNumber     *float64  `json:"value_number"`
	ValueBoolean    *bool     `json:"value_boolean"`
}
