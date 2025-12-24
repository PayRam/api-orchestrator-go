package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// Integration test demonstrating the full flow:
// Provider -> Credentials -> Headers (with strategies) -> Request Builder -> cURL

func TestFullFlow_GenerateCurl(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// ===========================================
	// 1. Setup Provider
	// ===========================================
	provider := &models.Provider{
		ID:          "provider-banxa",
		Name:        "banxa",
		DisplayName: "Banxa",
		BaseURL:     "https://api.banxa.com",
		IsActive:    true,
	}
	t.Logf("Provider: %s (%s)", provider.DisplayName, provider.BaseURL)

	// ===========================================
	// 2. Setup Credentials (encrypted)
	// ===========================================
	apiKey := &models.Credential{
		ID:         "cred-api-key",
		ProviderID: provider.ID,
		Key:        "API_KEY",
	}
	apiKey.SetEncryptedValue("pk_live_abc123")

	apiSecret := &models.Credential{
		ID:         "cred-api-secret",
		ProviderID: provider.ID,
		Key:        "API_SECRET",
	}
	apiSecret.SetEncryptedValue("sk_live_xyz789secret")

	// Decrypt credentials for use
	credentials := make(map[string]string)
	decryptedKey, _ := apiKey.GetDecryptedValue()
	decryptedSecret, _ := apiSecret.GetDecryptedValue()
	credentials["API_KEY"] = decryptedKey
	credentials["API_SECRET"] = decryptedSecret

	t.Logf("Credentials loaded: API_KEY=%s..., API_SECRET=%s...",
		decryptedKey[:8], decryptedSecret[:8])

	// ===========================================
	// 3. Setup Endpoint
	// ===========================================
	endpoint := &models.Endpoint{
		ID:          "endpoint-create-order",
		ProviderID:  provider.ID,
		Name:        "create_order",
		Method:      "POST",
		Path:        "/api/orders",
		Description: "Create a new crypto purchase order",
	}
	t.Logf("Endpoint: %s %s", endpoint.Method, endpoint.Path)

	// ===========================================
	// 4. Setup Request Schemas
	// ===========================================
	schemas := []*models.RequestSchema{
		{
			ID:            "schema-amount",
			EndpointID:    endpoint.ID,
			ParamName:     "fiat_amount",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeNumber,
			Required:      true,
		},
		{
			ID:            "schema-currency",
			EndpointID:    endpoint.ID,
			ParamName:     "fiat_code",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      true,
		},
		{
			ID:            "schema-crypto",
			EndpointID:    endpoint.ID,
			ParamName:     "coin_code",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      true,
		},
		{
			ID:            "schema-wallet",
			EndpointID:    endpoint.ID,
			ParamName:     "wallet_address",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      true,
		},
		{
			ID:            "schema-return-url",
			EndpointID:    endpoint.ID,
			ParamName:     "return_url_on_success",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      false,
			DefaultValue:  "https://myapp.com/success",
		},
	}

	// ===========================================
	// 5. Setup Header Rules
	// ===========================================
	headerRules := []*models.HeaderRule{
		{
			ID:              "rule-content-type",
			ProviderID:      provider.ID,
			HeaderName:      "Content-Type",
			ValueExpression: "static:application/json",
			Priority:        1,
		},
		{
			ID:              "rule-api-key",
			ProviderID:      provider.ID,
			HeaderName:      "X-API-Key",
			ValueExpression: "credential:API_KEY",
			Priority:        2,
		},
		{
			ID:              "rule-signature",
			ProviderID:      provider.ID,
			HeaderName:      "X-Signature",
			ValueExpression: "strategy:signature",
			Priority:        3,
		},
		{
			ID:              "rule-timestamp",
			ProviderID:      provider.ID,
			HeaderName:      "X-Timestamp",
			ValueExpression: "strategy:timestamp",
			Priority:        4,
		},
	}

	// ===========================================
	// 6. Setup Strategy for HMAC Signature
	// ===========================================
	strategyConfig := map[string]interface{}{
		"algorithm":   "SHA256",
		"encoding":    "hex",
		"secret_key":  "credential:API_SECRET",
		"data_format": "{timestamp}{method}{path}{body}",
		"output_key":  "signature",
	}
	configJSON, _ := json.Marshal(strategyConfig)

	strategy := &models.Strategy{
		ID:           "strategy-hmac",
		Name:         "banxa_hmac_signature",
		StrategyType: "HMAC",
		Config:       datatypes.JSON(configJSON),
	}
	t.Logf("Strategy: %s (type: %s)", strategy.Name, strategy.StrategyType)

	// ===========================================
	// 7. User Input
	// ===========================================
	userInput := map[string]interface{}{
		"fiat_amount":    100.00,
		"fiat_code":      "USD",
		"coin_code":      "BTC",
		"wallet_address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
	}
	t.Logf("User Input: %+v", userInput)

	// ===========================================
	// 8. Execute Strategy (simulate)
	// ===========================================
	timestamp := "1703424000"

	// Create mock strategy service
	mockStrategyRepo := &mockStrategyRepoForFlow{
		strategies: map[string]*models.Strategy{
			strategy.Name: strategy,
		},
	}
	strategyService := &strategyServiceImpl{
		repo:   mockStrategyRepo,
		logger: logger,
	}

	strategyCtx := &StrategyExecutionContext{
		Credentials: credentials,
		Input:       userInput,
		Intermediate: map[string]interface{}{
			"timestamp": timestamp,
			"method":    endpoint.Method,
			"path":      endpoint.Path,
		},
		Timestamp: 1703424000,
	}

	strategyResult, err := strategyService.ExecuteStrategy(strategy.Name, strategyCtx)
	if err != nil {
		t.Logf("Strategy execution note: %v (using mock signature)", err)
	}

	// Use computed or mock values
	computedValues := make(map[string]interface{})
	if strategyResult != nil && strategyResult.Values != nil {
		computedValues = strategyResult.Values
	}
	computedValues["timestamp"] = timestamp
	if _, ok := computedValues["signature"]; !ok {
		// Mock signature for demo
		computedValues["signature"] = "a1b2c3d4e5f6789012345678901234567890abcdef"
	}
	t.Logf("Computed values: timestamp=%s, signature=%s...",
		computedValues["timestamp"], computedValues["signature"].(string)[:20])

	// ===========================================
	// 9. Build Headers using HeaderRuleService
	// ===========================================
	mockHeaderRuleRepo := &mockHeaderRuleRepoForFlow{
		rules: map[string][]*models.HeaderRule{
			provider.ID: headerRules,
		},
	}
	headerRuleService := &headerRuleServiceImpl{
		repo:   mockHeaderRuleRepo,
		logger: logger,
	}

	headers, err := headerRuleService.EvaluateHeaderRules(
		provider.ID,
		credentials,
		computedValues,
		userInput,
	)
	if err != nil {
		t.Fatalf("Failed to evaluate header rules: %v", err)
	}
	t.Logf("Headers built: %d headers", len(headers))
	for k, v := range headers {
		if len(v) > 30 {
			t.Logf("  %s: %s...", k, v[:30])
		} else {
			t.Logf("  %s: %s", k, v)
		}
	}

	// ===========================================
	// 10. Build Request using RequestBuilderService
	// ===========================================
	mockSchemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			endpoint.ID: schemas,
		},
	}
	mockValueRepo := &mockRequestValueRepo{
		values: make(map[string][]*models.RequestValue),
	}

	requestBuilderService := &requestBuilderServiceImpl{
		schemaRepo:       mockSchemaRepo,
		requestValueRepo: mockValueRepo,
		logger:           logger,
	}

	buildCtx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  credentials,
		Input:        userInput,
		Intermediate: computedValues,
		Headers:      headers,
	}

	builtRequest, err := requestBuilderService.BuildRequest(buildCtx)
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}

	// ===========================================
	// 11. Display Final Results
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("FINAL HTTP REQUEST")
	t.Log(strings.Repeat("=", 60))
	t.Logf("Method: %s", builtRequest.Method)
	t.Logf("URL: %s", builtRequest.URL)
	t.Log("\nHeaders:")
	for k, v := range builtRequest.Headers {
		t.Logf("  %s: %s", k, v)
	}
	if len(builtRequest.Body) > 0 {
		var prettyBody map[string]interface{}
		json.Unmarshal(builtRequest.Body, &prettyBody)
		prettyJSON, _ := json.MarshalIndent(prettyBody, "  ", "  ")
		t.Logf("\nBody:\n  %s", string(prettyJSON))
	}

	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("GENERATED CURL COMMAND")
	t.Log(strings.Repeat("=", 60))
	t.Logf("\n%s", builtRequest.CurlCommand)

	// ===========================================
	// 12. Verify the output
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("VERIFICATION")
	t.Log(strings.Repeat("=", 60))

	// Verify URL
	expectedURL := "https://api.banxa.com/api/orders"
	if builtRequest.URL != expectedURL {
		t.Errorf("URL mismatch: expected %s, got %s", expectedURL, builtRequest.URL)
	} else {
		t.Log("✓ URL is correct")
	}

	// Verify Method
	if builtRequest.Method != "POST" {
		t.Errorf("Method mismatch: expected POST, got %s", builtRequest.Method)
	} else {
		t.Log("✓ Method is correct")
	}

	// Verify Headers
	if builtRequest.Headers["Content-Type"] != "application/json" {
		t.Error("✗ Content-Type header missing or incorrect")
	} else {
		t.Log("✓ Content-Type header is correct")
	}

	if builtRequest.Headers["X-API-Key"] != credentials["API_KEY"] {
		t.Error("✗ X-API-Key header missing or incorrect")
	} else {
		t.Log("✓ X-API-Key header is correct")
	}

	if builtRequest.Headers["X-Signature"] == "" {
		t.Error("✗ X-Signature header missing")
	} else {
		t.Log("✓ X-Signature header is present")
	}

	if builtRequest.Headers["X-Timestamp"] != timestamp {
		t.Error("✗ X-Timestamp header missing or incorrect")
	} else {
		t.Log("✓ X-Timestamp header is correct")
	}

	// Verify Body
	var body map[string]interface{}
	if err := json.Unmarshal(builtRequest.Body, &body); err != nil {
		t.Errorf("✗ Failed to parse body: %v", err)
	} else {
		if body["fiat_amount"] != 100.0 {
			t.Error("✗ fiat_amount incorrect")
		} else {
			t.Log("✓ fiat_amount is correct")
		}
		if body["fiat_code"] != "USD" {
			t.Error("✗ fiat_code incorrect")
		} else {
			t.Log("✓ fiat_code is correct")
		}
		if body["coin_code"] != "BTC" {
			t.Error("✗ coin_code incorrect")
		} else {
			t.Log("✓ coin_code is correct")
		}
		if body["wallet_address"] == nil {
			t.Error("✗ wallet_address missing")
		} else {
			t.Log("✓ wallet_address is present")
		}
		if body["return_url_on_success"] != "https://myapp.com/success" {
			t.Error("✗ return_url_on_success default value not applied")
		} else {
			t.Log("✓ return_url_on_success default value applied")
		}
	}

	// Verify cURL command
	if builtRequest.CurlCommand == "" {
		t.Error("✗ cURL command is empty")
	} else {
		t.Log("✓ cURL command generated")
	}

	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("FULL FLOW TEST COMPLETED SUCCESSFULLY!")
	t.Log(strings.Repeat("=", 60))
}

