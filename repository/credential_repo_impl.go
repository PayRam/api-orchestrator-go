package repository

import (
	"github.com/PayRam/api-orchestrator-go/model"
	"gorm.io/gorm"
)

type credentialRepoImpl struct {
	db *gorm.DB
}

// NewCredentialRepo creates a new credential repository implementation
func NewCredentialRepo(db *gorm.DB) CredentialRepo {
	return &credentialRepoImpl{db: db}
}

func (r *credentialRepoImpl) Create(credential *model.ProviderCredential) error {
	return r.db.Create(credential).Error
}

func (r *credentialRepoImpl) FindByID(id string) (*model.ProviderCredential, error) {
	var credential model.ProviderCredential
	err := r.db.Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *credentialRepoImpl) FindByProviderID(providerID string) ([]*model.ProviderCredential, error) {
	var credentials []*model.ProviderCredential
	err := r.db.Where("provider_id = ?", providerID).Find(&credentials).Error
	return credentials, err
}

func (r *credentialRepoImpl) FindByProviderIDAndKey(providerID string, key string) (*model.ProviderCredential, error) {
	var credential model.ProviderCredential
	err := r.db.Where("provider_id = ? AND key = ?", providerID, key).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *credentialRepoImpl) Update(credential *model.ProviderCredential) error {
	return r.db.Save(credential).Error
}

func (r *credentialRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.ProviderCredential{}).Error
}
