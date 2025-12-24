package repositories

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/gorm"
)

type credentialRepoImpl struct {
	db *gorm.DB
}

// NewCredentialRepo creates a new credential repository implementation
func NewCredentialRepo(db *gorm.DB) CredentialRepo {
	return &credentialRepoImpl{db: db}
}

func (r *credentialRepoImpl) Create(credential *models.Credential) error {
	return r.db.Create(credential).Error
}

func (r *credentialRepoImpl) FindByID(id string) (*models.Credential, error) {
	var credential models.Credential
	err := r.db.Where("id = ?", id).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *credentialRepoImpl) FindByProviderID(providerID string) ([]*models.Credential, error) {
	var credentials []*models.Credential
	err := r.db.Where("provider_id = ?", providerID).Find(&credentials).Error
	return credentials, err
}

func (r *credentialRepoImpl) FindByProviderIDAndKey(providerID string, key string) (*models.Credential, error) {
	var credential models.Credential
	err := r.db.Where("provider_id = ? AND key = ?", providerID, key).First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

func (r *credentialRepoImpl) Update(credential *models.Credential) error {
	return r.db.Save(credential).Error
}

func (r *credentialRepoImpl) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Credential{}).Error
}
