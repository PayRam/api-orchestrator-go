package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test helper to create a logger
func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// ==================================================
// TestNewDefaultExecutor
// ==================================================

func TestNewDefaultExecutor(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		executor := NewDefaultExecutor(nil, nil)
		assert.NotNil(t, executor)
		assert.NotNil(t, executor.client)
		assert.NotNil(t, executor.config)
		assert.Equal(t, 30*time.Second, executor.config.Timeout)
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &models.ExecutorConfig{
			Timeout:   10 * time.Second,
			UserAgent: "TestAgent/1.0",
		}
		executor := NewDefaultExecutor(config, newTestLogger())
		assert.NotNil(t, executor)
		assert.Equal(t, 10*time.Second, executor.config.Timeout)
		assert.Equal(t, "TestAgent/1.0", executor.config.UserAgent)
	})
}

func TestNewDefaultExecutorWithClient(t *testing.T) {
	t.Run("with custom client", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}
		executor := NewDefaultExecutorWithClient(customClient, nil, newTestLogger())
		assert.NotNil(t, executor)
		assert.Equal(t, customClient, executor.client)
	})
}

// ==================================================
// TestExecute_Success
// ==================================================

func TestExecute_Success(t *testing.T) {
	t.Run("GET request success", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/test", r.URL.Path)
			assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}))
		defer server.Close()

		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method:    http.MethodGet,
			URL:       server.URL + "/api/test",
			Headers:   map[string]string{"X-Custom-Header": "test-value"},
			RequestID: "test-123",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test-123", resp.RequestID)
		assert.True(t, resp.IsSuccess())
		assert.False(t, resp.HasError())
		assert.Contains(t, resp.RawBody, "success")
	})

	t.Run("POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			json.Unmarshal(body, &payload)
			assert.Equal(t, "test", payload["name"])

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"id": "new-123"})
		}))
		defer server.Close()

		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method:       http.MethodPost,
			URL:          server.URL + "/api/create",
			Headers:      map[string]string{"Content-Type": "application/json"},
			Body:         json.RawMessage(`{"name":"test"}`),
			ProviderName: "test-provider",
			EndpointName: "create",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, resp.IsSuccess())

		bodyMap, err := resp.BodyAsMap()
		require.NoError(t, err)
		assert.Equal(t, "new-123", bodyMap["id"])
	})

	t.Run("request with query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
			QueryParams: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// ==================================================
// TestExecute_ErrorResponses
// ==================================================

func TestExecute_ErrorResponses(t *testing.T) {
	t.Run("4xx client error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		}))
		defer server.Close()

		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodPost,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err) // No network error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.True(t, resp.IsClientError())
		assert.True(t, resp.HasError())
		assert.Contains(t, resp.Error, "400")
	})

	t.Run("5xx server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
		}))
		defer server.Close()

		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err) // No network error
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.True(t, resp.IsServerError())
		assert.True(t, resp.HasError())
	})

	t.Run("network error - invalid URL", func(t *testing.T) {
		executor := NewDefaultExecutor(nil, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    "http://invalid.localhost.nonexistent:99999/api/test",
		}

		resp, err := executor.Execute(req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Error)
	})
}

// ==================================================
// TestExecute_Retry
// ==================================================

