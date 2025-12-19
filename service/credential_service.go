package service

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// CredentialService defines the business logic interface for credentials
type CredentialService interface {
	// CreateCredential creates a new credential
	CreateCredential(credential *model.ProviderCredential) error

	// GetCredentialByID retrieves a credential by ID
	GetCredentialByID(id string) (*model.ProviderCredential, error)

	// GetCredentialsByProviderID retrieves all credentials for a provider
	GetCredentialsByProviderID(providerID string) ([]*model.ProviderCredential, error)

	// GetCredentialValue retrieves a specific credential value by provider ID and key
	GetCredentialValue(providerID string, key string) (string, error)

	// UpdateCredential updates an existing credential
	UpdateCredential(credential *model.ProviderCredential) error

	// DeleteCredential soft deletes a credential
	DeleteCredential(id string) error
}
