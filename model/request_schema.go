package model

// ProviderRequestSchema defines the structure for request parameters
type ProviderRequestSchema struct {
	ID            string `gorm:"primaryKey" json:"id"`
	EndpointID    string `gorm:"type:varchar(255);not null;index" json:"endpoint_id"`
	ParamName     string `gorm:"type:varchar(255);not null" json:"param_name"`    // e.g., "amount", "currency"
	ParamLocation string `gorm:"type:varchar(50);not null" json:"param_location"` // "path", "query", "body"
	ParamType     string `gorm:"type:varchar(50);not null" json:"param_type"`     // "string", "number", "boolean", "object"
	Required      bool   `gorm:"default:false" json:"required"`
	DefaultValue  string `gorm:"type:text" json:"default_value"`
	Description   string `gorm:"type:text" json:"description"`

	// Relations
	Endpoint ProviderEndpoint `gorm:"foreignKey:EndpointID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderRequestSchema) TableName() string {
	return "provider_request_schemas"
}
