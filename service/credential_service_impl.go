package service

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/repository"
	"go.uber.org/zap"
)

type credentialServiceImpl struct {
	repo            repository.CredentialRepo
	providerService ProviderService
	logger          *zap.Logger
}

// NewCredentialService creates a new credential service implementation
func NewCredentialService(repo repository.CredentialRepo, providerService ProviderService, logger *zap.Logger) CredentialService {
	return &credentialServiceImpl{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
}

func (s *credentialServiceImpl) CreateCredential(credential *model.ProviderCredential) error {
	s.logger.Info("Creating credential", zap.String("key", credential.Key))

	// Validate provider exists using service (not repo directly)
	_, err := s.providerService.GetProviderByID(credential.ProviderID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Business logic validations
	if credential.Key == "" {
		return fmt.Errorf("credential key cannot be empty")
	}
	if credential.Value == "" {
		return fmt.Errorf("credential value cannot be empty")
	}

	err = s.repo.Create(credential)
	if err != nil {
		s.logger.Error("Failed to create credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential created successfully")
	return nil
}

func (s *credentialServiceImpl) GetCredentialByID(id string) (*model.ProviderCredential, error) {
	s.logger.Debug("Fetching credential by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *credentialServiceImpl) GetCredentialsByProviderID(providerID string) ([]*model.ProviderCredential, error) {
	s.logger.Debug("Fetching credentials by provider ID", zap.String("provider_id", providerID))
	return s.repo.FindByProviderID(providerID)
}

func (s *credentialServiceImpl) GetCredentialValue(providerID string, key string) (string, error) {
	s.logger.Debug("Fetching credential value", zap.String("provider_id", providerID), zap.String("key", key))

	cred, err := s.repo.FindByProviderIDAndKey(providerID, key)
	if err != nil {
		return "", fmt.Errorf("credential not found: %w", err)
	}

	return cred.Value, nil
}

func (s *credentialServiceImpl) UpdateCredential(credential *model.ProviderCredential) error {
	s.logger.Info("Updating credential", zap.String("id", credential.ID))

	err := s.repo.Update(credential)
	if err != nil {
		s.logger.Error("Failed to update credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential updated successfully")
	return nil
}

func (s *credentialServiceImpl) DeleteCredential(id string) error {
	s.logger.Info("Deleting credential", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential deleted successfully")
	return nil
}
