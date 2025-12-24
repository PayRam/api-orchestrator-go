package models

import (
	"time"
)

// Endpoint represents an API endpoint for a provider.
// This model enables the request builder to dynamically construct API URLs and methods
// based on database-driven configurations.
type Endpoint struct {
	// ID is the unique identifier for the endpoint (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// ProviderID is the foreign key reference to the Provider.
	ProviderID string `gorm:"type:uuid;not null;index:idx_endpoint_provider_id" json:"provider_id" db:"provider_id"`

	// Provider is the associated Provider entity (foreign key relationship).
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`

	// Name is the unique identifier for this endpoint within the provider (e.g., "create_order", "get_order").
	// Used for programmatic lookups and references.
	Name string `gorm:"type:varchar(255);not null;index:idx_endpoint_name" json:"name" db:"name"`

	// Method is the HTTP method for this endpoint (e.g., "GET", "POST", "PUT", "DELETE").
	Method string `gorm:"type:varchar(10);not null" json:"method" db:"method"`

	// Path is the URL path for this endpoint (e.g., "/api/v1/orders/{id}").
	// May contain path parameters in {param} format.
	Path string `gorm:"type:varchar(500);not null" json:"path" db:"path"`

	// BaseURL is an optional override for the provider's base URL.
	// If empty, the provider's default base URL is used.
	BaseURL string `gorm:"type:varchar(500)" json:"base_url" db:"base_url"`

	// Description provides additional context about the endpoint.
	Description string `gorm:"type:text" json:"description" db:"description"`

	// CreatedAt is the timestamp when the endpoint was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the endpoint was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns "endpoints" as the database table name.
func (Endpoint) TableName() string {
	return "endpoints"
}
