package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
	"gorm.io/gorm"
)

type endpointRepoImpl struct {
	db *gorm.DB
}

// NewEndpointRepo creates a new endpoint repository implementation
func NewEndpointRepo(db *gorm.DB) EndpointRepo {
	return &endpointRepoImpl{db: db}
}

func (r *endpointRepoImpl) Create(endpoint *model.ProviderEndpoint) error {
	return r.db.Create(endpoint).Error
}

func (r *endpointRepoImpl) FindByID(id string) (*model.ProviderEndpoint, error) {
	var endpoint model.ProviderEndpoint
	err := r.db.Where("id = ?", id).First(&endpoint).Error
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *endpointRepoImpl) FindByProviderIDAndName(providerID string, name string) (*model.ProviderEndpoint, error) {
	var endpoint model.ProviderEndpoint
	err := r.db.Where("provider_id = ? AND name = ?", providerID, name).
		First(&endpoint).Error
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *endpointRepoImpl) FindByProviderID(providerID string) ([]*model.ProviderEndpoint, error) {
	var endpoints []*model.ProviderEndpoint
	err := r.db.Where("provider_id = ?", providerID).Find(&endpoints).Error
	return endpoints, err
}

func (r *endpointRepoImpl) Update(endpoint *model.ProviderEndpoint) error {
	return r.db.Save(endpoint).Error
}

func (r *endpointRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.ProviderEndpoint{}).Error
}
