package service

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// ProviderService defines the business logic interface for providers
type ProviderService interface {
	// CreateProvider creates a new provider
	CreateProvider(provider *model.Provider) error

	// GetProviderByID retrieves a provider by ID
	GetProviderByID(id string) (*model.Provider, error)

	// GetProviderByName retrieves a provider by name
	GetProviderByName(name string) (*model.Provider, error)

	// ListProviders retrieves all active providers
	ListProviders() ([]*model.Provider, error)

	// UpdateProvider updates an existing provider
	UpdateProvider(provider *model.Provider) error

	// DeleteProvider soft deletes a provider
	DeleteProvider(id string) error
}
