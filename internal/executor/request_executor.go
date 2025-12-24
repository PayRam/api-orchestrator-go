package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
)

// RequestExecutor defines the interface for executing HTTP requests.
type RequestExecutor interface {
	// Execute sends an HTTP request and returns the response.
	Execute(req *models.ExecutorRequest) (*models.ExecutorResponse, error)

	// ExecuteWithContext executes using the Context struct directly.
	ExecuteWithContext(ctx ExecutionContext) (*models.ExecutorResponse, error)
}

// ExecutionContext provides the context for request execution.
// This allows executing requests directly from the pipeline context.
type ExecutionContext interface {
	// GetMethod returns the HTTP method.
	GetMethod() string

	// GetURL returns the full URL.
	GetURL() string

	// GetHeaders returns the request headers.
	GetHeaders() map[string]string

	// GetBody returns the request body.
	GetBody() interface{}

	// GetRequestID returns a unique request identifier.
	GetRequestID() string

	// GetProviderName returns the provider name for logging.
	GetProviderName() string
}

// DefaultExecutor is the default implementation of RequestExecutor.
// It supports configurable timeout, retry logic with exponential backoff,
// and comprehensive logging with secret masking.
type DefaultExecutor struct {
	client *http.Client
	config *models.ExecutorConfig
	logger *zap.Logger
}

// NewDefaultExecutor creates a new DefaultExecutor with the given configuration.
func NewDefaultExecutor(config *models.ExecutorConfig, logger *zap.Logger) *DefaultExecutor {
	if config == nil {
		config = models.DefaultExecutorConfig()
	}
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &DefaultExecutor{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
		logger: logger,
	}
}

// NewDefaultExecutorWithClient creates a new DefaultExecutor with a custom HTTP client.
func NewDefaultExecutorWithClient(client *http.Client, config *models.ExecutorConfig, logger *zap.Logger) *DefaultExecutor {
	if config == nil {
		config = models.DefaultExecutorConfig()
	}
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &DefaultExecutor{
		client: client,
		config: config,
		logger: logger,
	}
}

// Execute sends an HTTP request and returns the response.
func (e *DefaultExecutor) Execute(req *models.ExecutorRequest) (*models.ExecutorResponse, error) {
	startTime := time.Now()

	// Merge per-request config with defaults
	config := e.mergeConfig(req.Config)

	// Log outbound request
	e.logOutboundRequest(req, config)

	// Build the HTTP request
	httpReq, err := e.buildHTTPRequest(req)
	if err != nil {
		return &models.ExecutorResponse{
			Error:     fmt.Sprintf("failed to build request: %v", err),
			RequestID: req.RequestID,
			Duration:  time.Since(startTime),
			Attempts:  0,
		}, err
	}

	// Execute with retry logic
	var response *models.ExecutorResponse
	var lastErr error
	attempts := 0
	retryDurations := []time.Duration{}

	maxAttempts := 1
	if config.RetryConfig != nil {
		maxAttempts = config.RetryConfig.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		attempts++
		attemptStart := time.Now()

		// Wait before retry (skip for first attempt)
		if attempt > 0 {
			backoff := e.calculateBackoff(attempt, config.RetryConfig)
			e.logger.Debug("Retrying request",
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff),
				zap.String("url", req.URL),
			)
			time.Sleep(backoff)
		}

		// Execute the request
		response, lastErr = e.doRequest(httpReq, req, attemptStart)
		retryDurations = append(retryDurations, time.Since(attemptStart))

		// Check if we should retry
		if lastErr == nil && !e.shouldRetry(response.StatusCode, config.RetryConfig) {
			break
		}

		// Clone the request for retry (body needs to be re-readable)
		if attempt < maxAttempts-1 {
			httpReq, err = e.buildHTTPRequest(req)
			if err != nil {
				break
			}
		}
	}

	// Set final response details
	if response == nil {
		response = &models.ExecutorResponse{
			Error:     lastErr.Error(),
			RequestID: req.RequestID,
		}
	}
	response.Duration = time.Since(startTime)
	response.Attempts = attempts
	response.RetryDurations = retryDurations
	response.RequestID = req.RequestID

	// Log response
	e.logResponse(req, response, config)

	return response, lastErr
}

