package model

import (
	"gorm.io/datatypes"
)

// Strategy represents a pluggable strategy (e.g., signature generation, token refresh)
type Strategy struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(255);not null;unique;index" json:"name"` // e.g., "banxa_signature", "transak_token"
	StrategyType string         `gorm:"type:varchar(100);not null" json:"strategy_type"`     // "SIGNATURE", "TOKEN_REFRESH", "PAYLOAD"
	Config       datatypes.JSON `gorm:"type:jsonb" json:"config"`                            // Strategy-specific config as JSON
}

// TableName overrides the table name for GORM
func (Strategy) TableName() string {
	return "strategies"
}
