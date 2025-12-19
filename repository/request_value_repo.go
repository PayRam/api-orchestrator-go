package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// RequestValueRepo defines the interface for request value data operations
type RequestValueRepo interface {
	// Create creates a new request value
	Create(value *model.ProviderRequestValue) error

	// FindByID retrieves a request value by ID
	FindByID(id string) (*model.ProviderRequestValue, error)

	// FindBySchemaID retrieves all request values for a schema
	FindBySchemaID(schemaID string) ([]*model.ProviderRequestValue, error)

	// Update updates an existing request value
	Update(value *model.ProviderRequestValue) error

	// Delete soft deletes a request value
	Delete(id string) error
}
