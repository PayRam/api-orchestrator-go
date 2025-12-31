package models

import (
	"time"
)

// Provider represents the main API provider entity (e.g., Banxa, Transak).
// This is the core model that defines external API providers in the dynamic orchestrator architecture.
type Provider struct {
	// ID is the unique identifier for the provider (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// Name is the unique internal identifier for the provider (e.g., "banxa", "transak").
	// Used for programmatic lookups and references.
	Name string `gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_name" json:"name" db:"name"`

	// DisplayName is the human-readable name for the provider (e.g., "Banxa", "Transak").
	// Used for UI display purposes.
	DisplayName string `gorm:"type:varchar(255)" json:"display_name" db:"display_name"`

	// PipelineType defines the type of pipeline used for this provider.
	// Currently always "generic" for the dynamic orchestrator architecture.
	PipelineType string `gorm:"type:varchar(100);not null;default:'generic'" json:"pipeline_type" db:"pipeline_type"`

	// BaseURL is the base URL for the provider's API (e.g., "https://api.banxa.com").
	// Used as the default base URL for endpoints that don't specify their own.
	BaseURL string `gorm:"type:varchar(500)" json:"base_url" db:"base_url"`

	// IsActive indicates whether the provider is currently active and available for use.
	IsActive bool `gorm:"default:true;index:idx_provider_is_active" json:"is_active" db:"is_active"`

	// CreatedAt is the timestamp when the provider was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the provider was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (Provider) TableName() string {
	return GetTableName("providers")
}
