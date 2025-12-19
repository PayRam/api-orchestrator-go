package model

// Provider represents the main API provider entity (e.g., Banxa, Transak)
type Provider struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"type:varchar(255);not null;unique;index" json:"name"` // e.g., "banxa", "transak"
	DisplayName string `gorm:"type:varchar(255)" json:"display_name"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TableName overrides the table name for GORM
func (Provider) TableName() string {
	return "providers"
}
