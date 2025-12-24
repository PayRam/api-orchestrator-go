package services

import (
	"encoding/json"

	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// BuiltRequest represents a fully constructed HTTP request ready for execution
type BuiltRequest struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string `json:"method"`

	// URL is the full URL including base URL and path with substituted parameters
	URL string `json:"url"`

	// Headers contains all HTTP headers for the request
	Headers map[string]string `json:"headers"`

	// QueryParams contains URL query parameters
	QueryParams map[string]string `json:"query_params,omitempty"`

	// Body contains the request body as JSON
	Body json.RawMessage `json:"body,omitempty"`

	// CurlCommand is a human-readable cURL representation of the request
	CurlCommand string `json:"curl_command"`

	// Metadata contains additional information about the request
	Metadata *RequestMetadata `json:"metadata,omitempty"`
}

// RequestMetadata contains metadata about the built request
type RequestMetadata struct {
	// ProviderName is the name of the provider
	ProviderName string `json:"provider_name"`

	// EndpointName is the name of the endpoint/action
	EndpointName string `json:"endpoint_name"`

	// PathParams contains the path parameters that were substituted
	PathParams map[string]string `json:"path_params,omitempty"`

	// MissingParams contains required parameters that were not provided
	MissingParams []string `json:"missing_params,omitempty"`

	// ResolvedSchemas contains the schemas that were resolved
	ResolvedSchemas int `json:"resolved_schemas"`
}

// RequestBuildContext provides the context for building a request
type RequestBuildContext struct {
	// Provider is the API provider
	Provider *models.Provider

	// Endpoint is the target endpoint
	Endpoint *models.Endpoint

	// Credentials contains decrypted credential values (key -> value)
	Credentials map[string]string

	// Input contains user-provided input parameters (key -> value)
	Input map[string]interface{}

	// Intermediate contains computed values from strategies (key -> value)
	Intermediate map[string]interface{}

	// Headers contains pre-computed headers from header rules
	Headers map[string]string
}

// RequestBuilderService defines the interface for building complete HTTP requests.
// It integrates with the schema system to construct URLs, query parameters, and request bodies
// with proper parameter substitution.
type RequestBuilderService interface {
	// BuildRequest constructs a complete HTTP request from the build context.
	// It resolves all parameter schemas, substitutes values, and generates headers.
	BuildRequest(ctx *RequestBuildContext) (*BuiltRequest, error)

	// BuildRequestForEndpoint is a convenience method that loads endpoint and provider
	// data automatically and builds the request.
	BuildRequestForEndpoint(
		providerName string,
		endpointName string,
		credentials map[string]string,
		input map[string]interface{},
		headers map[string]string,
	) (*BuiltRequest, error)

	// ValidateRequest validates that all required parameters are provided.
	ValidateRequest(ctx *RequestBuildContext) error

	// GenerateCurl generates a cURL command from a built request.
	GenerateCurl(request *BuiltRequest) string

	// GetSchemasByEndpoint retrieves all parameter schemas for an endpoint.
	GetSchemasByEndpoint(endpointID string) ([]*models.RequestSchema, error)

	// ResolveValue resolves a parameter value from the build context based on source type.
	ResolveValue(
		schema *models.RequestSchema,
		value *models.RequestValue,
		ctx *RequestBuildContext,
	) (interface{}, error)
}
