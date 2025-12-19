package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// CredentialRepo defines the interface for credential data operations
type CredentialRepo interface {
	// Create creates a new credential
	Create(credential *model.ProviderCredential) error

	// FindByID retrieves a credential by ID
	FindByID(id string) (*model.ProviderCredential, error)

	// FindByProviderID retrieves all credentials for a provider
	FindByProviderID(providerID string) ([]*model.ProviderCredential, error)

	// FindByProviderIDAndKey retrieves a specific credential by provider ID and key
	FindByProviderIDAndKey(providerID string, key string) (*model.ProviderCredential, error)

	// Update updates an existing credential
	Update(credential *model.ProviderCredential) error

	// Delete soft deletes a credential
	Delete(id string) error
}
