package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// StrategyRepo defines the interface for strategy data operations
type StrategyRepo interface {
	// Create creates a new strategy
	Create(strategy *model.Strategy) error

	// FindByID retrieves a strategy by ID
	FindByID(id string) (*model.Strategy, error)

	// FindByName retrieves a strategy by name
	FindByName(name string) (*model.Strategy, error)

	// List retrieves all active strategies
	List() ([]*model.Strategy, error)

	// FindByType retrieves all strategies of a specific type
	FindByType(strategyType string) ([]*model.Strategy, error)

	// Update updates an existing strategy
	Update(strategy *model.Strategy) error

	// Delete soft deletes a strategy
	Delete(id string) error
}
