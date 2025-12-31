package models

import (
	"fmt"
	"strings"
	"time"
)

// Action constants define the standard actions supported by the orchestrator.
// These are used to categorize response mappings by the type of API operation.
const (
	// ActionCreateOrder represents the action for creating a new order/transaction
	ActionCreateOrder = "create_order"

	// ActionGetOrder represents the action for retrieving order details
	ActionGetOrder = "get_order"

	// ActionGetQuote represents the action for getting price quotes
	ActionGetQuote = "get_quote"

	// ActionGetCountries represents the action for retrieving supported countries
	ActionGetCountries = "get_countries"

	// ActionGetPaymentMethods represents the action for retrieving payment methods
	ActionGetPaymentMethods = "get_payment_methods"

	// ActionGetCurrencies represents the action for retrieving supported currencies
	ActionGetCurrencies = "get_currencies"

	// ActionGetLimits represents the action for retrieving transaction limits
	ActionGetLimits = "get_limits"

	// ActionCreateWidgetURL represents the action for generating widget/checkout URLs
	ActionCreateWidgetURL = "create_widget_url"

	// ActionGetTransactionStatus represents the action for checking transaction status
	ActionGetTransactionStatus = "get_transaction_status"

	// ActionCancelOrder represents the action for canceling an order
	ActionCancelOrder = "cancel_order"
)

// ValidActions is a list of all valid action types
var ValidActions = []string{
	ActionCreateOrder,
	ActionGetOrder,
	ActionGetQuote,
	ActionGetCountries,
	ActionGetPaymentMethods,
	ActionGetCurrencies,
	ActionGetLimits,
	ActionCreateWidgetURL,
	ActionGetTransactionStatus,
	ActionCancelOrder,
}

// Transform constants define the standard transformation functions
const (
	// TransformUppercase converts the value to uppercase
	TransformUppercase = "uppercase"

	// TransformLowercase converts the value to lowercase
	TransformLowercase = "lowercase"

	// TransformToCents converts a decimal amount to cents (multiplies by 100)
	TransformToCents = "to_cents"

	// TransformFromCents converts cents to decimal amount (divides by 100)
	TransformFromCents = "from_cents"

	// TransformToString converts the value to a string
	TransformToString = "to_string"

	// TransformToInt converts the value to an integer
	TransformToInt = "to_int"

	// TransformToFloat converts the value to a float
	TransformToFloat = "to_float"

	// TransformToBool converts the value to a boolean
	TransformToBool = "to_bool"

	// TransformParseDate parses a date string (ISO 8601 format)
	TransformParseDate = "parse_date"

	// TransformParseTimestamp parses a Unix timestamp
	TransformParseTimestamp = "parse_timestamp"

	// TransformTrim trims whitespace from the value
	TransformTrim = "trim"

	// TransformBase64Decode decodes a base64 encoded string
	TransformBase64Decode = "base64_decode"

	// TransformJSONParse parses a JSON string into an object
	TransformJSONParse = "json_parse"
)

// ValidTransforms is a list of all valid transform types
var ValidTransforms = []string{
	TransformUppercase,
	TransformLowercase,
	TransformToCents,
	TransformFromCents,
	TransformToString,
	TransformToInt,
	TransformToFloat,
	TransformToBool,
	TransformParseDate,
	TransformParseTimestamp,
	TransformTrim,
	TransformBase64Decode,
	TransformJSONParse,
}

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
	// Supports standard JSONPath syntax: $.data.order.id, data.items[0].name, etc.
	SourceJSONPath string `gorm:"type:text;not null" json:"source_json_path"`

	// Transform is an optional transformation function to apply to the extracted value
	// Examples: "uppercase", "lowercase", "to_cents", "parse_date"
	Transform string `gorm:"type:varchar(50)" json:"transform"`

	// DefaultValue is the value to use if the source path doesn't exist or returns null
	DefaultValue string `gorm:"type:text" json:"default_value"`

	// IsRequired indicates if this field is required in the response
	// If true and the field is not found, an error will be raised
	IsRequired bool `gorm:"default:false" json:"is_required"`

	// Priority determines the order of evaluation when multiple mappings exist
	// Higher priority mappings are evaluated first
	Priority int `gorm:"default:0" json:"priority"`

	// Description provides documentation about what this mapping does
	Description string `gorm:"type:text" json:"description"`

	// Timestamps
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (ResponseMapping) TableName() string {
	return GetTableName("response_mappings")
}

// Validate validates the response mapping fields
func (rm *ResponseMapping) Validate() error {
	if rm.ProviderID == "" {
		return fmt.Errorf("provider_id is required")
	}
	if rm.Action == "" {
		return fmt.Errorf("action is required")
	}
	if !rm.IsValidAction() {
		return fmt.Errorf("invalid action: %s, must be one of: %v", rm.Action, ValidActions)
	}
	if rm.TargetField == "" {
		return fmt.Errorf("target_field is required")
	}
	if rm.SourceJSONPath == "" {
		return fmt.Errorf("source_json_path is required")
	}
	if rm.Transform != "" && !rm.IsValidTransform() {
		return fmt.Errorf("invalid transform: %s, must be one of: %v", rm.Transform, ValidTransforms)
	}
	return nil
}

// IsValidAction checks if the action is a valid action type
func (rm *ResponseMapping) IsValidAction() bool {
	for _, valid := range ValidActions {
		if rm.Action == valid {
			return true
		}
	}
	return false
}

// IsValidTransform checks if the transform is a valid transform type
func (rm *ResponseMapping) IsValidTransform() bool {
	if rm.Transform == "" {
		return true // Empty transform is valid (no transformation)
	}
	for _, valid := range ValidTransforms {
		if rm.Transform == valid {
			return true
		}
	}
	return false
}

// NormalizeJSONPath normalizes the JSONPath expression
// Ensures it starts with $ or is a simple dot notation path
func (rm *ResponseMapping) NormalizeJSONPath() string {
	path := strings.TrimSpace(rm.SourceJSONPath)
	if path == "" {
		return path
	}
	// If it doesn't start with $ or @, assume it's a root-level path
	if !strings.HasPrefix(path, "$") && !strings.HasPrefix(path, "@") {
		// Convert dot notation to JSONPath if needed
		if !strings.HasPrefix(path, ".") {
			path = "$." + path
		} else {
			path = "$" + path
		}
	}
	return path
}

// HasTransform returns true if a transform function is specified
func (rm *ResponseMapping) HasTransform() bool {
	return rm.Transform != ""
}

// HasDefaultValue returns true if a default value is specified
func (rm *ResponseMapping) HasDefaultValue() bool {
	return rm.DefaultValue != ""
}
