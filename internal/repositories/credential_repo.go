package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// CredentialRepo defines the interface for credential data operations
type CredentialRepo interface {
	// Create creates a new credential
	Create(credential *models.Credential) error

	// FindByID retrieves a credential by ID
	FindByID(id string) (*models.Credential, error)

	// FindByProviderID retrieves all credentials for a provider
	FindByProviderID(providerID string) ([]*models.Credential, error)

	// FindByProviderIDAndKey retrieves a specific credential by provider ID and key
	FindByProviderIDAndKey(providerID string, key string) (*models.Credential, error)

	// Update updates an existing credential
	Update(credential *models.Credential) error

	// Delete soft deletes a credential
	Delete(id string) error
}
