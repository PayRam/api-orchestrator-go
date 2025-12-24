package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type requestSchemaRepoImpl struct {
	db *gorm.DB
}

// NewRequestSchemaRepo creates a new request schema repository
func NewRequestSchemaRepo(db *gorm.DB) RequestSchemaRepo {
	return &requestSchemaRepoImpl{db: db}
}

func (r *requestSchemaRepoImpl) Create(schema *models.RequestSchema) error {
	return r.db.Create(schema).Error
}

func (r *requestSchemaRepoImpl) FindByID(id string) (*models.RequestSchema, error) {
	var schema models.RequestSchema
	err := r.db.First(&schema, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (r *requestSchemaRepoImpl) FindByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	var schemas []*models.RequestSchema
	err := r.db.Where("endpoint_id = ?", endpointID).Find(&schemas).Error
	if err != nil {
		return nil, err
	}
	return schemas, nil
}

func (r *requestSchemaRepoImpl) FindByEndpointIDAndLocation(endpointID string, location string) ([]*models.RequestSchema, error) {
	var schemas []*models.RequestSchema
	err := r.db.Where("endpoint_id = ? AND param_location = ?", endpointID, location).Find(&schemas).Error
	if err != nil {
		return nil, err
	}
	return schemas, nil
}

func (r *requestSchemaRepoImpl) FindRequiredByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	var schemas []*models.RequestSchema
	err := r.db.Where("endpoint_id = ? AND required = ?", endpointID, true).Find(&schemas).Error
	if err != nil {
		return nil, err
	}
	return schemas, nil
}

func (r *requestSchemaRepoImpl) Update(schema *models.RequestSchema) error {
	return r.db.Save(schema).Error
}

func (r *requestSchemaRepoImpl) Delete(id string) error {
	return r.db.Delete(&models.RequestSchema{}, "id = ?", id).Error
}
