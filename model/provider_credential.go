package model

// ProviderCredential stores provider authentication credentials
type ProviderCredential struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ProviderID  string `gorm:"type:varchar(255);not null;index" json:"provider_id"`
	Key         string `gorm:"type:varchar(255);not null" json:"key"` // e.g., "api_key", "api_secret"
	Value       string `gorm:"type:text;not null" json:"value"`       // encrypted/sensitive data
	Required    bool   `gorm:"default:false" json:"required"`
	Description string `gorm:"type:text" json:"description"`

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderCredential) TableName() string {
	return "provider_credentials"
}
