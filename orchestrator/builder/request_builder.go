package builder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/orchestrator/context"
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
type RequestBuilder struct{}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{}
}

// BuildRequest constructs the final request from context and metadata
func (rb *RequestBuilder) BuildRequest(
	ctx *context.OrchestratorContext,
	provider *model.Provider,
	endpoint *model.ProviderEndpoint,
	requestValues []*model.ProviderRequestValue,
) (*FinalRequest, error) {

	// Build URL - use endpoint BaseURL if provided, otherwise construct from endpoint path
	baseURL := endpoint.BaseURL
	if baseURL == "" {
		// If no BaseURL override in endpoint, we need to get it from context or config
		// For now, just use the endpoint path as-is (caller should provide full URL in endpoint)
		baseURL = ""
	}

	fullURL, err := rb.buildURL(baseURL, endpoint.Path, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Build query parameters
	queryParams := rb.buildQueryParams(requestValues, ctx)

	// Build body
	body, err := rb.buildBody(requestValues, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build body: %w", err)
	}

	// Create final request
	request := &FinalRequest{
		Method:      endpoint.Method,
		URL:         fullURL,
		Headers:     ctx.Headers,
		QueryParams: queryParams,
		Body:        body,
	}

	// Generate curl command
	request.CurlCommand = rb.generateCurlCommand(request)

	return request, nil
}

// buildURL constructs the full URL by replacing path parameters
func (rb *RequestBuilder) buildURL(baseURL, path string, ctx *context.OrchestratorContext) (string, error) {
	// Replace path parameters: /api/{id} => /api/123
	finalPath := path
	for key, value := range ctx.InputParams {
		placeholder := "{" + key + "}"
		if strings.Contains(finalPath, placeholder) {
			finalPath = strings.ReplaceAll(finalPath, placeholder, fmt.Sprint(value))
		}
	}

	// Combine base URL and path
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(finalPath, "/"), nil
}

// buildQueryParams builds query parameters from request values
func (rb *RequestBuilder) buildQueryParams(requestValues []*model.ProviderRequestValue, ctx *context.OrchestratorContext) map[string]string {
	params := make(map[string]string)

	for range requestValues {
		// ProviderRequestValue only has ID, SchemaID, and Value
		// The value can be static or a template - for now, use it directly
		// TODO: In full implementation, resolve SchemaID to get parameter name/location
		// For now, skip query param building until we have proper schema resolution
	}

	return params
}

// buildBody builds the request body from request values
func (rb *RequestBuilder) buildBody(requestValues []*model.ProviderRequestValue, ctx *context.OrchestratorContext) (json.RawMessage, error) {
	bodyMap := make(map[string]interface{})

	for range requestValues {
		// ProviderRequestValue only has ID, SchemaID, and Value
		// TODO: In full implementation, resolve SchemaID to get parameter name/location
		// For now, skip body building until we have proper schema resolution
	}

	if len(bodyMap) == 0 {
		return nil, nil
	}

	return json.Marshal(bodyMap)
}

// resolveValue resolves the value - simplified version
// In full implementation, this would resolve based on ProviderRequestSchema
func (rb *RequestBuilder) resolveValue(rv *model.ProviderRequestValue, ctx *context.OrchestratorContext) string {
	// For now, just return the value as-is
	// In full implementation, this would:
	// 1. Look up the ProviderRequestSchema by SchemaID
	// 2. Resolve the value based on the schema's ParamType
	// 3. Apply transformations as needed
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

	// Add body
	if len(req.Body) > 0 {
		parts = append(parts, fmt.Sprintf("-d '%s'", string(req.Body)))
	}

	// Add URL with query params
	finalURL := req.URL
	if len(req.QueryParams) > 0 {
		queryString := url.Values{}
		for key, value := range req.QueryParams {
			queryString.Add(key, value)
		}
		finalURL += "?" + queryString.Encode()
	}
	parts = append(parts, fmt.Sprintf("'%s'", finalURL))

	return strings.Join(parts, " ")
}
