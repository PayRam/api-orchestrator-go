package models

import (
	"time"

	"gorm.io/datatypes"
)

// RequestValue captures user-provided or configured values for API request parameters.
// This model plugs dynamic values into path/query/body request construction
// by joining with RequestSchema and Endpoint for resolution.
type RequestValue struct {
	// ID is the unique identifier for the request value (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// SchemaID is the foreign key reference to the RequestSchema.
	SchemaID string `gorm:"type:uuid;not null;index:idx_request_value_schema_id" json:"schema_id" db:"schema_id"`

	// Schema is the associated RequestSchema entity (foreign key relationship).
	Schema RequestSchema `gorm:"foreignKey:SchemaID;constraint:OnDelete:CASCADE" json:"-"`

	// Value stores the parameter value as JSON to support various data types.
	// For simple values, store as JSON string: "\"value\""
	// For complex values, store as JSON object: {"key": "value"}
	Value datatypes.JSON `gorm:"type:jsonb" json:"value" db:"value"`

	// SourceType indicates where the value comes from.
	// Valid values: "static", "input", "credential", "computed"
	// - static: hardcoded value
	// - input: from user input (ctx.Input)
	// - credential: from credentials (ctx.Credentials)
	// - computed: from intermediate values (ctx.Intermediate)
	SourceType string `gorm:"type:varchar(50);default:'static'" json:"source_type" db:"source_type"`

	// SourceKey is the key to look up when SourceType is not "static".
	// For input: the key in ctx.Input
	// For credential: the key in ctx.Credentials
	// For computed: the key in ctx.Intermediate
	SourceKey string `gorm:"type:varchar(255)" json:"source_key" db:"source_key"`

	// CreatedAt is the timestamp when the value was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the value was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (RequestValue) TableName() string {
	return GetTableName("request_values")
}

// SourceType constants for request values
const (
	SourceTypeStatic     = "static"
	SourceTypeInput      = "input"
	SourceTypeCredential = "credential"
	SourceTypeComputed   = "computed"
)
