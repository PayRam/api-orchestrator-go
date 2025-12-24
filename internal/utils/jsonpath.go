package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// NormalizeJSONPathForGJSON converts standard JSONPath to gjson syntax
// Standard JSONPath: $.data.id, $.data[0].name
// gjson syntax: data.id, data.0.name
func NormalizeJSONPathForGJSON(path string) string {
	// Remove leading $. or $ from the path
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")

	// Remove leading dot if present
	path = strings.TrimPrefix(path, ".")

	return path
}

// ExtractJSONPath extracts a value from a map using JSONPath syntax
func ExtractJSONPath(data map[string]interface{}, path string) (interface{}, error) {
	// Convert map to JSON string
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Normalize the path for gjson
	gjsonPath := NormalizeJSONPathForGJSON(path)

	// Use gjson to extract value
	result := gjson.GetBytes(jsonBytes, gjsonPath)
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
