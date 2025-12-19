package model

// ProviderResponseMapping defines how to normalize provider responses
type ProviderResponseMapping struct {
	ID             string `gorm:"primaryKey" json:"id"`
	ProviderID     string `gorm:"type:varchar(255);not null;index" json:"provider_id"`
	Action         string `gorm:"type:varchar(255);not null" json:"action"`       // e.g., "create_order", "get_order"
	TargetField    string `gorm:"type:varchar(255);not null" json:"target_field"` // Normalized field name
	SourceJSONPath string `gorm:"type:text;not null" json:"source_json_path"`     // JSONPath expression
	Transform      string `gorm:"type:varchar(255)" json:"transform"`             // Optional transformation function

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderResponseMapping) TableName() string {
	return "provider_response_mappings"
}
