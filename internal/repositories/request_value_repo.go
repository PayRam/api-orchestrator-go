package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// RequestValueRepo defines the interface for request value data operations
type RequestValueRepo interface {
	// Create creates a new request value
	Create(value *models.RequestValue) error

	// FindByID retrieves a request value by ID
	FindByID(id string) (*models.RequestValue, error)

	// FindBySchemaID retrieves all request values for a schema
	FindBySchemaID(schemaID string) ([]*models.RequestValue, error)

	// Update updates an existing request value
	Update(value *models.RequestValue) error

	// Delete soft deletes a request value
	Delete(id string) error
}
