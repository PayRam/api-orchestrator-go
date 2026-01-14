package builder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/PayRam/api-orchestrator-go/internal/context"
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

// FinalRequest represents the final built HTTP request ready for execution
type FinalRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Body        json.RawMessage   `json:"body,omitempty"`
	CurlCommand string            `json:"curl_command"` // Human-readable curl representation
}

// RequestBuilder is responsible for building the final HTTP request
type RequestBuilder struct {
	schemaRepo       repositories.RequestSchemaRepo
	requestValueRepo repositories.RequestValueRepo
	logger           *zap.Logger
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(
	schemaRepo repositories.RequestSchemaRepo,
	requestValueRepo repositories.RequestValueRepo,
	logger *zap.Logger,
) *RequestBuilder {
	return &RequestBuilder{
		schemaRepo:       schemaRepo,
		requestValueRepo: requestValueRepo,
		logger:           logger,
	}
}

// BuildRequest constructs the final request from context and metadata
func (rb *RequestBuilder) BuildRequest(
	ctx *context.OrchestratorContext,
	provider *models.Provider,
	endpoint *models.Endpoint,
	requestValues []*models.RequestValue,
) (*FinalRequest, error) {

	// Load request schemas for this endpoint
	schemas, err := rb.schemaRepo.FindByEndpointID(endpoint.ID)
	if err != nil {
		rb.logger.Warn("Failed to load request schemas, continuing without validation",
			zap.String("endpoint_id", endpoint.ID),
			zap.Error(err))
		schemas = []*models.RequestSchema{} // Continue with empty schemas
	}

	rb.logger.Debug("Loaded request schemas",
		zap.String("endpoint_id", endpoint.ID),
		zap.Int("schema_count", len(schemas)))

	// Load request values if not provided
	if len(requestValues) == 0 && len(schemas) > 0 {
		// Load request values for all schemas
		for _, schema := range schemas {
			values, err := rb.requestValueRepo.FindBySchemaID(schema.ID)
			if err != nil {
				rb.logger.Debug("No request values found for schema",
					zap.String("schema_id", schema.ID),
					zap.String("param_name", schema.ParamName))
				continue
			}
			requestValues = append(requestValues, values...)
		}
		rb.logger.Debug("Loaded request values",
			zap.Int("value_count", len(requestValues)))
	}

	// Organize parameters by location based on schemas and values
	pathParams, queryParamsInterface, bodyParams, headerParams, err := rb.organizeParametersBySchema(schemas, requestValues, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to organize parameters: %w", err)
	}

	// Convert query params to string map
	queryParams := make(map[string]string)
	for k, v := range queryParamsInterface {
		queryParams[k] = fmt.Sprint(v)
	}

	// Build URL - use endpoint BaseURL if provided, otherwise use provider BaseURL
	baseURL := endpoint.BaseURL
	if baseURL == "" {
		baseURL = provider.BaseURL
	}

	fullURL, err := rb.buildURLWithParams(baseURL, endpoint.Path, pathParams, queryParamsInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Build body from body parameters
	body, err := rb.buildBodyFromParams(bodyParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build body: %w", err)
	}

	// Merge schema-based headers with context headers
	finalHeaders := make(map[string]string)
	for k, v := range ctx.Headers {
		finalHeaders[k] = v
	}
	for k, v := range headerParams {
		finalHeaders[k] = v
	}

	// Create final request
	request := &FinalRequest{
		Method:      endpoint.Method,
		URL:         fullURL,
		Headers:     finalHeaders,
		QueryParams: queryParams,
		Body:        body,
	}

	// Generate curl command
	request.CurlCommand = rb.generateCurlCommand(request)

	rb.logger.Info("Request built with schema validation",
		zap.String("url", fullURL),
		zap.String("method", endpoint.Method),
		zap.Int("path_params", len(pathParams)),
		zap.Int("query_params", len(queryParams)),
		zap.Int("body_params", len(bodyParams)),
		zap.Int("header_params", len(headerParams)))

	return request, nil
}

// organizeParametersBySchema organizes input parameters based on request schemas and values
func (rb *RequestBuilder) organizeParametersBySchema(
	schemas []*models.RequestSchema,
	requestValues []*models.RequestValue,
	ctx *context.OrchestratorContext,
) (pathParams, queryParams, bodyParams map[string]interface{}, headerParams map[string]string, err error) {

	pathParams = make(map[string]interface{})
	queryParams = make(map[string]interface{})
	bodyParams = make(map[string]interface{})
	headerParams = make(map[string]string)

	// Create schema map for quick lookup
	schemaMap := make(map[string]*models.RequestSchema)
	for _, schema := range schemas {
		schemaMap[schema.ParamName] = schema
	}

	// Create request value map: schemaID -> []RequestValue
	valuesBySchema := make(map[string][]*models.RequestValue)
	for _, rv := range requestValues {
		valuesBySchema[rv.SchemaID] = append(valuesBySchema[rv.SchemaID], rv)
	}

	// Process each schema
	for _, schema := range schemas {
		var value interface{}
		var hasValue bool

		// Check if there are request values for this schema
		if values, ok := valuesBySchema[schema.ID]; ok && len(values) > 0 {
			// Use the first request value to resolve the actual value
			rv := values[0]
			value = rb.resolveValue(rv, ctx)
			hasValue = (value != nil)

			if hasValue {
				rb.logger.Debug("Resolved value from RequestValue",
					zap.String("param", schema.ParamName),
					zap.String("source_type", rv.SourceType),
					zap.String("source_key", rv.SourceKey))
			}
		} else {
			// No request value mapping, try direct lookup using schema's param name
			value, hasValue = ctx.Input[schema.ParamName]
		}

		// Handle missing values
		if !hasValue {
			// Parameter not provided
			if schema.Required {
				if schema.DefaultValue != "" {
					// Use default value
					value = schema.DefaultValue
					hasValue = true
					rb.logger.Debug("Using default value for required parameter",
						zap.String("param", schema.ParamName),
						zap.String("default", schema.DefaultValue))
				} else {
					return nil, nil, nil, nil, fmt.Errorf("required parameter missing: %s", schema.ParamName)
				}
			} else if schema.DefaultValue != "" {
				// Optional parameter with default
				value = schema.DefaultValue
				hasValue = true
				rb.logger.Debug("Using default value for optional parameter",
					zap.String("param", schema.ParamName),
					zap.String("default", schema.DefaultValue))
			}
		}

		if !hasValue {
			// Optional parameter without value, skip
			continue
		}

		// Organize by location
		switch schema.ParamLocation {
		case "path":
			pathParams[schema.ParamName] = value
		case "query":
			queryParams[schema.ParamName] = value
		case "body":
			bodyParams[schema.ParamName] = value
		case "header":
			headerParams[schema.ParamName] = fmt.Sprintf("%v", value)
		default:
			rb.logger.Warn("Unknown parameter location, defaulting to body",
				zap.String("param", schema.ParamName),
				zap.String("location", schema.ParamLocation))
			bodyParams[schema.ParamName] = value
		}
	}

	// Handle parameters without schemas (backward compatibility)
	for paramName, value := range ctx.Input {
		if _, hasSchema := schemaMap[paramName]; !hasSchema {
			rb.logger.Debug("Parameter has no schema, defaulting to body",
				zap.String("param", paramName))
			bodyParams[paramName] = value
		}
	}

	return pathParams, queryParams, bodyParams, headerParams, nil
}

// buildURLWithParams constructs the full URL with path and query parameters
func (rb *RequestBuilder) buildURLWithParams(
	baseURL, path string,
	pathParams, queryParams map[string]interface{},
) (string, error) {
	// Replace path parameters
	finalPath := path
	for key, value := range pathParams {
		placeholder := "{" + key + "}"
		if strings.Contains(finalPath, placeholder) {
			finalPath = strings.ReplaceAll(finalPath, placeholder, fmt.Sprint(value))
			rb.logger.Debug("Replaced path parameter",
				zap.String("param", key),
				zap.Any("value", value))
		}
	}

	// Combine base URL and path
	fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(finalPath, "/")

	// Add query parameters
	if len(queryParams) > 0 {
		queryString := url.Values{}
		for key, value := range queryParams {
			queryString.Add(key, fmt.Sprint(value))
		}
		fullURL += "?" + queryString.Encode()
	}

	return fullURL, nil
}

// buildBodyFromParams builds the request body from body parameters
func (rb *RequestBuilder) buildBodyFromParams(bodyParams map[string]interface{}) (json.RawMessage, error) {
	if len(bodyParams) == 0 {
		return nil, nil
	}

	bodyBytes, err := json.Marshal(bodyParams)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	rb.logger.Debug("Built request body",
		zap.Int("param_count", len(bodyParams)),
		zap.Int("body_size", len(bodyBytes)))

	return bodyBytes, nil
}

// resolveValue resolves the value based on source type
func (rb *RequestBuilder) resolveValue(rv *models.RequestValue, ctx *context.OrchestratorContext) interface{} {
	switch rv.SourceType {
	case models.SourceTypeInput:
		if val, ok := ctx.GetInput(rv.SourceKey); ok {
			return val
		}
	case models.SourceTypeCredential:
		if val, ok := ctx.GetCredential(rv.SourceKey); ok {
			return val
		}
	case models.SourceTypeComputed:
		if val := ctx.Get(rv.SourceKey); val != nil {
			return val
		}
	case models.SourceTypeStatic:
		// Return the JSON value as-is
		return rv.Value
	}
	return rv.Value
}

// generateCurlCommand generates a curl command for debugging
func (rb *RequestBuilder) generateCurlCommand(req *FinalRequest) string {
	var parts []string
	parts = append(parts, "curl")
	parts = append(parts, fmt.Sprintf("-X %s", req.Method))

	// Add headers
	for key, value := range req.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", key, value))
	}

	// Add body only for methods that support it (not GET, HEAD, OPTIONS)
	method := strings.ToUpper(req.Method)
	if len(req.Body) > 0 && method != "GET" && method != "HEAD" && method != "OPTIONS" {
		parts = append(parts, fmt.Sprintf("-d '%s'", string(req.Body)))
	}

	// Use the URL as-is (query params are already included in req.URL from buildURLWithParams)
	// Don't append query params again to avoid duplication
	parts = append(parts, fmt.Sprintf("'%s'", req.URL))

	return strings.Join(parts, " ")
}
