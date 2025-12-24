package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransformer_ToUppercase(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"ALREADY UPPER", "ALREADY UPPER"},
		{"", ""},
		{123, "123"},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "uppercase")
		assert.Equal(t, tt.expected, result)
	}
}

func TestTransformer_ToLowercase(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"already lower", "already lower"},
		{"", ""},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "lowercase")
		assert.Equal(t, tt.expected, result)
	}
}

func TestTransformer_Trim(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{"  hello  ", "hello"},
		{"no trim", "no trim"},
		{"\t\n hello \n\t", "hello"},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "trim")
		assert.Equal(t, tt.expected, result)
	}
}

func TestTransformer_ToString(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{"string", "string"},
		{123, "123"},
		{123.456, "123.456"},
		{int64(1000000000000), "1000000000000"},
		{true, "true"},
		{false, "false"},
		// nil returns nil due to early return in Apply
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "to_string")
		assert.Equal(t, tt.expected, result)
	}

	// Test nil separately - Apply returns nil early
	assert.Nil(t, transformer.Apply(nil, "to_string"))
}

func TestTransformer_ToInt(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{123, 123},
		{123.456, 123},
		{"456", 456},
		{"789.123", 789},
		{int64(1000), 1000},
		{true, 1},
		{false, 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "to_int")
		assert.Equal(t, tt.expected, result)
	}
}

func TestTransformer_ToFloat(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{123.456, 123.456},
		{123, float64(123)},
		{"456.789", 456.789},
		{int64(1000), float64(1000)},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "to_float")
		assert.Equal(t, tt.expected, result)
	}
}

func TestTransformer_ToBool(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"false", false},
		{"TRUE", true},
		{"yes", true},
		{"1", true},
		{"no", false},
		{"0", false},
		{1, true},
		{0, false},
		{1.0, true},
		{0.0, false},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "to_bool")
		assert.Equal(t, tt.expected, result, "input: %v", tt.input)
	}
}

func TestTransformer_ToCents(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{100.50, int64(10050)},
		{99.99, int64(9999)},
		{1.00, int64(100)},
		{"50.25", int64(5025)},
		{0, int64(0)},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "to_cents")
		assert.Equal(t, tt.expected, result, "input: %v", tt.input)
	}
}

func TestTransformer_FromCents(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input    interface{}
		expected interface{}
	}{
		{10050, 100.50},
		{9999, 99.99},
		{100, 1.00},
		{int64(5025), 50.25},
		{"2500", 25.00},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, "from_cents")
		assert.Equal(t, tt.expected, result, "input: %v", tt.input)
	}
}

func TestTransformer_ParseDate(t *testing.T) {
	transformer := NewTransformer()

	t.Run("parses RFC3339 date", func(t *testing.T) {
		result := transformer.Apply("2024-01-15T10:30:00Z", "parse_date")
		parsedTime, ok := result.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, 2024, parsedTime.Year())
		assert.Equal(t, time.January, parsedTime.Month())
		assert.Equal(t, 15, parsedTime.Day())
	})

	t.Run("parses simple date", func(t *testing.T) {
		result := transformer.Apply("2024-01-15", "parse_date")
		parsedTime, ok := result.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, 2024, parsedTime.Year())
	})
}

func TestTransformer_ParseTimestamp(t *testing.T) {
	transformer := NewTransformer()

	t.Run("parses seconds timestamp", func(t *testing.T) {
		result := transformer.Apply(1705315800, "parse_timestamp")
		parsedTime, ok := result.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, 2024, parsedTime.Year())
	})

	t.Run("parses milliseconds timestamp", func(t *testing.T) {
		result := transformer.Apply(1705315800000, "parse_timestamp")
		parsedTime, ok := result.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, 2024, parsedTime.Year())
	})

	t.Run("parses string timestamp", func(t *testing.T) {
		result := transformer.Apply("1705315800", "parse_timestamp")
		parsedTime, ok := result.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, 2024, parsedTime.Year())
	})
}

func TestTransformer_Base64(t *testing.T) {
	transformer := NewTransformer()

	t.Run("encodes to base64", func(t *testing.T) {
		result := transformer.Apply("hello world", "base64_encode")
		assert.Equal(t, "aGVsbG8gd29ybGQ=", result)
	})

	t.Run("decodes from base64", func(t *testing.T) {
		result := transformer.Apply("aGVsbG8gd29ybGQ=", "base64_decode")
		assert.Equal(t, "hello world", result)
	})
}

func TestTransformer_JSON(t *testing.T) {
	transformer := NewTransformer()

	t.Run("parses JSON string", func(t *testing.T) {
		result := transformer.Apply(`{"name":"test","value":123}`, "json_parse")
		m, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test", m["name"])
		assert.Equal(t, float64(123), m["value"])
	})

	t.Run("stringifies to JSON", func(t *testing.T) {
		input := map[string]interface{}{"name": "test", "value": 123}
		result := transformer.Apply(input, "json_stringify")
		assert.Contains(t, result, `"name":"test"`)
		assert.Contains(t, result, `"value":123`)
	})
}

func TestTransformer_Concat(t *testing.T) {
	transformer := NewTransformer()

	t.Run("concat with prefix and suffix", func(t *testing.T) {
		result := transformer.Apply("123", "concat:ORDER-:-DONE")
		assert.Equal(t, "ORDER-123-DONE", result)
	})

	t.Run("concat with prefix only", func(t *testing.T) {
		result := transformer.Apply("456", "concat:ID-")
		assert.Equal(t, "ID-456", result)
	})
}

