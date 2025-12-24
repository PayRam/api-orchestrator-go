package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

type requestBuilderServiceImpl struct {
	endpointService  EndpointService
	providerService  ProviderService
	schemaRepo       repositories.RequestSchemaRepo
	requestValueRepo repositories.RequestValueRepo
	logger           *zap.Logger
}

// NewRequestBuilderService creates a new request builder service
func NewRequestBuilderService(
	endpointService EndpointService,
	providerService ProviderService,
	schemaRepo repositories.RequestSchemaRepo,
	requestValueRepo repositories.RequestValueRepo,
	logger *zap.Logger,
) RequestBuilderService {
	return &requestBuilderServiceImpl{
		endpointService:  endpointService,
		providerService:  providerService,
		schemaRepo:       schemaRepo,
		requestValueRepo: requestValueRepo,
		logger:           logger,
	}
}

// BuildRequest constructs a complete HTTP request from the build context
func (s *requestBuilderServiceImpl) BuildRequest(ctx *RequestBuildContext) (*BuiltRequest, error) {
	s.logger.Debug("Building request",
		zap.String("provider", ctx.Provider.Name),
		zap.String("endpoint", ctx.Endpoint.Name))

	// Validate required fields
	if ctx.Provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	if ctx.Endpoint == nil {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Get all schemas for this endpoint
	schemas, err := s.schemaRepo.FindByEndpointID(ctx.Endpoint.ID)
	if err != nil {
		s.logger.Error("Failed to fetch schemas", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch schemas: %w", err)
	}

	// Organize schemas by location
	pathSchemas := filterSchemasByLocation(schemas, models.ParamLocationPath)
	querySchemas := filterSchemasByLocation(schemas, models.ParamLocationQuery)
	bodySchemas := filterSchemasByLocation(schemas, models.ParamLocationBody)

	// Build path with substitutions
	finalURL, pathParams, err := s.buildURL(ctx, pathSchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Build query parameters
	queryParams, err := s.buildQueryParams(ctx, querySchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to build query params: %w", err)
	}

	// Build request body
	body, err := s.buildBody(ctx, bodySchemas)
	if err != nil {
		return nil, fmt.Errorf("failed to build body: %w", err)
	}

	// Merge headers from context
	headers := make(map[string]string)
	for k, v := range ctx.Headers {
		headers[k] = v
	}

	// Set Content-Type if body is present
	if len(body) > 0 {
		if _, hasContentType := headers["Content-Type"]; !hasContentType {
			headers["Content-Type"] = "application/json"
		}
	}

	// Build the request object
	request := &BuiltRequest{
		Method:      ctx.Endpoint.Method,
		URL:         finalURL,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        body,
		Metadata: &RequestMetadata{
			ProviderName:    ctx.Provider.Name,
			EndpointName:    ctx.Endpoint.Name,
			PathParams:      pathParams,
			ResolvedSchemas: len(schemas),
		},
	}

	// Generate cURL command
	request.CurlCommand = s.GenerateCurl(request)

	s.logger.Debug("Request built successfully",
		zap.String("url", request.URL),
		zap.String("method", request.Method))

	return request, nil
}

// BuildRequestForEndpoint is a convenience method
func (s *requestBuilderServiceImpl) BuildRequestForEndpoint(
	providerName string,
	endpointName string,
	credentials map[string]string,
	input map[string]interface{},
	headers map[string]string,
) (*BuiltRequest, error) {
	s.logger.Debug("Building request for endpoint",
		zap.String("provider", providerName),
		zap.String("endpoint", endpointName))

	// Fetch provider
	provider, err := s.providerService.GetProviderByName(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Fetch endpoint
	endpoint, err := s.endpointService.GetEndpointByProviderAndName(providerName, endpointName)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Create build context
	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  credentials,
		Input:        input,
		Intermediate: make(map[string]interface{}),
		Headers:      headers,
	}

	return s.BuildRequest(ctx)
}

// ValidateRequest validates that all required parameters are provided
func (s *requestBuilderServiceImpl) ValidateRequest(ctx *RequestBuildContext) error {
	schemas, err := s.schemaRepo.FindRequiredByEndpointID(ctx.Endpoint.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch required schemas: %w", err)
	}

	var missingParams []string
	for _, schema := range schemas {
		_, found := s.resolveParamValue(ctx, schema)
		if !found && schema.DefaultValue == "" {
			missingParams = append(missingParams, schema.ParamName)
		}
	}

	if len(missingParams) > 0 {
		return fmt.Errorf("missing required parameters: %s", strings.Join(missingParams, ", "))
	}

	return nil
}

// GenerateCurl generates a cURL command from a built request
func (s *requestBuilderServiceImpl) GenerateCurl(request *BuiltRequest) string {
	var parts []string
	parts = append(parts, "curl")

	// Add method
	if request.Method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", request.Method))
	}

	// Add headers
	for key, value := range request.Headers {
		// Escape single quotes in header values
		escapedValue := strings.ReplaceAll(value, "'", "'\\''")
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", key, escapedValue))
	}

	// Add body
	if len(request.Body) > 0 {
		// Compact the JSON for cleaner output
		var compacted interface{}
		if err := json.Unmarshal(request.Body, &compacted); err == nil {
			if compactedJSON, err := json.Marshal(compacted); err == nil {
				// Escape single quotes in body
				escapedBody := strings.ReplaceAll(string(compactedJSON), "'", "'\\''")
				parts = append(parts, fmt.Sprintf("-d '%s'", escapedBody))
			}
		} else {
			parts = append(parts, fmt.Sprintf("-d '%s'", string(request.Body)))
		}
	}

	// Build final URL with query params
	finalURL := request.URL
	if len(request.QueryParams) > 0 {
		queryString := url.Values{}
		for key, value := range request.QueryParams {
			queryString.Add(key, value)
		}
		finalURL += "?" + queryString.Encode()
	}
	parts = append(parts, fmt.Sprintf("'%s'", finalURL))

	return strings.Join(parts, " \\\n  ")
}

// GetSchemasByEndpoint retrieves all parameter schemas for an endpoint
func (s *requestBuilderServiceImpl) GetSchemasByEndpoint(endpointID string) ([]*models.RequestSchema, error) {
	return s.schemaRepo.FindByEndpointID(endpointID)
}

// ResolveValue resolves a parameter value from the build context based on source type
func (s *requestBuilderServiceImpl) ResolveValue(
	schema *models.RequestSchema,
	value *models.RequestValue,
	ctx *RequestBuildContext,
) (interface{}, error) {
	if value == nil {
		// Try to get from input directly using schema param name
		if val, ok := ctx.Input[schema.ParamName]; ok {
			return s.convertValue(val, schema.ParamType)
		}
		// Use default value if set
		if schema.DefaultValue != "" {
			return s.convertValue(schema.DefaultValue, schema.ParamType)
		}
		return nil, fmt.Errorf("no value found for parameter: %s", schema.ParamName)
	}

	var rawValue interface{}

	switch value.SourceType {
	case models.SourceTypeStatic:
		// Parse the JSON value
		if err := json.Unmarshal(value.Value, &rawValue); err != nil {
			// If not valid JSON, use as string
			rawValue = string(value.Value)
		}

	case models.SourceTypeInput:
		if val, ok := ctx.Input[value.SourceKey]; ok {
			rawValue = val
		} else {
			return nil, fmt.Errorf("input not found: %s", value.SourceKey)
		}

	case models.SourceTypeCredential:
		if val, ok := ctx.Credentials[value.SourceKey]; ok {
			rawValue = val
		} else {
			return nil, fmt.Errorf("credential not found: %s", value.SourceKey)
		}

	case models.SourceTypeComputed:
		if val, ok := ctx.Intermediate[value.SourceKey]; ok {
			rawValue = val
		} else {
			return nil, fmt.Errorf("computed value not found: %s", value.SourceKey)
		}

	default:
		return nil, fmt.Errorf("unknown source type: %s", value.SourceType)
	}

	return s.convertValue(rawValue, schema.ParamType)
}

// buildURL constructs the full URL with path parameter substitution
func (s *requestBuilderServiceImpl) buildURL(
	ctx *RequestBuildContext,
	pathSchemas []*models.RequestSchema,
) (string, map[string]string, error) {
	// Determine base URL
	baseURL := ctx.Endpoint.BaseURL
	if baseURL == "" {
		baseURL = ctx.Provider.BaseURL
	}

	// Build path with substitutions
	path := ctx.Endpoint.Path
	pathParams := make(map[string]string)

	// Find path parameters using regex {paramName}
	pathParamRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := pathParamRegex.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		paramName := match[1]
		placeholder := match[0]

		// Try to resolve the value
		var value string
		resolved := false

		// First check path schemas
		for _, schema := range pathSchemas {
			if schema.ParamName == paramName {
				val, found := s.resolveParamValue(ctx, schema)
				if found {
					value = fmt.Sprint(val)
					resolved = true
					break
				}
			}
		}

		// Fall back to input directly
		if !resolved {
			if val, ok := ctx.Input[paramName]; ok {
				value = fmt.Sprint(val)
				resolved = true
			}
		}

		if resolved {
			path = strings.ReplaceAll(path, placeholder, url.PathEscape(value))
			pathParams[paramName] = value
		} else {
			s.logger.Warn("Unresolved path parameter", zap.String("param", paramName))
		}
	}

	// Combine base URL and path
	fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	return fullURL, pathParams, nil
}

// buildQueryParams builds query parameters from schemas
func (s *requestBuilderServiceImpl) buildQueryParams(
	ctx *RequestBuildContext,
	querySchemas []*models.RequestSchema,
) (map[string]string, error) {
	params := make(map[string]string)

	for _, schema := range querySchemas {
		val, found := s.resolveParamValue(ctx, schema)
		if found {
			params[schema.ParamName] = fmt.Sprint(val)
		} else if schema.Required && schema.DefaultValue == "" {
			s.logger.Warn("Required query parameter not found",
				zap.String("param", schema.ParamName))
		} else if schema.DefaultValue != "" {
			params[schema.ParamName] = schema.DefaultValue
		}
	}

	return params, nil
}

// buildBody builds the request body from schemas
func (s *requestBuilderServiceImpl) buildBody(
	ctx *RequestBuildContext,
	bodySchemas []*models.RequestSchema,
) (json.RawMessage, error) {
	if len(bodySchemas) == 0 {
		return nil, nil
	}

	bodyMap := make(map[string]interface{})

	for _, schema := range bodySchemas {
		val, found := s.resolveParamValue(ctx, schema)
		if found {
			bodyMap[schema.ParamName] = val
		} else if schema.Required && schema.DefaultValue == "" {
			s.logger.Warn("Required body parameter not found",
				zap.String("param", schema.ParamName))
		} else if schema.DefaultValue != "" {
			converted, err := s.convertValue(schema.DefaultValue, schema.ParamType)
			if err == nil {
				bodyMap[schema.ParamName] = converted
			} else {
				bodyMap[schema.ParamName] = schema.DefaultValue
			}
		}
	}

	if len(bodyMap) == 0 {
		return nil, nil
	}

	return json.Marshal(bodyMap)
}

// resolveParamValue resolves a parameter value from the build context
func (s *requestBuilderServiceImpl) resolveParamValue(
	ctx *RequestBuildContext,
	schema *models.RequestSchema,
) (interface{}, bool) {
	// First try to get from input using param name
	if val, ok := ctx.Input[schema.ParamName]; ok {
		converted, err := s.convertValue(val, schema.ParamType)
		if err == nil {
			return converted, true
		}
		return val, true
	}

	// Try to get associated request value
	values, err := s.requestValueRepo.FindBySchemaID(schema.ID)
	if err != nil || len(values) == 0 {
		// Use default value if available
		if schema.DefaultValue != "" {
			converted, err := s.convertValue(schema.DefaultValue, schema.ParamType)
			if err == nil {
				return converted, true
			}
			return schema.DefaultValue, true
		}
		return nil, false
	}

	// Use the first matching value
	value := values[0]

	switch value.SourceType {
	case models.SourceTypeStatic:
		var rawValue interface{}
		if err := json.Unmarshal(value.Value, &rawValue); err != nil {
			return string(value.Value), true
		}
		return rawValue, true

	case models.SourceTypeInput:
		if val, ok := ctx.Input[value.SourceKey]; ok {
			return val, true
		}

	case models.SourceTypeCredential:
		if val, ok := ctx.Credentials[value.SourceKey]; ok {
			return val, true
		}

	case models.SourceTypeComputed:
		if val, ok := ctx.Intermediate[value.SourceKey]; ok {
			return val, true
		}
	}

	return nil, false
}

// convertValue converts a value to the specified type
func (s *requestBuilderServiceImpl) convertValue(value interface{}, paramType string) (interface{}, error) {
	switch paramType {
	case models.ParamTypeString:
		return fmt.Sprint(value), nil

	case models.ParamTypeInteger:
		switch v := value.(type) {
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			return strconv.Atoi(v)
		default:
			return strconv.Atoi(fmt.Sprint(v))
		}

	case models.ParamTypeNumber:
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return strconv.ParseFloat(fmt.Sprint(v), 64)
		}

	case models.ParamTypeBoolean:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		default:
			return strconv.ParseBool(fmt.Sprint(v))
		}

	case models.ParamTypeObject, models.ParamTypeArray:
		// Return as-is for complex types
		return value, nil

	default:
		return value, nil
	}
}

// filterSchemasByLocation filters schemas by param location
func filterSchemasByLocation(schemas []*models.RequestSchema, location string) []*models.RequestSchema {
	var filtered []*models.RequestSchema
	for _, schema := range schemas {
		if schema.ParamLocation == location {
			filtered = append(filtered, schema)
		}
	}
	return filtered
}
