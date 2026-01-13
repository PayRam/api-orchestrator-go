package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// RequestValueService defines the business logic for request value operations
type RequestValueService interface {
	// CreateRequestValue creates a new request value
	CreateRequestValue(value *models.RequestValue) error

	// GetRequestValueByID retrieves a request value by ID
	GetRequestValueByID(id string) (*models.RequestValue, error)

	// GetRequestValuesBySchemaID retrieves all request values for a schema
	GetRequestValuesBySchemaID(schemaID string) ([]*models.RequestValue, error)

	// UpdateRequestValue updates an existing request value
	UpdateRequestValue(value *models.RequestValue) error

	// DeleteRequestValue deletes a request value by ID
	DeleteRequestValue(id string) error
}
