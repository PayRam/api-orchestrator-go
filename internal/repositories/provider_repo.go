package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// ProviderRepo defines the interface for provider data operations.
// This interface allows plug-and-play fetching and caching of provider metadata.
type ProviderRepo interface {
	// Create creates a new provider
	Create(provider *models.Provider) error

	// FindByID retrieves a provider by ID
	FindByID(id string) (*models.Provider, error)

	// FindByName retrieves a provider by name
	FindByName(name string) (*models.Provider, error)

	// FindAllActive retrieves all active providers
	FindAllActive() ([]*models.Provider, error)

	// List retrieves all providers (including inactive)
	List() ([]*models.Provider, error)

	// Update updates an existing provider
	Update(provider *models.Provider) error

	// Delete deletes a provider by ID
	Delete(id string) error
}
