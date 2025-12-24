package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Transformer handles value transformations for response normalization
type Transformer struct{}

// NewTransformer creates a new Transformer instance
func NewTransformer() *Transformer {
	return &Transformer{}
}

// Transform applies a transformation function to a value
func Transform(value interface{}, transformFunc string) interface{} {
	t := NewTransformer()
	return t.Apply(value, transformFunc)
}

// Apply applies a transformation function to a value
func (t *Transformer) Apply(value interface{}, transformFunc string) interface{} {
	if value == nil {
		return nil
	}

	// Handle parameterized transforms (e.g., "concat:prefix:suffix")
	parts := strings.SplitN(transformFunc, ":", 2)
	funcName := parts[0]
	var params string
	if len(parts) > 1 {
		params = parts[1]
	}

	switch funcName {
	case "uppercase", "to_upper":
		return t.toUppercase(value)
	case "lowercase", "to_lower":
		return t.toLowercase(value)
	case "trim":
		return t.trim(value)
	case "to_string":
		return t.toString(value)
	case "to_int":
		return t.toInt(value)
	case "to_float":
		return t.toFloat(value)
	case "to_bool":
		return t.toBool(value)
	case "to_cents":
		return t.toCents(value)
	case "from_cents":
		return t.fromCents(value)
	case "parse_date":
		return t.parseDate(value)
	case "parse_timestamp":
		return t.parseTimestamp(value)
	case "format_date":
		return t.formatDate(value)
	case "base64_decode":
		return t.base64Decode(value)
	case "base64_encode":
		return t.base64Encode(value)
	case "json_parse":
		return t.jsonParse(value)
	case "json_stringify":
		return t.jsonStringify(value)
	case "concat":
		return t.concat(value, params)
	case "prefix":
		return t.prefix(value, params)
	case "suffix":
		return t.suffix(value, params)
	case "replace":
		return t.replace(value, params)
	case "split":
		return t.split(value, params)
	case "index":
		return t.index(value, params)
	case "round":
		return t.round(value, params)
	case "floor":
		return t.floor(value)
	case "ceil":
		return t.ceil(value)
	case "abs":
		return t.abs(value)
	case "default":
		return t.defaultValue(value, params)
	default:
		return value
	}
}

// toUppercase converts the value to uppercase
func (t *Transformer) toUppercase(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(fmt.Sprint(value))
}

// toLowercase converts the value to lowercase
func (t *Transformer) toLowercase(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		return strings.ToLower(s)
	}
	return strings.ToLower(fmt.Sprint(value))
}

// trim removes whitespace from the value
func (t *Transformer) trim(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

// toString converts the value to a string
func (t *Transformer) toString(value interface{}) interface{} {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case float64:
		// Avoid scientific notation for large numbers
		if v == math.Floor(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprint(v)
	}
}

// toInt converts the value to an integer
func (t *Transformer) toInt(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		// Try to parse as float first (handles "123.45")
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f)
		}
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

// toFloat converts the value to a float64
func (t *Transformer) toFloat(value interface{}) interface{} {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// toBool converts the value to a boolean
func (t *Transformer) toBool(value interface{}) interface{} {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		return lower == "true" || lower == "1" || lower == "yes" || lower == "on"
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}

// toCents converts a decimal amount to cents (multiplies by 100)
func (t *Transformer) toCents(value interface{}) interface{} {
	f := t.toFloat(value).(float64)
	return int64(math.Round(f * 100))
}

// fromCents converts cents to decimal amount (divides by 100)
func (t *Transformer) fromCents(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return float64(v) / 100.0
	case int64:
		return float64(v) / 100.0
	case float64:
		return v / 100.0
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return float64(i) / 100.0
		}
	}
	return 0.0
}

// parseDate parses a date string (ISO 8601 format)
func (t *Transformer) parseDate(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		// Try multiple common formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02/01/2006",
			"01/02/2006",
			"Jan 2, 2006",
		}
		for _, format := range formats {
			if parsed, err := time.Parse(format, s); err == nil {
				return parsed
			}
		}
	}
	return value
}

// parseTimestamp parses a Unix timestamp (seconds or milliseconds)
func (t *Transformer) parseTimestamp(value interface{}) interface{} {
	var ts int64
	switch v := value.(type) {
	case int:
		ts = int64(v)
	case int64:
		ts = v
	case float64:
		ts = int64(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			ts = parsed
		} else {
			return value
		}
	default:
		return value
	}

	// If timestamp is in milliseconds (> year 3000 in seconds), convert to seconds
	if ts > 32503680000 {
		ts = ts / 1000
	}
	return time.Unix(ts, 0)
}

// formatDate formats a time value to ISO 8601 date string
func (t *Transformer) formatDate(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		if parsed := t.parseDate(v); parsed != v {
			if tm, ok := parsed.(time.Time); ok {
				return tm.Format("2006-01-02")
			}
		}
		return v
	}
	return value
}

