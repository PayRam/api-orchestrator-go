package models

import (
	"encoding/json"
	"time"
)

// RetryConfig defines the retry behavior for HTTP requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 = no retries).
	MaxRetries int `json:"max_retries"`

	// InitialBackoff is the initial wait time before the first retry.
	InitialBackoff time.Duration `json:"initial_backoff"`

	// MaxBackoff is the maximum wait time between retries.
	MaxBackoff time.Duration `json:"max_backoff"`

	// BackoffMultiplier is the multiplier applied to backoff after each retry.
	BackoffMultiplier float64 `json:"backoff_multiplier"`

	// RetryOnStatusCodes specifies which HTTP status codes should trigger a retry.
	// Common values: 429 (rate limit), 500, 502, 503, 504 (server errors).
	RetryOnStatusCodes []int `json:"retry_on_status_codes"`
}

// DefaultRetryConfig returns sensible defaults for retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:         3,
		InitialBackoff:     500 * time.Millisecond,
		MaxBackoff:         30 * time.Second,
		BackoffMultiplier:  2.0,
		RetryOnStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// ExecutorConfig contains configuration for the HTTP executor.
type ExecutorConfig struct {
	// Timeout is the maximum duration for a single request (excluding retries).
	Timeout time.Duration `json:"timeout"`

	// RetryConfig defines retry behavior. If nil, no retries are performed.
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// LogRequestBody enables logging of request body (may contain sensitive data).
	LogRequestBody bool `json:"log_request_body"`

	// LogResponseBody enables logging of response body.
	LogResponseBody bool `json:"log_response_body"`

	// MaxResponseBodyLog is the maximum length of response body to log (0 = unlimited).
	MaxResponseBodyLog int `json:"max_response_body_log"`

	// SensitiveHeaders is a list of header names to mask in logs.
	SensitiveHeaders []string `json:"sensitive_headers"`

	// UserAgent is the User-Agent header to use for requests.
	UserAgent string `json:"user_agent"`
}

// DefaultExecutorConfig returns sensible defaults for executor configuration.
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		Timeout:            30 * time.Second,
		RetryConfig:        nil, // No retries by default
		LogRequestBody:     false,
		LogResponseBody:    true,
		MaxResponseBodyLog: 1024, // 1KB
		SensitiveHeaders: []string{
			"Authorization",
			"X-API-Key",
			"X-Api-Key",
			"X-Secret",
			"X-Signature",
			"Cookie",
			"Set-Cookie",
		},
		UserAgent: "API-Orchestrator-Go/1.0",
	}
}

// ExecutorRequest represents a request to be executed by the HTTP executor.
type ExecutorRequest struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, PATCH).
	Method string `json:"method"`

	// URL is the fully constructed URL to call.
	URL string `json:"url"`

	// Headers contains all HTTP headers for the request.
	Headers map[string]string `json:"headers"`

	// QueryParams contains URL query parameters (appended to URL).
	QueryParams map[string]string `json:"query_params,omitempty"`

	// Body contains the request body (can be nil for GET/DELETE).
	Body json.RawMessage `json:"body,omitempty"`

	// Config is optional per-request configuration (overrides defaults).
	Config *ExecutorConfig `json:"config,omitempty"`

	// RequestID is a unique identifier for tracing/correlation.
	RequestID string `json:"request_id,omitempty"`

	// ProviderName is the name of the provider (for logging/metrics).
	ProviderName string `json:"provider_name,omitempty"`

	// EndpointName is the name of the endpoint/action (for logging/metrics).
	EndpointName string `json:"endpoint_name,omitempty"`
}

// ExecutorResponse represents the response from an HTTP execution.
type ExecutorResponse struct {
	// StatusCode is the HTTP status code returned.
	StatusCode int `json:"status_code"`

	// Status is the HTTP status string (e.g., "200 OK").
	Status string `json:"status"`

	// Headers contains response headers.
	Headers map[string]string `json:"headers"`

	// Body contains the raw response body.
	Body json.RawMessage `json:"body"`

	// RawBody contains the response body as a string (for debugging).
	RawBody string `json:"raw_body"`

	// ContentLength is the size of the response body in bytes.
	ContentLength int64 `json:"content_length"`

	// Duration is the total time taken for the request (including retries).
	Duration time.Duration `json:"duration"`

	// Attempts is the number of attempts made (1 = no retries occurred).
	Attempts int `json:"attempts"`

	// RetryDurations contains the duration of each retry attempt.
	RetryDurations []time.Duration `json:"retry_durations,omitempty"`

	// Error contains any error message if the request failed.
	Error string `json:"error,omitempty"`

	// IsRetryable indicates if the error is retryable.
	IsRetryable bool `json:"is_retryable"`

	// RequestID is echoed from the request for correlation.
	RequestID string `json:"request_id,omitempty"`
}

// IsSuccess returns true if the response indicates success (2xx status code).
func (r *ExecutorResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true if the response indicates a client error (4xx).
func (r *ExecutorResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the response indicates a server error (5xx).
func (r *ExecutorResponse) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// HasError returns true if there was any error during execution.
func (r *ExecutorResponse) HasError() bool {
	return r.Error != "" || r.StatusCode >= 400
}

// BodyAsMap attempts to unmarshal the body as a map.
func (r *ExecutorResponse) BodyAsMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(r.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BodyAsString returns the body as a string.
func (r *ExecutorResponse) BodyAsString() string {
	return r.RawBody
}
