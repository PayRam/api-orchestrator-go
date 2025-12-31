package models

import (
	"time"
)

// RequestSchema defines the structure and location of request parameters for an endpoint.
// This model makes request generation fully configurable from the database,
// specifying where each parameter should be placed (path, query, body) and its data type.
type RequestSchema struct {
	// ID is the unique identifier for the request schema (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// EndpointID is the foreign key reference to the Endpoint.
	EndpointID string `gorm:"type:uuid;not null;index:idx_request_schema_endpoint_id" json:"endpoint_id" db:"endpoint_id"`

	// Endpoint is the associated Endpoint entity (foreign key relationship).
	Endpoint Endpoint `gorm:"foreignKey:EndpointID;constraint:OnDelete:CASCADE" json:"-"`

	// ParamName is the name of the parameter (e.g., "amount", "currency", "wallet_address").
	ParamName string `gorm:"type:varchar(255);not null" json:"param_name" db:"param_name"`

	// ParamLocation specifies where the parameter should be placed in the request.
	// Valid values: "path", "query", "body", "header"
	ParamLocation string `gorm:"type:varchar(50);not null" json:"param_location" db:"param_location"`

	// ParamType specifies the data type of the parameter.
	// Valid values: "string", "number", "integer", "boolean", "object", "array"
	ParamType string `gorm:"type:varchar(50);not null" json:"param_type" db:"param_type"`

	// Required indicates whether this parameter is mandatory for the request.
	Required bool `gorm:"default:false" json:"required" db:"required"`

	// DefaultValue is the default value to use if no value is provided.
	DefaultValue string `gorm:"type:text" json:"default_value" db:"default_value"`

	// Description provides additional context about the parameter.
	Description string `gorm:"type:text" json:"description" db:"description"`

	// CreatedAt is the timestamp when the schema was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the schema was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (RequestSchema) TableName() string {
	return GetTableName("request_schemas")
}

// ParamLocation constants for request schema
const (
	ParamLocationPath   = "path"
	ParamLocationQuery  = "query"
	ParamLocationBody   = "body"
	ParamLocationHeader = "header"
)

// ParamType constants for request schema
const (
	ParamTypeString  = "string"
	ParamTypeNumber  = "number"
	ParamTypeInteger = "integer"
	ParamTypeBoolean = "boolean"
	ParamTypeObject  = "object"
	ParamTypeArray   = "array"
)
