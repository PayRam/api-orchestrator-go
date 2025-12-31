package models

import (
	"testing"
)

func TestSetTablePrefix(t *testing.T) {
	// Reset for test
	ResetTablePrefix()

	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "with orch_ prefix",
			prefix:   "orch_",
			expected: "orch_providers",
		},
		{
			name:     "with orchestrator_ prefix",
			prefix:   "orchestrator_",
			expected: "orchestrator_providers",
		},
		{
			name:     "with empty prefix",
			prefix:   "",
			expected: "providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetTablePrefix()
			SetTablePrefix(tt.prefix)

			result := GetTableName("providers")
			if result != tt.expected {
				t.Errorf("GetTableName() = %v, want %v", result, tt.expected)
			}

			// Verify GetTablePrefix returns the set prefix
			if GetTablePrefix() != tt.prefix {
				t.Errorf("GetTablePrefix() = %v, want %v", GetTablePrefix(), tt.prefix)
			}
		})
	}
}

func TestTableNameMethods(t *testing.T) {
	ResetTablePrefix()
	SetTablePrefix("test_")

	tests := []struct {
		name     string
		model    interface{ TableName() string }
		expected string
	}{
		{"Provider", Provider{}, "test_providers"},
		{"Credential", Credential{}, "test_credentials"},
		{"Endpoint", Endpoint{}, "test_endpoints"},
		{"HeaderRule", HeaderRule{}, "test_header_rules"},
		{"Strategy", Strategy{}, "test_strategies"},
		{"RequestSchema", RequestSchema{}, "test_request_schemas"},
		{"RequestValue", RequestValue{}, "test_request_values"},
		{"ResponseMapping", ResponseMapping{}, "test_response_mappings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.TableName()
			if result != tt.expected {
				t.Errorf("%s.TableName() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSetTablePrefixOnlyOnce(t *testing.T) {
	ResetTablePrefix()

	SetTablePrefix("first_")
	SetTablePrefix("second_") // This should be ignored

	if GetTablePrefix() != "first_" {
		t.Errorf("SetTablePrefix should only work once, got %v", GetTablePrefix())
	}

	result := GetTableName("providers")
	if result != "first_providers" {
		t.Errorf("GetTableName() = %v, want first_providers", result)
	}
}
