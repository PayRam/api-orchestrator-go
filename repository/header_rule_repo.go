package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// HeaderRuleRepo defines the interface for header rule data operations
type HeaderRuleRepo interface {
	// Create creates a new header rule
	Create(rule *model.ProviderHeaderRule) error

	// FindByID retrieves a header rule by ID
	FindByID(id string) (*model.ProviderHeaderRule, error)

	// FindByProviderID retrieves all header rules for a provider (ordered by priority)
	FindByProviderID(providerID string) ([]*model.ProviderHeaderRule, error)

	// Update updates an existing header rule
	Update(rule *model.ProviderHeaderRule) error

	// Delete soft deletes a header rule
	Delete(id string) error
}
