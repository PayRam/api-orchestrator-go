package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
	"gorm.io/gorm"
)

type requestValueRepoImpl struct {
	db *gorm.DB
}

// NewRequestValueRepo creates a new request value repository implementation
func NewRequestValueRepo(db *gorm.DB) RequestValueRepo {
	return &requestValueRepoImpl{db: db}
}

func (r *requestValueRepoImpl) Create(value *model.ProviderRequestValue) error {
	return r.db.Create(value).Error
}

func (r *requestValueRepoImpl) FindByID(id string) (*model.ProviderRequestValue, error) {
	var value model.ProviderRequestValue
	err := r.db.Where("id = ?", id).First(&value).Error
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *requestValueRepoImpl) FindBySchemaID(schemaID string) ([]*model.ProviderRequestValue, error) {
	var values []*model.ProviderRequestValue
	err := r.db.Where("schema_id = ?", schemaID).Find(&values).Error
	return values, err
}

func (r *requestValueRepoImpl) Update(value *model.ProviderRequestValue) error {
	return r.db.Save(value).Error
}

func (r *requestValueRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.ProviderRequestValue{}).Error
}