// ExecuteWithContext executes using the Context struct directly.
func (e *DefaultExecutor) ExecuteWithContext(ctx ExecutionContext) (*models.ExecutorResponse, error) {
	// Convert body to JSON if needed
	var body json.RawMessage
	if ctx.GetBody() != nil {
		switch b := ctx.GetBody().(type) {
		case json.RawMessage:
			body = b
		case []byte:
			body = json.RawMessage(b)
		case string:
			body = json.RawMessage(b)
		default:
			jsonBody, err := json.Marshal(b)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = jsonBody
		}
	}

	req := &models.ExecutorRequest{
		Method:       ctx.GetMethod(),
		URL:          ctx.GetURL(),
		Headers:      ctx.GetHeaders(),
		Body:         body,
		RequestID:    ctx.GetRequestID(),
		ProviderName: ctx.GetProviderName(),
	}

	return e.Execute(req)
}

// buildHTTPRequest creates an http.Request from ExecutorRequest.
func (e *DefaultExecutor) buildHTTPRequest(req *models.ExecutorRequest) (*http.Request, error) {
	// Build URL with query params
	fullURL := req.URL
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(req.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		q := parsedURL.Query()
		for k, v := range req.QueryParams {
			q.Set(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}

	// Create body reader
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set default Content-Type for requests with body
	if len(req.Body) > 0 && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set User-Agent if configured
	if e.config.UserAgent != "" && httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", e.config.UserAgent)
	}

	return httpReq, nil
}

// doRequest executes a single HTTP request.
func (e *DefaultExecutor) doRequest(httpReq *http.Request, req *models.ExecutorRequest, startTime time.Time) (*models.ExecutorResponse, error) {
	resp, err := e.client.Do(httpReq)
	if err != nil {
		isRetryable := e.isNetworkErrorRetryable(err)
		return &models.ExecutorResponse{
			Error:       err.Error(),
			IsRetryable: isRetryable,
			RequestID:   req.RequestID,
			Duration:    time.Since(startTime),
			Attempts:    1,
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &models.ExecutorResponse{
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Error:       fmt.Sprintf("failed to read response body: %v", err),
			IsRetryable: false,
			RequestID:   req.RequestID,
			Duration:    time.Since(startTime),
			Attempts:    1,
		}, err
	}

	// Extract response headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	response := &models.ExecutorResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       headers,
		Body:          json.RawMessage(bodyBytes),
		RawBody:       string(bodyBytes),
		ContentLength: resp.ContentLength,
		Duration:      time.Since(startTime),
		Attempts:      1,
		RequestID:     req.RequestID,
	}

	// Set error for non-2xx responses
	if resp.StatusCode >= 400 {
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		response.IsRetryable = e.shouldRetry(resp.StatusCode, e.config.RetryConfig)
	}

	return response, nil
}

// shouldRetry determines if a request should be retried based on status code.
func (e *DefaultExecutor) shouldRetry(statusCode int, retryConfig *models.RetryConfig) bool {
	if retryConfig == nil {
		return false
	}
	for _, code := range retryConfig.RetryOnStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// calculateBackoff calculates the backoff duration for a retry attempt.
func (e *DefaultExecutor) calculateBackoff(attempt int, retryConfig *models.RetryConfig) time.Duration {
	if retryConfig == nil {
		return 0
	}

	backoff := float64(retryConfig.InitialBackoff) * math.Pow(retryConfig.BackoffMultiplier, float64(attempt-1))
	maxBackoff := float64(retryConfig.MaxBackoff)

	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return time.Duration(backoff)
}

// isNetworkErrorRetryable checks if a network error should trigger a retry.
func (e *DefaultExecutor) isNetworkErrorRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Network errors (timeout, connection refused, etc.) are typically retryable
	errMsg := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"i/o timeout",
		"TLS handshake timeout",
	}
	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(retryableErr)) {
			return true
		}
	}
	return false
}

// mergeConfig merges per-request config with default config.
func (e *DefaultExecutor) mergeConfig(reqConfig *models.ExecutorConfig) *models.ExecutorConfig {
	if reqConfig == nil {
		return e.config
	}

	// Start with defaults
	merged := *e.config

	// Override with per-request config
	if reqConfig.Timeout > 0 {
		merged.Timeout = reqConfig.Timeout
	}
	if reqConfig.RetryConfig != nil {
		merged.RetryConfig = reqConfig.RetryConfig
	}
	if reqConfig.LogRequestBody {
		merged.LogRequestBody = reqConfig.LogRequestBody
	}
	if reqConfig.LogResponseBody {
		merged.LogResponseBody = reqConfig.LogResponseBody
	}
	if reqConfig.MaxResponseBodyLog > 0 {
		merged.MaxResponseBodyLog = reqConfig.MaxResponseBodyLog
	}
	if len(reqConfig.SensitiveHeaders) > 0 {
		merged.SensitiveHeaders = reqConfig.SensitiveHeaders
	}
	if reqConfig.UserAgent != "" {
		merged.UserAgent = reqConfig.UserAgent
	}

	return &merged
}

