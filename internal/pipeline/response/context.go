package response

import (
	"encoding/json"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// ResponseContext holds the runtime context for response normalization.
// It carries all data needed through the pipeline steps.
type ResponseContext struct {
	// Provider is the provider that generated the response
	Provider models.Provider

	// Action is the action type (create_order, get_order, etc.)
	Action string

	// RawResponse is the original response from the provider as a map
	RawResponse map[string]interface{}

	// RawResponseBytes is the original response as raw JSON bytes
	RawResponseBytes json.RawMessage

	// Mappings are the response mappings for this provider and action
	Mappings []*models.ResponseMapping

	// ExtractedValues holds values extracted via JSONPath
	ExtractedValues map[string]interface{}

	// TransformedValues holds values after transformation
	TransformedValues map[string]interface{}

	// StatusCode is the HTTP status code from the response
	StatusCode int

	// RequestID is the unique identifier for this request
	RequestID string

	// ExecutionTime is how long the original request took
	ExecutionTime time.Duration

	// Error holds any error message
	Error string

	// Metadata holds additional context information
	Metadata map[string]interface{}

	// StartTime is when the context was created
	StartTime time.Time
}

// NewResponseContext creates a new ResponseContext
func NewResponseContext(
	provider models.Provider,
	action string,
	rawResponse map[string]interface{},
	statusCode int,
) *ResponseContext {
	// Convert map to JSON bytes
	rawBytes, _ := json.Marshal(rawResponse)

	return &ResponseContext{
		Provider:          provider,
		Action:            action,
		RawResponse:       rawResponse,
		RawResponseBytes:  rawBytes,
		Mappings:          make([]*models.ResponseMapping, 0),
		ExtractedValues:   make(map[string]interface{}),
		TransformedValues: make(map[string]interface{}),
		StatusCode:        statusCode,
		Metadata:          make(map[string]interface{}),
		StartTime:         time.Now(),
	}
}

// NewResponseContextFromBytes creates a ResponseContext from raw JSON bytes
func NewResponseContextFromBytes(
	provider models.Provider,
	action string,
	rawResponseBytes json.RawMessage,
	statusCode int,
) (*ResponseContext, error) {
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(rawResponseBytes, &rawResponse); err != nil {
		return nil, err
	}

	return &ResponseContext{
		Provider:          provider,
		Action:            action,
		RawResponse:       rawResponse,
		RawResponseBytes:  rawResponseBytes,
		Mappings:          make([]*models.ResponseMapping, 0),
		ExtractedValues:   make(map[string]interface{}),
		TransformedValues: make(map[string]interface{}),
		StatusCode:        statusCode,
		Metadata:          make(map[string]interface{}),
		StartTime:         time.Now(),
	}, nil
}

// SetMappings sets the response mappings for this context
func (ctx *ResponseContext) SetMappings(mappings []*models.ResponseMapping) {
	ctx.Mappings = mappings
}

// SetExtractedValue sets an extracted value for a target field
func (ctx *ResponseContext) SetExtractedValue(targetField string, value interface{}) {
	ctx.ExtractedValues[targetField] = value
}

// GetExtractedValue gets an extracted value by target field
func (ctx *ResponseContext) GetExtractedValue(targetField string) (interface{}, bool) {
	value, ok := ctx.ExtractedValues[targetField]
	return value, ok
}

// SetTransformedValue sets a transformed value for a target field
func (ctx *ResponseContext) SetTransformedValue(targetField string, value interface{}) {
	ctx.TransformedValues[targetField] = value
}

// GetTransformedValue gets a transformed value by target field
func (ctx *ResponseContext) GetTransformedValue(targetField string) (interface{}, bool) {
	value, ok := ctx.TransformedValues[targetField]
	return value, ok
}

// GetFinalValue gets the final value for a target field (transformed if available, otherwise extracted)
func (ctx *ResponseContext) GetFinalValue(targetField string) (interface{}, bool) {
	if value, ok := ctx.TransformedValues[targetField]; ok {
		return value, true
	}
	return ctx.GetExtractedValue(targetField)
}

// SetMetadata sets a metadata value
func (ctx *ResponseContext) SetMetadata(key string, value interface{}) {
	ctx.Metadata[key] = value
}

// GetMetadata gets a metadata value
func (ctx *ResponseContext) GetMetadata(key string) (interface{}, bool) {
	value, ok := ctx.Metadata[key]
	return value, ok
}

// IsSuccess returns true if the response indicates success (2xx status code)
func (ctx *ResponseContext) IsSuccess() bool {
	return ctx.StatusCode >= 200 && ctx.StatusCode < 300
}

// Duration returns the time elapsed since context creation
func (ctx *ResponseContext) Duration() time.Duration {
	return time.Since(ctx.StartTime)
}
