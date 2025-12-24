package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// StrategyRepo defines the interface for strategy data operations
type StrategyRepo interface {
	// Create creates a new strategy
	Create(strategy *models.Strategy) error

	// FindByID retrieves a strategy by ID
	FindByID(id string) (*models.Strategy, error)

	// FindByName retrieves a strategy by name
	FindByName(name string) (*models.Strategy, error)

	// List retrieves all active strategies
	List() ([]*models.Strategy, error)

	// FindByType retrieves all strategies of a specific type
	FindByType(strategyType string) ([]*models.Strategy, error)

	// Update updates an existing strategy
	Update(strategy *models.Strategy) error

	// Delete soft deletes a strategy
	Delete(id string) error
}
