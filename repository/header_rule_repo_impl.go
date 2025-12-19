package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
	"gorm.io/gorm"
)

type headerRuleRepoImpl struct {
	db *gorm.DB
}

// NewHeaderRuleRepo creates a new header rule repository implementation
func NewHeaderRuleRepo(db *gorm.DB) HeaderRuleRepo {
	return &headerRuleRepoImpl{db: db}
}

func (r *headerRuleRepoImpl) Create(rule *model.ProviderHeaderRule) error {
	return r.db.Create(rule).Error
}

func (r *headerRuleRepoImpl) FindByID(id string) (*model.ProviderHeaderRule, error) {
	var rule model.ProviderHeaderRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *headerRuleRepoImpl) FindByProviderID(providerID string) ([]*model.ProviderHeaderRule, error) {
	var rules []*model.ProviderHeaderRule
	err := r.db.Where("provider_id = ?", providerID).
		Order("priority ASC").
		Find(&rules).Error
	return rules, err
}

func (r *headerRuleRepoImpl) Update(rule *model.ProviderHeaderRule) error {
	return r.db.Save(rule).Error
}

func (r *headerRuleRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.ProviderHeaderRule{}).Error
}
