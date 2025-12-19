package service

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/repository"
	"go.uber.org/zap"
)

type strategyServiceImpl struct {
	repo   repository.StrategyRepo
	logger *zap.Logger
}

// NewStrategyService creates a new strategy service implementation
func NewStrategyService(repo repository.StrategyRepo, logger *zap.Logger) StrategyService {
	return &strategyServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *strategyServiceImpl) CreateStrategy(strategy *model.Strategy) error {
	s.logger.Info("Creating strategy", zap.String("name", strategy.Name))

	// Business logic validations
	if strategy.Name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}
	if strategy.StrategyType == "" {
		return fmt.Errorf("strategy type cannot be empty")
	}

	// Check if strategy already exists
	existing, _ := s.repo.FindByName(strategy.Name)
	if existing != nil {
		return fmt.Errorf("strategy with name %s already exists", strategy.Name)
	}

	err := s.repo.Create(strategy)
	if err != nil {
		s.logger.Error("Failed to create strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy created successfully")
	return nil
}

func (s *strategyServiceImpl) GetStrategyByID(id string) (*model.Strategy, error) {
	s.logger.Debug("Fetching strategy by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *strategyServiceImpl) GetStrategyByName(name string) (*model.Strategy, error) {
	s.logger.Debug("Fetching strategy by name", zap.String("name", name))
	return s.repo.FindByName(name)
}

func (s *strategyServiceImpl) ListStrategies() ([]*model.Strategy, error) {
	s.logger.Debug("Listing all strategies")
	return s.repo.List()
}

func (s *strategyServiceImpl) GetStrategiesByType(strategyType string) ([]*model.Strategy, error) {
	s.logger.Debug("Fetching strategies by type", zap.String("type", strategyType))
	return s.repo.FindByType(strategyType)
}

func (s *strategyServiceImpl) UpdateStrategy(strategy *model.Strategy) error {
	s.logger.Info("Updating strategy", zap.String("id", strategy.ID))

	err := s.repo.Update(strategy)
	if err != nil {
		s.logger.Error("Failed to update strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy updated successfully")
	return nil
}

func (s *strategyServiceImpl) DeleteStrategy(id string) error {
	s.logger.Info("Deleting strategy", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy deleted successfully")
	return nil
}
