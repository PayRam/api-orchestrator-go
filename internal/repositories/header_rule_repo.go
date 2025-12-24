package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// HeaderRuleRepo defines the interface for header rule data operations
type HeaderRuleRepo interface {
	// Create creates a new header rule
	Create(rule *models.HeaderRule) error

	// FindByID retrieves a header rule by ID
	FindByID(id string) (*models.HeaderRule, error)

	// FindByProviderID retrieves all header rules for a provider (ordered by priority)
	FindByProviderID(providerID string) ([]*models.HeaderRule, error)

	// Update updates an existing header rule
	Update(rule *models.HeaderRule) error

	// Delete soft deletes a header rule
	Delete(id string) error
}
