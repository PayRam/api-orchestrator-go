package utils

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// ExtractJSONPath extracts a value from a map using JSONPath syntax
func ExtractJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	// Convert map to JSON string
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Use gjson to extract value
	result := gjson.GetBytes(jsonBytes, path)
	if !result.Exists() {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	return result.Value(), nil
}

// ExtractJSONPathString extracts a string value using JSONPath
func ExtractJSONPathString(data map[string]interface{}, path string) (string, error) {
	value, err := ExtractJSONPath(data, path)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(value), nil
}