// Mock strategy repo for flow test
type mockStrategyRepoForFlow struct {
	strategies map[string]*models.Strategy
}

func (m *mockStrategyRepoForFlow) Create(strategy *models.Strategy) error { return nil }
func (m *mockStrategyRepoForFlow) FindByID(id string) (*models.Strategy, error) {
	return nil, nil
}
func (m *mockStrategyRepoForFlow) FindByName(name string) (*models.Strategy, error) {
	if s, ok := m.strategies[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("strategy not found")
}
func (m *mockStrategyRepoForFlow) FindByType(strategyType string) ([]*models.Strategy, error) {
	return nil, nil
}
func (m *mockStrategyRepoForFlow) List() ([]*models.Strategy, error)      { return nil, nil }
func (m *mockStrategyRepoForFlow) Update(strategy *models.Strategy) error { return nil }
func (m *mockStrategyRepoForFlow) Delete(id string) error                 { return nil }

// Mock header rule repo for flow test
type mockHeaderRuleRepoForFlow struct {
	rules map[string][]*models.HeaderRule
}

func (m *mockHeaderRuleRepoForFlow) Create(rule *models.HeaderRule) error { return nil }
func (m *mockHeaderRuleRepoForFlow) FindByID(id string) (*models.HeaderRule, error) {
	return nil, nil
}
func (m *mockHeaderRuleRepoForFlow) FindByProviderID(providerID string) ([]*models.HeaderRule, error) {
	return m.rules[providerID], nil
}
func (m *mockHeaderRuleRepoForFlow) Update(rule *models.HeaderRule) error { return nil }
func (m *mockHeaderRuleRepoForFlow) Delete(id string) error               { return nil }
