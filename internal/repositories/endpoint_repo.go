package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// EndpointRepo defines the interface for endpoint data operations
type EndpointRepo interface {
	// Create creates a new endpoint
	Create(endpoint *models.Endpoint) error

	// FindByID retrieves an endpoint by ID
	FindByID(id string) (*models.Endpoint, error)

	// FindByProviderIDAndName retrieves an endpoint by provider ID and name (action)
	FindByProviderIDAndName(providerID string, name string) (*models.Endpoint, error)

	// FindByProviderID retrieves all endpoints for a provider
	FindByProviderID(providerID string) ([]*models.Endpoint, error)

	// Update updates an existing endpoint
	Update(endpoint *models.Endpoint) error

	// Delete soft deletes an endpoint
	Delete(id string) error
}