func TestTransformer_PrefixSuffix(t *testing.T) {
	transformer := NewTransformer()

	t.Run("adds prefix", func(t *testing.T) {
		result := transformer.Apply("123", "prefix:ORDER-")
		assert.Equal(t, "ORDER-123", result)
	})

	t.Run("adds suffix", func(t *testing.T) {
		result := transformer.Apply("123", "suffix:-DONE")
		assert.Equal(t, "123-DONE", result)
	})
}

func TestTransformer_Replace(t *testing.T) {
	transformer := NewTransformer()

	result := transformer.Apply("hello world", "replace:world:universe")
	assert.Equal(t, "hello universe", result)
}

func TestTransformer_Split(t *testing.T) {
	transformer := NewTransformer()

	t.Run("splits by comma", func(t *testing.T) {
		result := transformer.Apply("a,b,c", "split:,")
		arr, ok := result.([]string)
		assert.True(t, ok)
		assert.Equal(t, []string{"a", "b", "c"}, arr)
	})

	t.Run("splits by default comma", func(t *testing.T) {
		result := transformer.Apply("x,y,z", "split")
		arr, ok := result.([]string)
		assert.True(t, ok)
		assert.Equal(t, []string{"x", "y", "z"}, arr)
	})
}

func TestTransformer_Index(t *testing.T) {
	transformer := NewTransformer()

	t.Run("gets element from array", func(t *testing.T) {
		input := []interface{}{"a", "b", "c"}
		result := transformer.Apply(input, "index:1")
		assert.Equal(t, "b", result)
	})

	t.Run("gets element from string array", func(t *testing.T) {
		input := []string{"x", "y", "z"}
		result := transformer.Apply(input, "index:0")
		assert.Equal(t, "x", result)
	})
}

func TestTransformer_Round(t *testing.T) {
	transformer := NewTransformer()

	tests := []struct {
		input     interface{}
		transform string
		expected  interface{}
	}{
		{123.456, "round:2", 123.46},
		{123.454, "round:2", 123.45},
		{123.456, "round:0", 123.0},
		{123.456, "round", 123.0},
	}

	for _, tt := range tests {
		result := transformer.Apply(tt.input, tt.transform)
		assert.Equal(t, tt.expected, result, "input: %v, transform: %s", tt.input, tt.transform)
	}
}

func TestTransformer_FloorCeil(t *testing.T) {
	transformer := NewTransformer()

	t.Run("floor", func(t *testing.T) {
		assert.Equal(t, 123.0, transformer.Apply(123.9, "floor"))
		assert.Equal(t, -124.0, transformer.Apply(-123.1, "floor"))
	})

	t.Run("ceil", func(t *testing.T) {
		assert.Equal(t, 124.0, transformer.Apply(123.1, "ceil"))
		assert.Equal(t, -123.0, transformer.Apply(-123.9, "ceil"))
	})
}

func TestTransformer_Abs(t *testing.T) {
	transformer := NewTransformer()

	assert.Equal(t, 123.0, transformer.Apply(-123, "abs"))
	assert.Equal(t, 456.789, transformer.Apply(-456.789, "abs"))
	assert.Equal(t, 100.0, transformer.Apply(100, "abs"))
}

func TestTransformer_Default(t *testing.T) {
	transformer := NewTransformer()

	// Note: nil input returns nil due to early return in Apply
	// The default transform only applies when value reaches the transform function
	t.Run("nil returns nil due to early return", func(t *testing.T) {
		result := transformer.Apply(nil, "default:unknown")
		assert.Nil(t, result)
	})

	t.Run("uses default for empty string", func(t *testing.T) {
		result := transformer.Apply("", "default:N/A")
		assert.Equal(t, "N/A", result)
	})

	t.Run("keeps value if not empty", func(t *testing.T) {
		result := transformer.Apply("value", "default:N/A")
		assert.Equal(t, "value", result)
	})
}

func TestTransformer_ApplyMultiple(t *testing.T) {
	transformer := NewTransformer()

	result := transformer.ApplyMultiple("  hello world  ", []string{"trim", "uppercase"})
	assert.Equal(t, "HELLO WORLD", result)
}

func TestTransform_GlobalFunction(t *testing.T) {
	// Test the global Transform function
	result := Transform("hello", "uppercase")
	assert.Equal(t, "HELLO", result)
}

func TestIsValidTransform(t *testing.T) {
	validTransforms := []string{
		"uppercase", "lowercase", "trim", "to_string", "to_int", "to_float",
		"to_bool", "to_cents", "from_cents", "parse_date", "parse_timestamp",
		"base64_decode", "base64_encode", "json_parse", "json_stringify",
		"concat:prefix:suffix", "prefix:PRE-", "suffix:-SUF",
		"replace:old:new", "split:,", "index:0", "round:2",
		"floor", "ceil", "abs", "default:N/A",
	}

	for _, tf := range validTransforms {
		assert.True(t, IsValidTransform(tf), "should be valid: %s", tf)
	}

	assert.False(t, IsValidTransform("invalid_transform"))
	assert.False(t, IsValidTransform("unknown"))
}

func TestTransformer_NilInput(t *testing.T) {
	transformer := NewTransformer()

	// Apply returns nil early for nil input
	assert.Nil(t, transformer.Apply(nil, "uppercase"))
	assert.Nil(t, transformer.Apply(nil, "lowercase"))
	assert.Nil(t, transformer.Apply(nil, "trim"))
	// to_string, to_int, to_float will not be reached due to early nil return
	assert.Nil(t, transformer.Apply(nil, "to_string"))
	assert.Nil(t, transformer.Apply(nil, "to_int"))
	assert.Nil(t, transformer.Apply(nil, "to_float"))
}

func TestTransformer_UnknownTransform(t *testing.T) {
	transformer := NewTransformer()

	// Unknown transform should return value unchanged
	result := transformer.Apply("test", "unknown_transform")
	assert.Equal(t, "test", result)
}
