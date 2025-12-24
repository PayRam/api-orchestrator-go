package services

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

type endpointServiceImpl struct {
	repo            repositories.EndpointRepo
	providerService ProviderService
	logger          *zap.Logger
}

// NewEndpointService creates a new endpoint service implementation
func NewEndpointService(repo repositories.EndpointRepo, providerService ProviderService, logger *zap.Logger) EndpointService {
	return &endpointServiceImpl{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
}

func (s *endpointServiceImpl) CreateEndpoint(endpoint *models.Endpoint) error {
	s.logger.Info("Creating endpoint", zap.String("name", endpoint.Name))

	// Validate provider exists using service
	_, err := s.providerService.GetProviderByID(endpoint.ProviderID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Business logic validations
	if endpoint.Name == "" {
		return fmt.Errorf("endpoint name cannot be empty")
	}
	if endpoint.Method == "" {
		return fmt.Errorf("endpoint method cannot be empty")
	}
	if endpoint.Path == "" {
		return fmt.Errorf("endpoint path cannot be empty")
	}

	err = s.repo.Create(endpoint)
	if err != nil {
		s.logger.Error("Failed to create endpoint", zap.Error(err))
		return err
	}

	s.logger.Info("Endpoint created successfully")
	return nil
}

func (s *endpointServiceImpl) GetEndpointByID(id string) (*models.Endpoint, error) {
	s.logger.Debug("Fetching endpoint by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *endpointServiceImpl) GetEndpointByProviderAndName(providerName string, name string) (*models.Endpoint, error) {
	s.logger.Debug("Fetching endpoint by provider and name",
		zap.String("provider", providerName),
		zap.String("name", name))

	// Use provider service to get provider
	provider, err := s.providerService.GetProviderByName(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	return s.repo.FindByProviderIDAndName(provider.ID, name)
}

func (s *endpointServiceImpl) GetEndpointsByProviderID(providerID string) ([]*models.Endpoint, error) {
	s.logger.Debug("Fetching endpoints by provider ID", zap.String("provider_id", providerID))
	return s.repo.FindByProviderID(providerID)
}

func (s *endpointServiceImpl) UpdateEndpoint(endpoint *models.Endpoint) error {
	s.logger.Info("Updating endpoint", zap.String("id", endpoint.ID))

	err := s.repo.Update(endpoint)
	if err != nil {
		s.logger.Error("Failed to update endpoint", zap.Error(err))
		return err
	}

	s.logger.Info("Endpoint updated successfully")
	return nil
}

func (s *endpointServiceImpl) DeleteEndpoint(id string) error {
	s.logger.Info("Deleting endpoint", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete endpoint", zap.Error(err))
		return err
	}

	s.logger.Info("Endpoint deleted successfully")
	return nil
}
