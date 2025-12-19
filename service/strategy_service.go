package service

import (
	"github.com/PayRam/api-orchestrator-go/model"
)

// StrategyService defines the business logic interface for strategies
type StrategyService interface {
	// CreateStrategy creates a new strategy
	CreateStrategy(strategy *model.Strategy) error

	// GetStrategyByID retrieves a strategy by ID
	GetStrategyByID(id string) (*model.Strategy, error)

	// GetStrategyByName retrieves a strategy by name
	GetStrategyByName(name string) (*model.Strategy, error)

	// ListStrategies retrieves all active strategies
	ListStrategies() ([]*model.Strategy, error)

	// GetStrategiesByType retrieves all strategies of a specific type
	GetStrategiesByType(strategyType string) ([]*model.Strategy, error)

	// UpdateStrategy updates an existing strategy
	UpdateStrategy(strategy *model.Strategy) error

	// DeleteStrategy soft deletes a strategy
	DeleteStrategy(id string) error
}