// base64Decode decodes a base64 encoded string
func (t *Transformer) base64Decode(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
			return string(decoded)
		}
		// Try URL-safe encoding
		if decoded, err := base64.URLEncoding.DecodeString(s); err == nil {
			return string(decoded)
		}
	}
	return value
}

// base64Encode encodes a string to base64
func (t *Transformer) base64Encode(value interface{}) interface{} {
	s := fmt.Sprint(value)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// jsonParse parses a JSON string into an object
func (t *Transformer) jsonParse(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		var result interface{}
		if err := json.Unmarshal([]byte(s), &result); err == nil {
			return result
		}
	}
	return value
}

// jsonStringify converts a value to a JSON string
func (t *Transformer) jsonStringify(value interface{}) interface{} {
	if bytes, err := json.Marshal(value); err == nil {
		return string(bytes)
	}
	return fmt.Sprint(value)
}

// concat concatenates prefix and suffix to the value
// params format: "prefix:suffix" or just "prefix"
func (t *Transformer) concat(value interface{}, params string) interface{} {
	s := fmt.Sprint(value)
	parts := strings.SplitN(params, ":", 2)
	prefix := ""
	suffix := ""
	if len(parts) >= 1 {
		prefix = parts[0]
	}
	if len(parts) >= 2 {
		suffix = parts[1]
	}
	return prefix + s + suffix
}

// prefix adds a prefix to the value
func (t *Transformer) prefix(value interface{}, params string) interface{} {
	return params + fmt.Sprint(value)
}

// suffix adds a suffix to the value
func (t *Transformer) suffix(value interface{}, params string) interface{} {
	return fmt.Sprint(value) + params
}

// replace replaces occurrences in the string
// params format: "old:new"
func (t *Transformer) replace(value interface{}, params string) interface{} {
	s := fmt.Sprint(value)
	parts := strings.SplitN(params, ":", 2)
	if len(parts) == 2 {
		return strings.ReplaceAll(s, parts[0], parts[1])
	}
	return s
}

// split splits a string and returns an array
func (t *Transformer) split(value interface{}, params string) interface{} {
	s := fmt.Sprint(value)
	delimiter := params
	if delimiter == "" {
		delimiter = ","
	}
	return strings.Split(s, delimiter)
}

// index gets an element from an array by index
func (t *Transformer) index(value interface{}, params string) interface{} {
	idx, err := strconv.Atoi(params)
	if err != nil {
		return value
	}

	switch v := value.(type) {
	case []interface{}:
		if idx >= 0 && idx < len(v) {
			return v[idx]
		}
	case []string:
		if idx >= 0 && idx < len(v) {
			return v[idx]
		}
	}
	return value
}

// round rounds a number to specified decimal places
// params: number of decimal places (default: 0)
func (t *Transformer) round(value interface{}, params string) interface{} {
	f := t.toFloat(value).(float64)
	places := 0
	if params != "" {
		if p, err := strconv.Atoi(params); err == nil {
			places = p
		}
	}
	multiplier := math.Pow(10, float64(places))
	return math.Round(f*multiplier) / multiplier
}

// floor rounds down to the nearest integer
func (t *Transformer) floor(value interface{}) interface{} {
	f := t.toFloat(value).(float64)
	return math.Floor(f)
}

// ceil rounds up to the nearest integer
func (t *Transformer) ceil(value interface{}) interface{} {
	f := t.toFloat(value).(float64)
	return math.Ceil(f)
}

// abs returns the absolute value
func (t *Transformer) abs(value interface{}) interface{} {
	f := t.toFloat(value).(float64)
	return math.Abs(f)
}

// defaultValue returns the params if value is nil or empty
func (t *Transformer) defaultValue(value interface{}, params string) interface{} {
	if value == nil {
		return params
	}
	if s, ok := value.(string); ok && s == "" {
		return params
	}
	return value
}

// ApplyMultiple applies multiple transformations in sequence
func (t *Transformer) ApplyMultiple(value interface{}, transforms []string) interface{} {
	result := value
	for _, transform := range transforms {
		result = t.Apply(result, transform)
	}
	return result
}

// IsValidTransform checks if a transform function is valid
func IsValidTransform(transformFunc string) bool {
	// Extract the function name (before any colon)
	parts := strings.SplitN(transformFunc, ":", 2)
	funcName := parts[0]

	validFuncs := []string{
		"uppercase", "to_upper", "lowercase", "to_lower", "trim",
		"to_string", "to_int", "to_float", "to_bool",
		"to_cents", "from_cents",
		"parse_date", "parse_timestamp", "format_date",
		"base64_decode", "base64_encode",
		"json_parse", "json_stringify",
		"concat", "prefix", "suffix", "replace", "split", "index",
		"round", "floor", "ceil", "abs",
		"default",
	}

	for _, valid := range validFuncs {
		if funcName == valid {
			return true
		}
	}
	return false
}
