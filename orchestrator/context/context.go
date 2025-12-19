package context

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OrchestratorContext represents the runtime context for orchestration
type OrchestratorContext struct {
	RequestID      uuid.UUID              `json:"request_id"`
	ProviderName   string                 `json:"provider_name"`
	Action         string                 `json:"action"`
	InputParams    map[string]interface{} `json:"input_params"`
	Credentials    map[string]string      `json:"credentials"`
	Headers        map[string]string      `json:"headers"`
	ComputedValues map[string]interface{} `json:"computed_values"` // Values computed by strategies
	Metadata       json.RawMessage        `json:"metadata,omitempty"`
	StartTime      time.Time              `json:"start_time"`
}

// NewContext creates a new orchestration context
func NewContext(providerName, action string, params map[string]interface{}) *OrchestratorContext {
	return &OrchestratorContext{
		RequestID:      uuid.New(),
		ProviderName:   providerName,
		Action:         action,
		InputParams:    params,
		Credentials:    make(map[string]string),
		Headers:        make(map[string]string),
		ComputedValues: make(map[string]interface{}),
		StartTime:      time.Now(),
	}
}

// SetCredential adds a credential to the context
func (ctx *OrchestratorContext) SetCredential(key, value string) {
	ctx.Credentials[key] = value
}

// GetCredential retrieves a credential from the context
func (ctx *OrchestratorContext) GetCredential(key string) (string, bool) {
	val, ok := ctx.Credentials[key]
	return val, ok
}

// SetHeader adds a header to the context
func (ctx *OrchestratorContext) SetHeader(key, value string) {
	ctx.Headers[key] = value
}

// SetComputedValue adds a computed value to the context
func (ctx *OrchestratorContext) SetComputedValue(key string, value interface{}) {
	ctx.ComputedValues[key] = value
}

// GetComputedValue retrieves a computed value from the context
func (ctx *OrchestratorContext) GetComputedValue(key string) (interface{}, bool) {
	val, ok := ctx.ComputedValues[key]
	return val, ok
}
