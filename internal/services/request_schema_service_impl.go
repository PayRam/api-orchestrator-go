package services

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

type requestSchemaServiceImpl struct {
	repo            repositories.RequestSchemaRepo
	endpointService EndpointService
	logger          *zap.Logger
}

// NewRequestSchemaService creates a new request schema service implementation
func NewRequestSchemaService(repo repositories.RequestSchemaRepo, endpointService EndpointService, logger *zap.Logger) RequestSchemaService {
	return &requestSchemaServiceImpl{
		repo:            repo,
		endpointService: endpointService,
		logger:          logger,
	}
}

func (s *requestSchemaServiceImpl) CreateRequestSchema(schema *models.RequestSchema) error {
	s.logger.Info("Creating request schema", 
		zap.String("endpoint_id", schema.EndpointID),
		zap.String("param_name", schema.ParamName))

	// Validate endpoint exists using service
	_, err := s.endpointService.GetEndpointByID(schema.EndpointID)
	if err != nil {
		return fmt.Errorf("endpoint not found: %w", err)
	}

	// Business logic validations
	if schema.ParamName == "" {
		return fmt.Errorf("param name cannot be empty")
	}
	if schema.ParamLocation == "" {
		return fmt.Errorf("param location cannot be empty")
	}
	if schema.ParamType == "" {
		return fmt.Errorf("param type cannot be empty")
	}

	// Validate param location
	validLocations := map[string]bool{
		models.ParamLocationPath:   true,
		models.ParamLocationQuery:  true,
		models.ParamLocationBody:   true,
		models.ParamLocationHeader: true,
	}
	if !validLocations[schema.ParamLocation] {
		return fmt.Errorf("invalid param location: %s (must be path, query, body, or header)", schema.ParamLocation)
	}

	// Validate param type
	validTypes := map[string]bool{
		models.ParamTypeString:  true,
		models.ParamTypeNumber:  true,
		models.ParamTypeInteger: true,
		models.ParamTypeBoolean: true,
		models.ParamTypeObject:  true,
		models.ParamTypeArray:   true,
	}
	if !validTypes[schema.ParamType] {
		return fmt.Errorf("invalid param type: %s (must be string, number, integer, boolean, object, or array)", schema.ParamType)
	}

	err = s.repo.Create(schema)
	if err != nil {
		s.logger.Error("Failed to create request schema", zap.Error(err))
		return err
	}

	s.logger.Info("Request schema created successfully")
	return nil
}

func (s *requestSchemaServiceImpl) GetRequestSchemaByID(id string) (*models.RequestSchema, error) {
	s.logger.Debug("Fetching request schema by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *requestSchemaServiceImpl) GetRequestSchemasByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	s.logger.Debug("Fetching request schemas by endpoint ID", zap.String("endpoint_id", endpointID))
	return s.repo.FindByEndpointID(endpointID)
}

func (s *requestSchemaServiceImpl) GetRequestSchemasByEndpointAndLocation(endpointID string, location string) ([]*models.RequestSchema, error) {
	s.logger.Debug("Fetching request schemas by endpoint and location",
		zap.String("endpoint_id", endpointID),
		zap.String("location", location))
	return s.repo.FindByEndpointIDAndLocation(endpointID, location)
}

func (s *requestSchemaServiceImpl) GetRequiredSchemasByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	s.logger.Debug("Fetching required request schemas by endpoint ID", zap.String("endpoint_id", endpointID))
	return s.repo.FindRequiredByEndpointID(endpointID)
}

func (s *requestSchemaServiceImpl) UpdateRequestSchema(schema *models.RequestSchema) error {
	s.logger.Info("Updating request schema", zap.String("id", schema.ID))

	// Validate schema exists
	existing, err := s.repo.FindByID(schema.ID)
	if err != nil {
		return fmt.Errorf("request schema not found: %w", err)
	}

	// Validate endpoint exists if changed
	if schema.EndpointID != existing.EndpointID {
		_, err := s.endpointService.GetEndpointByID(schema.EndpointID)
		if err != nil {
			return fmt.Errorf("endpoint not found: %w", err)
		}
	}

	// Business logic validations
	if schema.ParamName == "" {
		return fmt.Errorf("param name cannot be empty")
	}
	if schema.ParamLocation == "" {
		return fmt.Errorf("param location cannot be empty")
	}
	if schema.ParamType == "" {
		return fmt.Errorf("param type cannot be empty")
	}

	err = s.repo.Update(schema)
	if err != nil {
		s.logger.Error("Failed to update request schema", zap.Error(err))
		return err
	}

	s.logger.Info("Request schema updated successfully")
	return nil
}

func (s *requestSchemaServiceImpl) DeleteRequestSchema(id string) error {
	s.logger.Info("Deleting request schema", zap.String("id", id))

	// Validate schema exists
	_, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("request schema not found: %w", err)
	}

	err = s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete request schema", zap.Error(err))
		return err
	}

	s.logger.Info("Request schema deleted successfully")
	return nil
}
