package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PayRam/api-orchestrator-go/orchestrator/builder"
	"go.uber.org/zap"
)

// HTTPExecutor is responsible for executing HTTP requests
type HTTPExecutor struct {
	client *http.Client
	logger *zap.Logger
}

// ExecutionResult represents the result of an HTTP execution
type ExecutionResult struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          json.RawMessage   `json:"body"`
	RawBody       string            `json:"raw_body"`
	ExecutionTime time.Duration     `json:"execution_time"`
	Error         string            `json:"error,omitempty"`
}

// NewHTTPExecutor creates a new HTTP executor
func NewHTTPExecutor(logger *zap.Logger) *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Execute executes an HTTP request and returns the result
func (e *HTTPExecutor) Execute(request *builder.FinalRequest) (*ExecutionResult, error) {
	e.logger.Info("Executing HTTP request",
		zap.String("method", request.Method),
		zap.String("url", request.URL))

	startTime := time.Now()

	// Create HTTP request
	var bodyReader io.Reader
	if request.Body != nil {
		bodyReader = bytes.NewReader(request.Body)
	}

	req, err := http.NewRequest(request.Method, request.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	// Add query parameters
	if len(request.QueryParams) > 0 {
		q := req.URL.Query()
		for key, value := range request.QueryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := e.client.Do(req)
	executionTime := time.Since(startTime)

	if err != nil {
		e.logger.Error("HTTP request failed", zap.Error(err))
		return &ExecutionResult{
			ExecutionTime: executionTime,
			Error:         err.Error(),
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	e.logger.Info("HTTP request completed",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("execution_time", executionTime))

	result := &ExecutionResult{
		StatusCode:    resp.StatusCode,
		Headers:       headers,
		Body:          json.RawMessage(bodyBytes),
		RawBody:       string(bodyBytes),
		ExecutionTime: executionTime,
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return result, nil
}
