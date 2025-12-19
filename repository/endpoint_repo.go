package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// EndpointRepo defines the interface for endpoint data operations
type EndpointRepo interface {
	// Create creates a new endpoint
	Create(endpoint *model.ProviderEndpoint) error

	// FindByID retrieves an endpoint by ID
	FindByID(id string) (*model.ProviderEndpoint, error)

	// FindByProviderIDAndName retrieves an endpoint by provider ID and name (action)
	FindByProviderIDAndName(providerID string, name string) (*model.ProviderEndpoint, error)

	// FindByProviderID retrieves all endpoints for a provider
	FindByProviderID(providerID string) ([]*model.ProviderEndpoint, error)

	// Update updates an existing endpoint
	Update(endpoint *model.ProviderEndpoint) error

	// Delete soft deletes an endpoint
	Delete(id string) error
}
