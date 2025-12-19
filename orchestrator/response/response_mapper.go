package response

import (
	"encoding/json"
	"time"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/orchestrator/executor"
	"github.com/PayRam/api-orchestrator-go/utils"
	"go.uber.org/zap"
)

// UnifiedResponse represents the normalized response from API calls
type UnifiedResponse struct {
	Success       bool                   `json:"success"`
	StatusCode    int                    `json:"status_code"`
	Data          map[string]interface{} `json:"data"`         // Normalized response data
	RawResponse   json.RawMessage        `json:"raw_response"` // Original provider response
	Error         string                 `json:"error,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	ProviderName  string                 `json:"provider_name"`
	Action        string                 `json:"action"`
	RequestID     string                 `json:"request_id"`
}

// ResponseMapper is responsible for normalizing provider responses
type ResponseMapper struct {
	logger *zap.Logger
}

// NewResponseMapper creates a new response mapper
func NewResponseMapper(logger *zap.Logger) *ResponseMapper {
	return &ResponseMapper{
		logger: logger,
	}
}

// MapResponse normalizes the response based on response mappings
func (rm *ResponseMapper) MapResponse(
	result *executor.ExecutionResult,
	mappings []*model.ProviderResponseMapping,
	providerName string,
	action string,
	requestID string,
) (*UnifiedResponse, error) {

	rm.logger.Debug("Mapping response", zap.Int("mappings_count", len(mappings)))

	// Determine success
	success := result.StatusCode >= 200 && result.StatusCode < 300

	// Create unified response
	unifiedResp := &UnifiedResponse{
		Success:       success,
		StatusCode:    result.StatusCode,
		Data:          make(map[string]interface{}),
		RawResponse:   result.Body,
		Error:         result.Error,
		ExecutionTime: result.ExecutionTime,
		ProviderName:  providerName,
		Action:        action,
		RequestID:     requestID,
	}

	// If no mappings, return raw response
	if len(mappings) == 0 {
		rm.logger.Debug("No mappings defined, returning raw response")
		return unifiedResp, nil
	}

	// Parse raw response
	var rawData map[string]interface{}
	if err := json.Unmarshal(result.Body, &rawData); err != nil {
		rm.logger.Warn("Failed to parse response JSON", zap.Error(err))
		return unifiedResp, nil
	}

	// Apply mappings
	for _, mapping := range mappings {
		value, err := rm.extractValue(rawData, mapping.SourceJSONPath)
		if err != nil {
			rm.logger.Debug("Field not found",
				zap.String("field", mapping.SourceJSONPath))
			continue
		}

		// Apply transformation if specified
		if mapping.Transform != "" {
			value = rm.applyTransformation(value, mapping.Transform)
		}

		unifiedResp.Data[mapping.TargetField] = value
	}

	return unifiedResp, nil
}

// extractValue extracts a value from a map using a field path or JSONPath
func (rm *ResponseMapper) extractValue(data map[string]interface{}, path string) (interface{}, error) {
	// For now, use simple JSONPath with gjson
	return utils.ExtractJSONPath(data, path)
}

// applyTransformation applies a transformation function to a value
func (rm *ResponseMapper) applyTransformation(value interface{}, transformFunc string) interface{} {
	// TODO: Implement transformation functions
	// For now, return as-is
	rm.logger.Debug("Applying transformation", zap.String("func", transformFunc))
	return value
}
