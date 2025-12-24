package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// RequestSchemaRepo defines the interface for request schema data operations
type RequestSchemaRepo interface {
	// Create creates a new request schema
	Create(schema *models.RequestSchema) error

	// FindByID retrieves a request schema by ID
	FindByID(id string) (*models.RequestSchema, error)

	// FindByEndpointID retrieves all request schemas for an endpoint
	FindByEndpointID(endpointID string) ([]*models.RequestSchema, error)

	// FindByEndpointIDAndLocation retrieves schemas for an endpoint filtered by param location
	FindByEndpointIDAndLocation(endpointID string, location string) ([]*models.RequestSchema, error)

	// FindRequiredByEndpointID retrieves all required schemas for an endpoint
	FindRequiredByEndpointID(endpointID string) ([]*models.RequestSchema, error)

	// Update updates an existing request schema
	Update(schema *models.RequestSchema) error

	// Delete soft deletes a request schema
	Delete(id string) error
}
