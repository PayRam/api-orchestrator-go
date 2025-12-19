package model

// ProviderEndpoint represents an API endpoint for a provider
type ProviderEndpoint struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ProviderID  string `gorm:"type:varchar(255);not null;index" json:"provider_id"`
	Name        string `gorm:"type:varchar(255);not null;index" json:"name"` // e.g., "create_order", "get_order"
	Method      string `gorm:"type:varchar(10);not null" json:"method"`      // "GET", "POST", "PUT", "DELETE"
	Path        string `gorm:"type:varchar(500);not null" json:"path"`       // e.g., "/api/v1/orders/{id}"
	BaseURL     string `gorm:"type:varchar(500)" json:"base_url"`            // Optional override for provider base URL
	Description string `gorm:"type:text" json:"description"`

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderEndpoint) TableName() string {
	return "provider_endpoints"
}
