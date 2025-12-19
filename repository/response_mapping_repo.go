package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// ResponseMappingRepo defines the interface for response mapping data operations
type ResponseMappingRepo interface {
	// Create creates a new response mapping
	Create(mapping *model.ProviderResponseMapping) error

	// FindByID retrieves a response mapping by ID
	FindByID(id string) (*model.ProviderResponseMapping, error)

	// FindByProviderIDAndAction retrieves all response mappings for a provider and action
	FindByProviderIDAndAction(providerID string, action string) ([]*model.ProviderResponseMapping, error)

	// Update updates an existing response mapping
	Update(mapping *model.ProviderResponseMapping) error

	// Delete soft deletes a response mapping
	Delete(id string) error
}
