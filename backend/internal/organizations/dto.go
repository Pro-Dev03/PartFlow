package organizations

import "github.com/google/uuid"

// CreateRequest represents the request to create an organization
type CreateRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Slug     string                 `json:"slug" binding:"required"`
	Email    string                 `json:"email,omitempty"`
	Phone    string                 `json:"phone,omitempty"`
	Address  string                 `json:"address,omitempty"`
	City     string                 `json:"city,omitempty"`
	Country  string                 `json:"country,omitempty"`
	LogoURL  string                 `json:"logo_url,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// UpdateRequest represents the request to update an organization
type UpdateRequest struct {
	Name               *string                `json:"name,omitempty"`
	Slug               *string                `json:"slug,omitempty"`
	Email              *string                `json:"email,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	Address            *string                `json:"address,omitempty"`
	City               *string                `json:"city,omitempty"`
	Country            *string                `json:"country,omitempty"`
	LogoURL            *string                `json:"logo_url,omitempty"`
	Settings           map[string]interface{} `json:"settings,omitempty"`
	SubscriptionPlan   *string                `json:"subscription_plan,omitempty"`
	SubscriptionStatus *string                `json:"subscription_status,omitempty"`
}

// Response represents the organization response
type Response struct {
	ID                 uuid.UUID              `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Email              *string                `json:"email,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	Address            *string                `json:"address,omitempty"`
	City               *string                `json:"city,omitempty"`
	Country            *string                `json:"country,omitempty"`
	LogoURL            *string                `json:"logo_url,omitempty"`
	Settings           map[string]interface{} `json:"settings,omitempty"`
	SubscriptionPlan   string                 `json:"subscription_plan"`
	SubscriptionStatus string                 `json:"subscription_status"`
	TrialEndsAt        *string                `json:"trial_ends_at,omitempty"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

// ToResponse converts an Organization to Response
func (o *Organization) ToResponse() *Response {
	var trialEndsAt *string
	if o.TrialEndsAt != nil {
		t := o.TrialEndsAt.Format("2006-01-02T15:04:05Z")
		trialEndsAt = &t
	}

	return &Response{
		ID:                 o.ID,
		Name:               o.Name,
		Slug:               o.Slug,
		Email:              o.Email,
		Phone:              o.Phone,
		Address:            o.Address,
		City:               o.City,
		Country:            o.Country,
		LogoURL:            o.LogoURL,
		Settings:           o.Settings,
		SubscriptionPlan:   o.SubscriptionPlan,
		SubscriptionStatus: o.SubscriptionStatus,
		TrialEndsAt:        trialEndsAt,
		CreatedAt:          o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
