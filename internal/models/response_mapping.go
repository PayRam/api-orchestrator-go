package models

import (
	"time"
)

// ResponseMapping defines how to normalize provider responses into a unified format.
// It maps provider-specific JSON paths to standardized field names.
type ResponseMapping struct {
	// ID is the unique identifier for the response mapping
	ID string `gorm:"type:uuid;primaryKey" json:"id"`

	// ProviderID references the provider this mapping belongs to
	ProviderID string `gorm:"type:uuid;not null;index" json:"provider_id"`

	// Action is the action name this mapping applies to (e.g., "create_order", "get_order")
	Action string `gorm:"type:varchar(100);not null;index" json:"action"`

	// TargetField is the normalized/standardized field name in the unified response
	TargetField string `gorm:"type:varchar(100);not null" json:"target_field"`

	// SourceJSONPath is the JSONPath expression to extract the value from provider response
	SourceJSONPath string `gorm:"type:text;not null" json:"source_json_path"`

	// Transform is an optional transformation function to apply to the extracted value
	// Examples: "uppercase", "lowercase", "to_cents", "parse_date"
	Transform string `gorm:"type:varchar(50)" json:"transform"`

	// DefaultValue is the value to use if the source path doesn't exist
	DefaultValue string `gorm:"type:text" json:"default_value"`

	// IsRequired indicates if this field is required in the response
	IsRequired bool `gorm:"default:false" json:"is_required"`

	// Priority determines the order of evaluation when multiple mappings target the same field
	Priority int `gorm:"default:0" json:"priority"`

	// Timestamps
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM
func (ResponseMapping) TableName() string {
	return "response_mappings"
}
