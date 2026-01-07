package models

import (
	"time"
)

// HeaderRule defines rules for building HTTP headers dynamically for each provider.
// This model stores ordered, flexible rules for constructing request headers
// based on credentials, strategies, or fixed expressions.
type HeaderRule struct {
	// ID is the unique identifier for the header rule (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// ProviderID is the foreign key reference to the Provider.
	ProviderID string `gorm:"type:uuid;not null;index:idx_header_rule_provider_id" json:"provider_id" db:"provider_id"`

	// Provider is the associated Provider entity (foreign key relationship).
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`

	// HeaderName is the HTTP header name (e.g., "Authorization", "x-api-key", "Content-Type").
	HeaderName string `gorm:"type:varchar(255);not null" json:"header_name" db:"header_name"`

	// ValueExpression defines how the header value is constructed.
	// Examples: "CREDENTIAL.API_KEY", "STRATEGY.sig", "Bearer ${CREDENTIAL.TOKEN}"
	ValueExpression string `gorm:"type:text;not null" json:"value_expression" db:"value_expression"`

	// Priority determines the order of header generation during pipeline execution.
	// Lower values are processed first.
	Priority int `gorm:"default:0;index:idx_header_rule_priority" json:"priority" db:"priority"`

	// CreatedAt is the timestamp when the header rule was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the header rule was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (HeaderRule) TableName() string {
	return GetTableName("header_rules")
}