// logOutboundRequest logs the outbound request with secret masking.
func (e *DefaultExecutor) logOutboundRequest(req *models.ExecutorRequest, config *models.ExecutorConfig) {
	// Mask sensitive headers
	maskedHeaders := e.maskSensitiveHeaders(req.Headers, config.SensitiveHeaders)

	fields := []zap.Field{
		zap.String("method", req.Method),
		zap.String("url", req.URL),
		zap.Any("headers", maskedHeaders),
		zap.String("request_id", req.RequestID),
	}

	if req.ProviderName != "" {
		fields = append(fields, zap.String("provider", req.ProviderName))
	}
	if req.EndpointName != "" {
		fields = append(fields, zap.String("endpoint", req.EndpointName))
	}
	if len(req.QueryParams) > 0 {
		fields = append(fields, zap.Any("query_params", req.QueryParams))
	}
	if config.LogRequestBody && len(req.Body) > 0 {
		fields = append(fields, zap.String("body", string(req.Body)))
	}

	e.logger.Info("Executing HTTP request", fields...)

	// Generate and log cURL command
	curlCmd := e.generateCurlCommand(req, maskedHeaders)
	e.logger.Debug("cURL command", zap.String("curl", curlCmd))
}

// logResponse logs the response with optional body truncation.
func (e *DefaultExecutor) logResponse(req *models.ExecutorRequest, resp *models.ExecutorResponse, config *models.ExecutorConfig) {
	fields := []zap.Field{
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.Duration("duration", resp.Duration),
		zap.Int("attempts", resp.Attempts),
		zap.String("request_id", resp.RequestID),
	}

	if req.ProviderName != "" {
		fields = append(fields, zap.String("provider", req.ProviderName))
	}
	if resp.Error != "" {
		fields = append(fields, zap.String("error", resp.Error))
	}
	if config.LogResponseBody && len(resp.RawBody) > 0 {
		body := resp.RawBody
		if config.MaxResponseBodyLog > 0 && len(body) > config.MaxResponseBodyLog {
			body = body[:config.MaxResponseBodyLog] + "...(truncated)"
		}
		fields = append(fields, zap.String("body", body))
	}

	if resp.HasError() {
		e.logger.Warn("HTTP request completed with error", fields...)
	} else {
		e.logger.Info("HTTP request completed", fields...)
	}
}

// maskSensitiveHeaders replaces values of sensitive headers with "***".
func (e *DefaultExecutor) maskSensitiveHeaders(headers map[string]string, sensitiveHeaders []string) map[string]string {
	if headers == nil {
		return nil
	}

	masked := make(map[string]string, len(headers))
	for k, v := range headers {
		masked[k] = v
	}

	// Track which headers have already been masked to avoid double-masking
	alreadyMasked := make(map[string]bool)

	for _, sensitive := range sensitiveHeaders {
		for k := range masked {
			if strings.EqualFold(k, sensitive) && !alreadyMasked[k] {
				// Show first 4 chars if long enough, otherwise just mask
				if len(masked[k]) > 8 {
					masked[k] = masked[k][:4] + "***"
				} else {
					masked[k] = "***"
				}
				alreadyMasked[k] = true
			}
		}
	}

	return masked
}

// generateCurlCommand generates a cURL command for debugging.
func (e *DefaultExecutor) generateCurlCommand(req *models.ExecutorRequest, maskedHeaders map[string]string) string {
	var parts []string
	parts = append(parts, "curl")
	parts = append(parts, "-X", req.Method)

	for k, v := range maskedHeaders {
		parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", k, v))
	}

	if len(req.Body) > 0 {
		// Escape single quotes in body
		bodyStr := strings.ReplaceAll(string(req.Body), "'", "'\\''")
		parts = append(parts, "-d", fmt.Sprintf("'%s'", bodyStr))
	}

	// Add query params to URL
	fullURL := req.URL
	if len(req.QueryParams) > 0 {
		parsedURL, err := url.Parse(req.URL)
		if err == nil {
			q := parsedURL.Query()
			for k, v := range req.QueryParams {
				q.Set(k, v)
			}
			parsedURL.RawQuery = q.Encode()
			fullURL = parsedURL.String()
		}
	}
	parts = append(parts, fmt.Sprintf("'%s'", fullURL))

	return strings.Join(parts, " ")
}
