package integrations

import (
	"context"
	"fmt"
)

// IntegrationType represents different types of external integrations
type IntegrationType string

const (
	IntegrationTypePayment     IntegrationType = "payment"
	IntegrationTypeMessaging   IntegrationType = "messaging"
	IntegrationTypeStorage     IntegrationType = "storage"
	IntegrationTypeShipping    IntegrationType = "shipping"
	IntegrationTypeAnalytics   IntegrationType = "analytics"
	IntegrationTypeNotification IntegrationType = "notification"
)

// IntegrationStatus represents the status of an integration
type IntegrationStatus string

const (
	StatusActive    IntegrationStatus = "active"
	StatusInactive  IntegrationStatus = "inactive"
	StatusError     IntegrationStatus = "error"
	StatusConfiguring IntegrationStatus = "configuring"
)

// IntegrationConfig holds configuration for an integration
type IntegrationConfig struct {
	Type        IntegrationType
	Name        string
	APIKey      string
	APISecret   string
	WebhookURL  string
	Environment string // "development", "production", etc.
	Settings    map[string]interface{}
}

// Integration represents an external service integration
type Integration interface {
	Initialize(config IntegrationConfig) error
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	TestConnection(ctx context.Context) error
	GetStatus() IntegrationStatus
	GetConfig() IntegrationConfig
}

// Manager manages all external integrations
// Inspired by Fynexa's integration management pattern
type Manager struct {
	integrations map[IntegrationType]Integration
}

// NewManager creates a new integration manager
func NewManager() *Manager {
	return &Manager{
		integrations: make(map[IntegrationType]Integration),
	}
}

// RegisterIntegration registers an integration for a specific type
func (m *Manager) RegisterIntegration(integrationType IntegrationType, integration Integration) error {
	if _, exists := m.integrations[integrationType]; exists {
		return fmt.Errorf("integration of type %s already registered", integrationType)
	}
	m.integrations[integrationType] = integration
	return nil
}

// GetIntegration retrieves an integration by type
func (m *Manager) GetIntegration(integrationType IntegrationType) (Integration, error) {
	integration, exists := m.integrations[integrationType]
	if !exists {
		return nil, fmt.Errorf("integration of type %s not found", integrationType)
	}
	return integration, nil
}

// InitializeAll initializes all registered integrations
func (m *Manager) InitializeAll(configs map[IntegrationType]IntegrationConfig) error {
	for integrationType, config := range configs {
		integration, err := m.GetIntegration(integrationType)
		if err != nil {
			return fmt.Errorf("failed to get integration %s: %w", integrationType, err)
		}
		if err := integration.Initialize(config); err != nil {
			return fmt.Errorf("failed to initialize integration %s: %w", integrationType, err)
		}
	}
	return nil
}

// ConnectAll connects all registered integrations
func (m *Manager) ConnectAll(ctx context.Context) error {
	for integrationType, integration := range m.integrations {
		if err := integration.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect integration %s: %w", integrationType, err)
		}
	}
	return nil
}

// DisconnectAll disconnects all registered integrations
func (m *Manager) DisconnectAll(ctx context.Context) error {
	for integrationType, integration := range m.integrations {
		if err := integration.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect integration %s: %w", integrationType, err)
		}
	}
	return nil
}

// GetStatuses returns the status of all integrations
func (m *Manager) GetStatuses() map[IntegrationType]IntegrationStatus {
	statuses := make(map[IntegrationType]IntegrationStatus)
	for integrationType, integration := range m.integrations {
		statuses[integrationType] = integration.GetStatus()
	}
	return statuses
}

// TestAllConnections tests all integration connections
func (m *Manager) TestAllConnections(ctx context.Context) map[IntegrationType]error {
	results := make(map[IntegrationType]error)
	for integrationType, integration := range m.integrations {
		err := integration.TestConnection(ctx)
		results[integrationType] = err
	}
	return results
}
