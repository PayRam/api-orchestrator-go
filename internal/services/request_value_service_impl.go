package services

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

// requestValueServiceImpl implements RequestValueService
type requestValueServiceImpl struct {
	repo       repositories.RequestValueRepo
	schemaRepo repositories.RequestSchemaRepo
	logger     *zap.Logger
}

// NewRequestValueService creates a new request value service
func NewRequestValueService(
	repo repositories.RequestValueRepo,
	schemaRepo repositories.RequestSchemaRepo,
	logger *zap.Logger,
) RequestValueService {
	return &requestValueServiceImpl{
		repo:       repo,
		schemaRepo: schemaRepo,
		logger:     logger,
	}
}

// CreateRequestValue creates a new request value with validation
func (s *requestValueServiceImpl) CreateRequestValue(value *models.RequestValue) error {
	// Validate schema exists
	if _, err := s.schemaRepo.FindByID(value.SchemaID); err != nil {
		s.logger.Error("Failed to find schema for request value",
			zap.String("schema_id", value.SchemaID),
			zap.Error(err))
		return fmt.Errorf("schema not found: %w", err)
	}

	// Validate source type
	validSourceTypes := map[string]bool{
		models.SourceTypeStatic:     true,
		models.SourceTypeInput:      true,
		models.SourceTypeCredential: true,
		models.SourceTypeComputed:   true,
	}
	if !validSourceTypes[value.SourceType] {
		return fmt.Errorf("invalid source_type: %s (must be one of: static, input, credential, computed)", value.SourceType)
	}

	// Validate source_key is provided for non-static sources
	if value.SourceType != models.SourceTypeStatic && value.SourceKey == "" {
		return fmt.Errorf("source_key is required when source_type is '%s'", value.SourceType)
	}

	if err := s.repo.Create(value); err != nil {
		s.logger.Error("Failed to create request value",
			zap.String("id", value.ID),
			zap.String("schema_id", value.SchemaID),
			zap.Error(err))
		return fmt.Errorf("failed to create request value: %w", err)
	}

	s.logger.Info("Request value created",
		zap.String("id", value.ID),
		zap.String("schema_id", value.SchemaID),
		zap.String("source_type", value.SourceType))

	return nil
}

// GetRequestValueByID retrieves a request value by ID
func (s *requestValueServiceImpl) GetRequestValueByID(id string) (*models.RequestValue, error) {
	value, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get request value",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get request value: %w", err)
	}

	s.logger.Debug("Request value retrieved",
		zap.String("id", id),
		zap.String("schema_id", value.SchemaID))

	return value, nil
}

// GetRequestValuesBySchemaID retrieves all request values for a schema
func (s *requestValueServiceImpl) GetRequestValuesBySchemaID(schemaID string) ([]*models.RequestValue, error) {
	values, err := s.repo.FindBySchemaID(schemaID)
	if err != nil {
		s.logger.Error("Failed to get request values for schema",
			zap.String("schema_id", schemaID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get request values: %w", err)
	}

	s.logger.Debug("Request values retrieved for schema",
		zap.String("schema_id", schemaID),
		zap.Int("count", len(values)))

	return values, nil
}

// UpdateRequestValue updates an existing request value with validation
func (s *requestValueServiceImpl) UpdateRequestValue(value *models.RequestValue) error {
	// Validate schema exists
	if _, err := s.schemaRepo.FindByID(value.SchemaID); err != nil {
		s.logger.Error("Failed to find schema for request value",
			zap.String("schema_id", value.SchemaID),
			zap.Error(err))
		return fmt.Errorf("schema not found: %w", err)
	}

	// Validate source type
	validSourceTypes := map[string]bool{
		models.SourceTypeStatic:     true,
		models.SourceTypeInput:      true,
		models.SourceTypeCredential: true,
		models.SourceTypeComputed:   true,
	}
	if !validSourceTypes[value.SourceType] {
		return fmt.Errorf("invalid source_type: %s (must be one of: static, input, credential, computed)", value.SourceType)
	}

	// Validate source_key is provided for non-static sources
	if value.SourceType != models.SourceTypeStatic && value.SourceKey == "" {
		return fmt.Errorf("source_key is required when source_type is '%s'", value.SourceType)
	}

	if err := s.repo.Update(value); err != nil {
		s.logger.Error("Failed to update request value",
			zap.String("id", value.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update request value: %w", err)
	}

	s.logger.Info("Request value updated",
		zap.String("id", value.ID),
		zap.String("schema_id", value.SchemaID))

	return nil
}

// DeleteRequestValue deletes a request value by ID
func (s *requestValueServiceImpl) DeleteRequestValue(id string) error {
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete request value",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete request value: %w", err)
	}

	s.logger.Info("Request value deleted",
		zap.String("id", id))

	return nil
}
