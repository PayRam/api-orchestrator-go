package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type requestValueRepoImpl struct {
	db *gorm.DB
}

// NewRequestValueRepo creates a new request value repository implementation
func NewRequestValueRepo(db *gorm.DB) RequestValueRepo {
	return &requestValueRepoImpl{db: db}
}

func (r *requestValueRepoImpl) Create(value *models.RequestValue) error {
	return r.db.Create(value).Error
}

func (r *requestValueRepoImpl) FindByID(id string) (*models.RequestValue, error) {
	var value models.RequestValue
	err := r.db.Where("id = ?", id).First(&value).Error
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *requestValueRepoImpl) FindBySchemaID(schemaID string) ([]*models.RequestValue, error) {
	var values []*models.RequestValue
	err := r.db.Where("schema_id = ?", schemaID).Find(&values).Error
	return values, err
}

func (r *requestValueRepoImpl) Update(value *models.RequestValue) error {
	return r.db.Save(value).Error
}

func (r *requestValueRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.RequestValue{}).Error
}
