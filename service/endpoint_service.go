package service

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// EndpointService defines the business logic interface for endpoints
type EndpointService interface {
	// CreateEndpoint creates a new endpoint
	CreateEndpoint(endpoint *model.ProviderEndpoint) error

	// GetEndpointByID retrieves an endpoint by ID
	GetEndpointByID(id string) (*model.ProviderEndpoint, error)

	// GetEndpointByProviderAndName retrieves an endpoint by provider name and endpoint name (action)
	GetEndpointByProviderAndName(providerName string, name string) (*model.ProviderEndpoint, error)

	// GetEndpointsByProviderID retrieves all endpoints for a provider
	GetEndpointsByProviderID(providerID string) ([]*model.ProviderEndpoint, error)

	// UpdateEndpoint updates an existing endpoint
	UpdateEndpoint(endpoint *model.ProviderEndpoint) error

	// DeleteEndpoint soft deletes an endpoint
	DeleteEndpoint(id string) error
}
