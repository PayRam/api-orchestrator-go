package model

// ProviderRequestValue defines field value mappings
type ProviderRequestValue struct {
	ID       string `gorm:"primaryKey" json:"id"`
	SchemaID string `gorm:"type:varchar(255);not null;index" json:"schema_id"`
	Value    string `gorm:"type:text" json:"value"`

	// Relations
	Schema ProviderRequestSchema `gorm:"foreignKey:SchemaID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderRequestValue) TableName() string {
	return "provider_request_values"
}
