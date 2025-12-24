package models

import (
	"time"

	"gorm.io/datatypes"
)

// Strategy represents a pluggable, reusable strategy for dynamic request building.
// Strategies can be used for signature generation, token refresh, payload generation, etc.
// This model is designed as a standalone, reusable component that can be referenced
// by header rules or request mappings across all providers.
type Strategy struct {
	// ID is the unique identifier for the strategy (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// Name is the unique identifier for the strategy (e.g., "alchemy_signature", "banxa_hmac").
	// Used for programmatic lookups and references.
	Name string `gorm:"type:varchar(255);not null;uniqueIndex:idx_strategy_name" json:"name" db:"name"`

	// StrategyType defines the category of the strategy.
	// Examples: "SIGNATURE", "TOKEN_REFRESH", "PAYLOAD", "HMAC", "JWT"
	StrategyType string `gorm:"type:varchar(100);not null;index:idx_strategy_type" json:"strategy_type" db:"strategy_type"`

	// Config stores dynamic, strategy-specific configuration as JSON.
	// This allows flexible configuration without schema changes.
	// Example: {"algorithm": "SHA256", "encoding": "base64", "secret_key_ref": "CREDENTIAL.API_SECRET"}
	Config datatypes.JSON `gorm:"type:jsonb" json:"config" db:"config"`

	// CreatedAt is the timestamp when the strategy was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the strategy was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns "strategies" as the database table name.
func (Strategy) TableName() string {
	return "strategies"
}
