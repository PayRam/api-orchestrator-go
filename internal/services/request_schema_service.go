package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// RequestSchemaService defines the business logic interface for request schemas
type RequestSchemaService interface {
	// CreateRequestSchema creates a new request schema
	CreateRequestSchema(schema *models.RequestSchema) error

	// GetRequestSchemaByID retrieves a request schema by ID
	GetRequestSchemaByID(id string) (*models.RequestSchema, error)

	// GetRequestSchemasByEndpointID retrieves all request schemas for an endpoint
	GetRequestSchemasByEndpointID(endpointID string) ([]*models.RequestSchema, error)

	// GetRequestSchemasByEndpointAndLocation retrieves schemas for an endpoint filtered by location
	GetRequestSchemasByEndpointAndLocation(endpointID string, location string) ([]*models.RequestSchema, error)

	// GetRequiredSchemasByEndpointID retrieves all required schemas for an endpoint
	GetRequiredSchemasByEndpointID(endpointID string) ([]*models.RequestSchema, error)

	// UpdateRequestSchema updates an existing request schema
	UpdateRequestSchema(schema *models.RequestSchema) error

	// DeleteRequestSchema soft deletes a request schema
	DeleteRequestSchema(id string) error
}
