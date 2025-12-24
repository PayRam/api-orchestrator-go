package utils

import (
	"strings"
	"time"
)

// Transform applies a transformation function to a value
func Transform(value interface{}, transformFunc string) interface{} {
	switch transformFunc {
	case "to_upper":
		if s, ok := value.(string); ok {
			return strings.ToUpper(s)
		}
	case "to_lower":
		if s, ok := value.(string); ok {
			return strings.ToLower(s)
		}
	case "trim":
		if s, ok := value.(string); ok {
			return strings.TrimSpace(s)
		}
	case "format_date":
		// Example: format date
		if s, ok := value.(string); ok {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t.Format("2006-01-02")
			}
		}
	}
	return value
}
