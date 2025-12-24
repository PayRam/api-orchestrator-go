package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type strategyRepoImpl struct {
	db *gorm.DB
}

// NewStrategyRepo creates a new strategy repository implementation
func NewStrategyRepo(db *gorm.DB) StrategyRepo {
	return &strategyRepoImpl{db: db}
}

func (r *strategyRepoImpl) Create(strategy *models.Strategy) error {
	return r.db.Create(strategy).Error
}

func (r *strategyRepoImpl) FindByID(id string) (*models.Strategy, error) {
	var strategy models.Strategy
	err := r.db.Where("id = ?", id).First(&strategy).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *strategyRepoImpl) FindByName(name string) (*models.Strategy, error) {
	var strategy models.Strategy
	err := r.db.Where("name = ?", name).First(&strategy).Error
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *strategyRepoImpl) List() ([]*models.Strategy, error) {
	var strategies []*models.Strategy
	err := r.db.Find(&strategies).Error
	return strategies, err
}

func (r *strategyRepoImpl) FindByType(strategyType string) ([]*models.Strategy, error) {
	var strategies []*models.Strategy
	err := r.db.Where("strategy_type = ?", strategyType).Find(&strategies).Error
	return strategies, err
}

func (r *strategyRepoImpl) Update(strategy *models.Strategy) error {
	return r.db.Save(strategy).Error
}

func (r *strategyRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Strategy{}).Error
}
