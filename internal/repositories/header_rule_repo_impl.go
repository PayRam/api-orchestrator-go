package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type headerRuleRepoImpl struct {
	db *gorm.DB
}

// NewHeaderRuleRepo creates a new header rule repository implementation
func NewHeaderRuleRepo(db *gorm.DB) HeaderRuleRepo {
	return &headerRuleRepoImpl{db: db}
}

func (r *headerRuleRepoImpl) Create(rule *models.HeaderRule) error {
	return r.db.Create(rule).Error
}

func (r *headerRuleRepoImpl) FindByID(id string) (*models.HeaderRule, error) {
	var rule models.HeaderRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *headerRuleRepoImpl) FindByProviderID(providerID string) ([]*models.HeaderRule, error) {
	var rules []*models.HeaderRule
	err := r.db.Where("provider_id = ?", providerID).
		Order("priority ASC").
		Find(&rules).Error
	return rules, err
}

func (r *headerRuleRepoImpl) Update(rule *models.HeaderRule) error {
	return r.db.Save(rule).Error
}

func (r *headerRuleRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.HeaderRule{}).Error
}
