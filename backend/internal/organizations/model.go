package organizations

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a multi-tenant organization
type Organization struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	Slug               string     `json:"slug" db:"slug"`
	Email              *string    `json:"email,omitempty" db:"email"`
	Phone              *string    `json:"phone,omitempty" db:"phone"`
	Address            *string    `json:"address,omitempty" db:"address"`
	City               *string    `json:"city,omitempty" db:"city"`
	Country            *string    `json:"country,omitempty" db:"country"`
	LogoURL            *string    `json:"logo_url,omitempty" db:"logo_url"`
	Settings           map[string]interface{} `json:"settings,omitempty" db:"settings"`
	SubscriptionPlan   string     `json:"subscription_plan" db:"subscription_plan"`
	SubscriptionStatus string     `json:"subscription_status" db:"subscription_status"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the Organization model
func (Organization) TableName() string {
	return "organizations"
}

// NewOrganization creates a new Organization instance
func NewOrganization(name, slug string) *Organization {
	return &Organization{
		ID:                 uuid.New(),
		Name:               name,
		Slug:               slug,
		SubscriptionPlan:   "free",
		SubscriptionStatus: "active",
		Settings:           make(map[string]interface{}),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}
