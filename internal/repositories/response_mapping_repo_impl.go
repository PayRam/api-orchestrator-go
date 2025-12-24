package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type responseMappingRepoImpl struct {
	db *gorm.DB
}

// NewResponseMappingRepo creates a new response mapping repository implementation
func NewResponseMappingRepo(db *gorm.DB) ResponseMappingRepo {
	return &responseMappingRepoImpl{db: db}
}

func (r *responseMappingRepoImpl) Create(mapping *models.ResponseMapping) error {
	return r.db.Create(mapping).Error
}

func (r *responseMappingRepoImpl) FindByID(id string) (*models.ResponseMapping, error) {
	var mapping models.ResponseMapping
	err := r.db.Where("id = ?", id).First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *responseMappingRepoImpl) FindByProviderIDAndAction(providerID string, action string) ([]*models.ResponseMapping, error) {
	var mappings []*models.ResponseMapping
	err := r.db.Where("provider_id = ? AND action = ?", providerID, action).Find(&mappings).Error
	return mappings, err
}

func (r *responseMappingRepoImpl) Update(mapping *models.ResponseMapping) error {
	return r.db.Save(mapping).Error
}

func (r *responseMappingRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.ResponseMapping{}).Error
}
