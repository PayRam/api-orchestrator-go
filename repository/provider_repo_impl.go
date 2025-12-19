package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
	"gorm.io/gorm"
)

type providerRepoImpl struct {
	db *gorm.DB
}

// NewProviderRepo creates a new provider repository implementation
func NewProviderRepo(db *gorm.DB) ProviderRepo {
	return &providerRepoImpl{db: db}
}

func (r *providerRepoImpl) Create(provider *model.Provider) error {
	return r.db.Create(provider).Error
}

func (r *providerRepoImpl) FindByID(id string) (*model.Provider, error) {
	var provider model.Provider
	err := r.db.Where("id = ?", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepoImpl) FindByName(name string) (*model.Provider, error) {
	var provider model.Provider
	err := r.db.Where("name = ?", name).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *providerRepoImpl) List() ([]*model.Provider, error) {
	var providers []*model.Provider
	err := r.db.Where("is_active = ?", true).Find(&providers).Error
	return providers, err
}

func (r *providerRepoImpl) Update(provider *model.Provider) error {
	return r.db.Save(provider).Error
}

func (r *providerRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Provider{}).Error
}
