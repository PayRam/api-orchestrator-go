package model

// ProviderHeaderRule defines rules for building HTTP headers dynamically
type ProviderHeaderRule struct {
	ID              string `gorm:"primaryKey" json:"id"`
	ProviderID      string `gorm:"type:varchar(255);not null;index" json:"provider_id"`
	HeaderName      string `gorm:"type:varchar(255);not null" json:"header_name"` // e.g., "Authorization", "X-API-Key"
	ValueExpression string `gorm:"type:text;not null" json:"value_expression"`    // e.g., "CREDENTIAL.API_KEY", "STRATEGY.sig"
	Priority        int    `gorm:"default:0" json:"priority"`                     // Order of application

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ProviderHeaderRule) TableName() string {
	return "provider_header_rules"
}