func TestExecute_RetryLogic(t *testing.T) {
	t.Run("retry on 500 with eventual success", func(t *testing.T) {
		var attemptCount int32 = 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attemptCount, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}))
		defer server.Close()

		config := &models.ExecutorConfig{
			Timeout: 10 * time.Second,
			RetryConfig: &models.RetryConfig{
				MaxRetries:         3,
				InitialBackoff:     10 * time.Millisecond,
				MaxBackoff:         100 * time.Millisecond,
				BackoffMultiplier:  2.0,
				RetryOnStatusCodes: []int{500, 502, 503, 504},
			},
		}
		executor := NewDefaultExecutor(config, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 3, resp.Attempts)
		assert.Len(t, resp.RetryDurations, 3)
	})

	t.Run("retry exhausted - all attempts fail", func(t *testing.T) {
		var attemptCount int32 = 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attemptCount, 1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		config := &models.ExecutorConfig{
			Timeout: 10 * time.Second,
			RetryConfig: &models.RetryConfig{
				MaxRetries:         2,
				InitialBackoff:     10 * time.Millisecond,
				MaxBackoff:         50 * time.Millisecond,
				BackoffMultiplier:  2.0,
				RetryOnStatusCodes: []int{503},
			},
		}
		executor := NewDefaultExecutor(config, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err) // No network error
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, 3, resp.Attempts) // 1 initial + 2 retries
		assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount))
	})

	t.Run("no retry on 400 error", func(t *testing.T) {
		var attemptCount int32 = 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attemptCount, 1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		config := &models.ExecutorConfig{
			Timeout: 10 * time.Second,
			RetryConfig: &models.RetryConfig{
				MaxRetries:         3,
				InitialBackoff:     10 * time.Millisecond,
				RetryOnStatusCodes: []int{500, 502, 503, 504}, // 400 not included
			},
		}
		executor := NewDefaultExecutor(config, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodPost,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, 1, resp.Attempts) // No retries
		assert.Equal(t, int32(1), atomic.LoadInt32(&attemptCount))
	})

	t.Run("retry on rate limit 429", func(t *testing.T) {
		var attemptCount int32 = 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attemptCount, 1)
			if count == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &models.ExecutorConfig{
			Timeout:     10 * time.Second,
			RetryConfig: models.DefaultRetryConfig(),
		}
		config.RetryConfig.InitialBackoff = 10 * time.Millisecond
		config.RetryConfig.MaxBackoff = 50 * time.Millisecond

		executor := NewDefaultExecutor(config, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, resp.Attempts)
	})
}

// ==================================================
// TestExecute_Timeout
// ==================================================

func TestExecute_Timeout(t *testing.T) {
	t.Run("request times out", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &models.ExecutorConfig{
			Timeout: 100 * time.Millisecond,
		}
		executor := NewDefaultExecutor(config, newTestLogger())
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    server.URL + "/api/test",
		}

		resp, err := executor.Execute(req)
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, strings.ToLower(resp.Error), "timeout")
	})
}

// ==================================================
// TestExecute_HeaderMasking
// ==================================================

func TestMaskSensitiveHeaders(t *testing.T) {
	executor := NewDefaultExecutor(nil, newTestLogger())

	t.Run("masks sensitive headers", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer sk_live_1234567890abcdef", // 34 chars > 8, shows "Bear***"
			"X-API-Key":     "pk_test_1234567890abcdef",        // 26 chars > 8, shows "pk_t***"
			"X-Custom":      "public-value",
		}

		masked := executor.maskSensitiveHeaders(headers, executor.config.SensitiveHeaders)

		assert.Equal(t, "application/json", masked["Content-Type"])
		assert.Contains(t, masked["Authorization"], "***")
		assert.Contains(t, masked["X-API-Key"], "***")
		assert.Equal(t, "public-value", masked["X-Custom"])

		// Verify long values show first 4 chars
		assert.True(t, len(masked["Authorization"]) > 3, "Should show partial value for long Authorization")
		assert.True(t, len(masked["X-API-Key"]) > 3, "Should show partial value for long X-API-Key")
	})

	t.Run("handles short header values", func(t *testing.T) {
		headers := map[string]string{
			"X-API-Key": "short",
		}

		masked := executor.maskSensitiveHeaders(headers, executor.config.SensitiveHeaders)
		assert.Equal(t, "***", masked["X-API-Key"])
	})

	t.Run("handles nil headers", func(t *testing.T) {
		masked := executor.maskSensitiveHeaders(nil, executor.config.SensitiveHeaders)
		assert.Nil(t, masked)
	})
}

// ==================================================
// TestExecute_CurlGeneration
// ==================================================

