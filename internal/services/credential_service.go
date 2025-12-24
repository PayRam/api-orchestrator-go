package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// CredentialService defines the business logic interface for credentials.
// It handles secure storage and retrieval of provider credentials with
// encryption/decryption support.
type CredentialService interface {
	// CreateCredential creates a new credential with encrypted value
	CreateCredential(credential *models.Credential) error

	// CreateCredentialWithPlainValue creates a credential by encrypting the plain value first
	CreateCredentialWithPlainValue(providerID, key, plainValue, description string, required bool) (*models.Credential, error)

	// GetCredentialByID retrieves a credential by ID (value remains encrypted)
	GetCredentialByID(id string) (*models.Credential, error)

	// GetCredentialsByProviderID retrieves all credentials for a provider (values remain encrypted)
	GetCredentialsByProviderID(providerID string) ([]*models.Credential, error)

	// GetDecryptedCredentialsByProviderID retrieves all credentials for a provider with decrypted values
	// Returns a map of key -> decrypted value for easy lookup
	GetDecryptedCredentialsByProviderID(providerID string) (map[string]string, error)

	// GetCredentialValue retrieves a specific decrypted credential value by provider ID and key
	GetCredentialValue(providerID string, key string) (string, error)

	// UpdateCredential updates an existing credential
	UpdateCredential(credential *models.Credential) error

	// UpdateCredentialValue updates only the value of a credential (encrypts before storing)
	UpdateCredentialValue(id string, plainValue string) error

	// DeleteCredential soft deletes a credential
	DeleteCredential(id string) error
}
