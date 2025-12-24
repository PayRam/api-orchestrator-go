package services

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type credentialServiceImpl struct {
	repo            repositories.CredentialRepo
	providerService ProviderService
	logger          *zap.Logger
}

// NewCredentialService creates a new credential service implementation
func NewCredentialService(repo repositories.CredentialRepo, providerService ProviderService, logger *zap.Logger) CredentialService {
	return &credentialServiceImpl{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
}

func (s *credentialServiceImpl) CreateCredential(credential *models.Credential) error {
	s.logger.Info("Creating credential", zap.String("key", credential.Key))

	// Validate provider exists using service
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

	// Generate ID if not set
	if credential.ID == "" {
		credential.ID = uuid.New().String()
	}

	err = s.repo.Create(credential)
	if err != nil {
		s.logger.Error("Failed to create credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential created successfully", zap.String("id", credential.ID))
	return nil
}

func (s *credentialServiceImpl) CreateCredentialWithPlainValue(
	providerID, key, plainValue, description string,
	required bool,
) (*models.Credential, error) {
	s.logger.Info("Creating credential with plain value",
		zap.String("provider_id", providerID),
		zap.String("key", key))

	// Validate provider exists
	_, err := s.providerService.GetProviderByID(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Validate inputs
	if key == "" {
		return nil, fmt.Errorf("credential key cannot be empty")
	}
	if plainValue == "" {
		return nil, fmt.Errorf("credential value cannot be empty")
	}

	// Encrypt the value
	encryptedValue, err := models.EncryptValue(plainValue)
	if err != nil {
		s.logger.Error("Failed to encrypt credential value", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	credential := &models.Credential{
		ID:          uuid.New().String(),
		ProviderID:  providerID,
		Key:         key,
		Value:       encryptedValue,
		Required:    required,
		Description: description,
	}

	err = s.repo.Create(credential)
	if err != nil {
		s.logger.Error("Failed to create credential", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Credential created successfully", zap.String("id", credential.ID))
	return credential, nil
}

func (s *credentialServiceImpl) GetCredentialByID(id string) (*models.Credential, error) {
	s.logger.Debug("Fetching credential by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *credentialServiceImpl) GetCredentialsByProviderID(providerID string) ([]*models.Credential, error) {
	s.logger.Debug("Fetching credentials by provider ID", zap.String("provider_id", providerID))
	return s.repo.FindByProviderID(providerID)
}

func (s *credentialServiceImpl) GetDecryptedCredentialsByProviderID(providerID string) (map[string]string, error) {
	s.logger.Debug("Fetching decrypted credentials by provider ID", zap.String("provider_id", providerID))

	credentials, err := s.repo.FindByProviderID(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials: %w", err)
	}

	result := make(map[string]string)
	for _, cred := range credentials {
		decryptedValue, err := cred.GetDecryptedValue()
		if err != nil {
			s.logger.Warn("Failed to decrypt credential",
				zap.String("key", cred.Key),
				zap.Error(err))
			continue
		}
		result[cred.Key] = decryptedValue
	}

	s.logger.Debug("Decrypted credentials retrieved",
		zap.String("provider_id", providerID),
		zap.Int("count", len(result)))

	return result, nil
}

func (s *credentialServiceImpl) GetCredentialValue(providerID string, key string) (string, error) {
	s.logger.Debug("Fetching credential value",
		zap.String("provider_id", providerID),
		zap.String("key", key))

	cred, err := s.repo.FindByProviderIDAndKey(providerID, key)
	if err != nil {
		return "", fmt.Errorf("credential not found: %w", err)
	}

	// Decrypt the value before returning
	decryptedValue, err := cred.GetDecryptedValue()
	if err != nil {
		s.logger.Error("Failed to decrypt credential value",
			zap.String("key", key),
			zap.Error(err))
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return decryptedValue, nil
}

func (s *credentialServiceImpl) UpdateCredential(credential *models.Credential) error {
	s.logger.Info("Updating credential", zap.String("id", credential.ID))

	err := s.repo.Update(credential)
	if err != nil {
		s.logger.Error("Failed to update credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential updated successfully")
	return nil
}

func (s *credentialServiceImpl) UpdateCredentialValue(id string, plainValue string) error {
	s.logger.Info("Updating credential value", zap.String("id", id))

	// Fetch existing credential
	credential, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("credential not found: %w", err)
	}

	// Encrypt the new value
	err = credential.SetEncryptedValue(plainValue)
	if err != nil {
		s.logger.Error("Failed to encrypt new value", zap.Error(err))
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	err = s.repo.Update(credential)
	if err != nil {
		s.logger.Error("Failed to update credential", zap.Error(err))
		return err
	}

	s.logger.Info("Credential value updated successfully")
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
