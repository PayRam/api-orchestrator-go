package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// EndpointService defines the business logic interface for endpoints
type EndpointService interface {
	// CreateEndpoint creates a new endpoint
	CreateEndpoint(endpoint *models.Endpoint) error

	// GetEndpointByID retrieves an endpoint by ID
	GetEndpointByID(id string) (*models.Endpoint, error)

	// GetEndpointByProviderAndName retrieves an endpoint by provider name and endpoint name (action)
	GetEndpointByProviderAndName(providerName string, name string) (*models.Endpoint, error)

	// GetEndpointsByProviderID retrieves all endpoints for a provider
	GetEndpointsByProviderID(providerID string) ([]*models.Endpoint, error)

	// UpdateEndpoint updates an existing endpoint
	UpdateEndpoint(endpoint *models.Endpoint) error

	// DeleteEndpoint soft deletes an endpoint
	DeleteEndpoint(id string) error
}