func TestGenerateCurlCommand(t *testing.T) {
	executor := NewDefaultExecutor(nil, newTestLogger())

	t.Run("generates curl for GET request", func(t *testing.T) {
		req := &models.ExecutorRequest{
			Method:  http.MethodGet,
			URL:     "https://api.example.com/test",
			Headers: map[string]string{"Accept": "application/json"},
		}

		curl := executor.generateCurlCommand(req, req.Headers)
		assert.Contains(t, curl, "curl")
		assert.Contains(t, curl, "-X GET")
		assert.Contains(t, curl, "Accept: application/json")
		assert.Contains(t, curl, "https://api.example.com/test")
	})

	t.Run("generates curl for POST with body", func(t *testing.T) {
		req := &models.ExecutorRequest{
			Method:  http.MethodPost,
			URL:     "https://api.example.com/create",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    json.RawMessage(`{"name":"test"}`),
		}

		curl := executor.generateCurlCommand(req, req.Headers)
		assert.Contains(t, curl, "-X POST")
		assert.Contains(t, curl, "-d")
		assert.Contains(t, curl, `{"name":"test"}`)
	})

	t.Run("generates curl with query params", func(t *testing.T) {
		req := &models.ExecutorRequest{
			Method: http.MethodGet,
			URL:    "https://api.example.com/search",
			QueryParams: map[string]string{
				"q":    "test",
				"page": "1",
			},
		}

		curl := executor.generateCurlCommand(req, nil)
		assert.Contains(t, curl, "q=test")
		assert.Contains(t, curl, "page=1")
	})
}

// ==================================================
// TestExecutorResponse_HelperMethods
// ==================================================

func TestExecutorResponse_Helpers(t *testing.T) {
	t.Run("IsSuccess", func(t *testing.T) {
		assert.True(t, (&models.ExecutorResponse{StatusCode: 200}).IsSuccess())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 201}).IsSuccess())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 299}).IsSuccess())
		assert.False(t, (&models.ExecutorResponse{StatusCode: 300}).IsSuccess())
		assert.False(t, (&models.ExecutorResponse{StatusCode: 400}).IsSuccess())
	})

	t.Run("IsClientError", func(t *testing.T) {
		assert.False(t, (&models.ExecutorResponse{StatusCode: 200}).IsClientError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 400}).IsClientError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 404}).IsClientError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 499}).IsClientError())
		assert.False(t, (&models.ExecutorResponse{StatusCode: 500}).IsClientError())
	})

	t.Run("IsServerError", func(t *testing.T) {
		assert.False(t, (&models.ExecutorResponse{StatusCode: 400}).IsServerError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 500}).IsServerError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 503}).IsServerError())
	})

	t.Run("HasError", func(t *testing.T) {
		assert.False(t, (&models.ExecutorResponse{StatusCode: 200}).HasError())
		assert.True(t, (&models.ExecutorResponse{StatusCode: 400}).HasError())
		assert.True(t, (&models.ExecutorResponse{Error: "network error"}).HasError())
	})

	t.Run("BodyAsMap", func(t *testing.T) {
		resp := &models.ExecutorResponse{
			Body: json.RawMessage(`{"key":"value","number":42}`),
		}
		m, err := resp.BodyAsMap()
		require.NoError(t, err)
		assert.Equal(t, "value", m["key"])
		assert.Equal(t, float64(42), m["number"])
	})

	t.Run("BodyAsString", func(t *testing.T) {
		resp := &models.ExecutorResponse{
			RawBody: "test body content",
		}
		assert.Equal(t, "test body content", resp.BodyAsString())
	})
}

// ==================================================
// TestRetryConfig_Defaults
// ==================================================

func TestRetryConfig_Defaults(t *testing.T) {
	config := models.DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 30*time.Second, config.MaxBackoff)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
	assert.Contains(t, config.RetryOnStatusCodes, 429)
	assert.Contains(t, config.RetryOnStatusCodes, 500)
	assert.Contains(t, config.RetryOnStatusCodes, 502)
	assert.Contains(t, config.RetryOnStatusCodes, 503)
	assert.Contains(t, config.RetryOnStatusCodes, 504)
}

// ==================================================
// TestExecutorConfig_Defaults
// ==================================================

func TestExecutorConfig_Defaults(t *testing.T) {
	config := models.DefaultExecutorConfig()
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Nil(t, config.RetryConfig) // No retries by default
	assert.False(t, config.LogRequestBody)
	assert.True(t, config.LogResponseBody)
	assert.Equal(t, 1024, config.MaxResponseBodyLog)
	assert.Contains(t, config.SensitiveHeaders, "Authorization")
	assert.Contains(t, config.SensitiveHeaders, "X-API-Key")
	assert.Equal(t, "API-Orchestrator-Go/1.0", config.UserAgent)
}

