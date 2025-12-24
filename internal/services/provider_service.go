package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// ProviderService defines the business logic interface for providers.
// Service can call only its repo (not others) - follows single responsibility principle.
type ProviderService interface {
	// CreateProvider creates a new provider
	CreateProvider(provider *models.Provider) error

	// GetProviderByID retrieves a provider by ID
	GetProviderByID(id string) (*models.Provider, error)

	// GetProviderByName retrieves a provider by name
	GetProviderByName(name string) (*models.Provider, error)

	// GetAllActive retrieves all active providers
	GetAllActive() ([]*models.Provider, error)

	// ListProviders retrieves all providers (including inactive)
	ListProviders() ([]*models.Provider, error)

	// UpdateProvider updates an existing provider
	UpdateProvider(provider *models.Provider) error

	// DeleteProvider deletes a provider by ID
	DeleteProvider(id string) error
}
