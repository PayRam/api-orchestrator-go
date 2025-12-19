package service

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/repository"
	"go.uber.org/zap"
)

type providerServiceImpl struct {
	repo   repository.ProviderRepo
	logger *zap.Logger
}

// NewProviderService creates a new provider service implementation
func NewProviderService(repo repository.ProviderRepo, logger *zap.Logger) ProviderService {
	return &providerServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *providerServiceImpl) CreateProvider(provider *model.Provider) error {
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

func (s *providerServiceImpl) GetProviderByID(id string) (*model.Provider, error) {
	s.logger.Debug("Fetching provider by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *providerServiceImpl) GetProviderByName(name string) (*model.Provider, error) {
	s.logger.Debug("Fetching provider by name", zap.String("name", name))
	return s.repo.FindByName(name)
}

func (s *providerServiceImpl) ListProviders() ([]*model.Provider, error) {
	s.logger.Debug("Listing all providers")
	return s.repo.List()
}

func (s *providerServiceImpl) UpdateProvider(provider *model.Provider) error {
	s.logger.Info("Updating provider", zap.String("id", provider.ID))

	// Business logic validations
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	err := s.repo.Update(provider)
	if err != nil {
		s.logger.Error("Failed to update provider", zap.Error(err))
		return err
	}

	s.logger.Info("Provider updated successfully")
	return nil
}

func (s *providerServiceImpl) DeleteProvider(id string) error {
	s.logger.Info("Deleting provider", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete provider", zap.Error(err))
		return err
	}

	s.logger.Info("Provider deleted successfully")
	return nil
}
