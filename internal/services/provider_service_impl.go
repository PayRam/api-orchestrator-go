package services

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

type providerServiceImpl struct {
	repo   repositories.ProviderRepo
	logger *zap.Logger
}

// NewProviderService creates a new provider service implementation.
// The service only depends on ProviderRepo - not on other repositories.
func NewProviderService(repo repositories.ProviderRepo, logger *zap.Logger) ProviderService {
	return &providerServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *providerServiceImpl) CreateProvider(provider *models.Provider) error {
	s.logger.Info("Creating provider", zap.String("name", provider.Name))

	// Business logic validations
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Check if provider already exists
	existing, _ := s.repo.FindByName(provider.Name)
	if existing != nil {
		return fmt.Errorf("provider with name %s already exists", provider.Name)
	}

	err := s.repo.Create(provider)
	if err != nil {
		s.logger.Error("Failed to create provider", zap.Error(err))
		return err
	}

	s.logger.Info("Provider created successfully", zap.String("id", provider.ID))
	return nil
}

func (s *providerServiceImpl) GetProviderByID(id string) (*models.Provider, error) {
	s.logger.Debug("Fetching provider by ID", zap.String("id", id))

	if id == "" {
		return nil, fmt.Errorf("provider ID cannot be empty")
	}

	return s.repo.FindByID(id)
}

func (s *providerServiceImpl) GetProviderByName(name string) (*models.Provider, error) {
	s.logger.Debug("Fetching provider by name", zap.String("name", name))

	if name == "" {
		return nil, fmt.Errorf("provider name cannot be empty")
	}

	return s.repo.FindByName(name)
}

func (s *providerServiceImpl) GetAllActive() ([]*models.Provider, error) {
	s.logger.Debug("Fetching all active providers")
	return s.repo.FindAllActive()
}

func (s *providerServiceImpl) ListProviders() ([]*models.Provider, error) {
	s.logger.Debug("Listing all providers")
	return s.repo.List()
}

func (s *providerServiceImpl) UpdateProvider(provider *models.Provider) error {
	s.logger.Info("Updating provider", zap.String("id", provider.ID))

	// Business logic validations
	if provider.ID == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Verify provider exists
	existing, err := s.repo.FindByID(provider.ID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("provider with ID %s not found", provider.ID)
	}

	err = s.repo.Update(provider)
	if err != nil {
		s.logger.Error("Failed to update provider", zap.Error(err))
		return err
	}

	s.logger.Info("Provider updated successfully")
	return nil
}

func (s *providerServiceImpl) DeleteProvider(id string) error {
	s.logger.Info("Deleting provider", zap.String("id", id))

	if id == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}

	// Verify provider exists
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("provider with ID %s not found", id)
	}

	err = s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete provider", zap.Error(err))
		return err
	}

	s.logger.Info("Provider deleted successfully")
	return nil
}
