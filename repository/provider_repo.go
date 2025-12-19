package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// ProviderRepo defines the interface for provider data operations
type ProviderRepo interface {
	// Create creates a new provider
	Create(provider *model.Provider) error

	// FindByID retrieves a provider by ID
	FindByID(id string) (*model.Provider, error)

	// FindByName retrieves a provider by name
	FindByName(name string) (*model.Provider, error)

	// List retrieves all active providers
	List() ([]*model.Provider, error)

	// Update updates an existing provider
	Update(provider *model.Provider) error

	// Delete soft deletes a provider
	Delete(id string) error
}