// ==================================================
// TestBackoffCalculation
// ==================================================

func TestBackoffCalculation(t *testing.T) {
	executor := NewDefaultExecutor(nil, newTestLogger())

	retryConfig := &models.RetryConfig{
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	t.Run("exponential backoff", func(t *testing.T) {
		// attempt 1 (first retry): 100ms * 2^0 = 100ms
		assert.Equal(t, 100*time.Millisecond, executor.calculateBackoff(1, retryConfig))
		// attempt 2: 100ms * 2^1 = 200ms
		assert.Equal(t, 200*time.Millisecond, executor.calculateBackoff(2, retryConfig))
		// attempt 3: 100ms * 2^2 = 400ms
		assert.Equal(t, 400*time.Millisecond, executor.calculateBackoff(3, retryConfig))
		// attempt 4: 100ms * 2^3 = 800ms
		assert.Equal(t, 800*time.Millisecond, executor.calculateBackoff(4, retryConfig))
	})

	t.Run("respects max backoff", func(t *testing.T) {
		// attempt 5: 100ms * 2^4 = 1600ms, but max is 1000ms
		assert.Equal(t, 1*time.Second, executor.calculateBackoff(5, retryConfig))
	})

	t.Run("nil config returns 0", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), executor.calculateBackoff(1, nil))
	})
}

// ==================================================
// TestNetworkErrorRetryable
// ==================================================

func TestNetworkErrorRetryable(t *testing.T) {
	executor := NewDefaultExecutor(nil, newTestLogger())

	t.Run("timeout is retryable", func(t *testing.T) {
		err := fmt.Errorf("context deadline exceeded (timeout)")
		assert.True(t, executor.isNetworkErrorRetryable(err))
	})

	t.Run("connection refused is retryable", func(t *testing.T) {
		err := fmt.Errorf("dial tcp: connection refused")
		assert.True(t, executor.isNetworkErrorRetryable(err))
	})

	t.Run("connection reset is retryable", func(t *testing.T) {
		err := fmt.Errorf("read: connection reset by peer")
		assert.True(t, executor.isNetworkErrorRetryable(err))
	})

	t.Run("nil error is not retryable", func(t *testing.T) {
		assert.False(t, executor.isNetworkErrorRetryable(nil))
	})

	t.Run("generic error is not retryable", func(t *testing.T) {
		err := fmt.Errorf("invalid JSON payload")
		assert.False(t, executor.isNetworkErrorRetryable(err))
	})
}

// ==================================================
// TestMergeConfig
// ==================================================

func TestMergeConfig(t *testing.T) {
	defaultConfig := &models.ExecutorConfig{
		Timeout:            30 * time.Second,
		UserAgent:          "Default/1.0",
		LogRequestBody:     false,
		LogResponseBody:    true,
		MaxResponseBodyLog: 1024,
	}
	executor := NewDefaultExecutor(defaultConfig, newTestLogger())

	t.Run("nil request config uses defaults", func(t *testing.T) {
		merged := executor.mergeConfig(nil)
		assert.Equal(t, 30*time.Second, merged.Timeout)
		assert.Equal(t, "Default/1.0", merged.UserAgent)
	})

	t.Run("request config overrides defaults", func(t *testing.T) {
		reqConfig := &models.ExecutorConfig{
			Timeout:   10 * time.Second,
			UserAgent: "Custom/2.0",
		}
		merged := executor.mergeConfig(reqConfig)
		assert.Equal(t, 10*time.Second, merged.Timeout)
		assert.Equal(t, "Custom/2.0", merged.UserAgent)
	})

	t.Run("partial override preserves other defaults", func(t *testing.T) {
		reqConfig := &models.ExecutorConfig{
			Timeout: 5 * time.Second,
		}
		merged := executor.mergeConfig(reqConfig)
		assert.Equal(t, 5*time.Second, merged.Timeout)
		assert.Equal(t, "Default/1.0", merged.UserAgent)
		assert.True(t, merged.LogResponseBody)
	})
}
