package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// ResponseMappingRepo defines the interface for response mapping data operations
type ResponseMappingRepo interface {
	// Create creates a new response mapping
	Create(mapping *models.ResponseMapping) error

	// FindByID retrieves a response mapping by ID
	FindByID(id string) (*models.ResponseMapping, error)

	// FindByProviderIDAndAction retrieves all response mappings for a provider and action
	FindByProviderIDAndAction(providerID string, action string) ([]*models.ResponseMapping, error)

	// Update updates an existing response mapping
	Update(mapping *models.ResponseMapping) error

	// Delete soft deletes a response mapping
	Delete(id string) error
}
