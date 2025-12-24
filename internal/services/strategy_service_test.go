package services

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// mockStrategyRepo is a mock implementation of StrategyRepo for testing
type mockStrategyRepo struct {
	strategies map[string]*models.Strategy
	nameIndex  map[string]string
}

func newMockStrategyRepo() *mockStrategyRepo {
	return &mockStrategyRepo{
		strategies: make(map[string]*models.Strategy),
		nameIndex:  make(map[string]string),
	}
}

func (m *mockStrategyRepo) Create(strategy *models.Strategy) error {
	if _, exists := m.strategies[strategy.ID]; exists {
		return errors.New("strategy already exists")
	}
	m.strategies[strategy.ID] = strategy
	m.nameIndex[strategy.Name] = strategy.ID
	return nil
}

func (m *mockStrategyRepo) FindByID(id string) (*models.Strategy, error) {
	if strategy, exists := m.strategies[id]; exists {
		return strategy, nil
	}
	return nil, errors.New("strategy not found")
}

func (m *mockStrategyRepo) FindByName(name string) (*models.Strategy, error) {
	if id, exists := m.nameIndex[name]; exists {
		return m.strategies[id], nil
	}
	return nil, errors.New("strategy not found")
}

func (m *mockStrategyRepo) List() ([]*models.Strategy, error) {
	result := make([]*models.Strategy, 0, len(m.strategies))
	for _, s := range m.strategies {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockStrategyRepo) FindByType(strategyType string) ([]*models.Strategy, error) {
	result := make([]*models.Strategy, 0)
	for _, s := range m.strategies {
		if s.StrategyType == strategyType {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockStrategyRepo) Update(strategy *models.Strategy) error {
	if _, exists := m.strategies[strategy.ID]; !exists {
		return errors.New("strategy not found")
	}
	// Update name index if name changed
	for name, id := range m.nameIndex {
		if id == strategy.ID && name != strategy.Name {
			delete(m.nameIndex, name)
			break
		}
	}
	m.strategies[strategy.ID] = strategy
	m.nameIndex[strategy.Name] = strategy.ID
	return nil
}

func (m *mockStrategyRepo) Delete(id string) error {
	if strategy, exists := m.strategies[id]; exists {
		delete(m.nameIndex, strategy.Name)
		delete(m.strategies, id)
		return nil
	}
	return errors.New("strategy not found")
}

func TestStrategyService_CreateStrategy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	tests := []struct {
		name     string
		strategy *models.Strategy
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid strategy",
			strategy: &models.Strategy{
				ID:           "s1",
				Name:         "test_hmac",
				StrategyType: "HMAC",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			strategy: &models.Strategy{
				ID:           "s2",
				Name:         "",
				StrategyType: "HMAC",
			},
			wantErr: true,
			errMsg:  "strategy name cannot be empty",
		},
		{
			name: "missing type",
			strategy: &models.Strategy{
				ID:           "s3",
				Name:         "test_strategy",
				StrategyType: "",
			},
			wantErr: true,
			errMsg:  "strategy type cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateStrategy(tt.strategy)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateStrategy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStrategyService_ExecuteHMAC(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create HMAC strategy
	config := map[string]interface{}{
		"algorithm":        "SHA256",
		"secret_key_ref":   "api_secret",
		"encoding":         "hex",
		"output_key":       "signature",
		"message_template": "${input:method}${input:path}${timestamp}",
		"header_name":      "X-Signature",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "hmac-1",
		Name:         "test_hmac_strategy",
		StrategyType: "HMAC",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	// Create execution context
	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_key":    "test-key",
			"api_secret": "test-secret-123",
		},
		Input: map[string]interface{}{
			"method": "POST",
			"path":   "/api/orders",
		},
		Intermediate: make(map[string]interface{}),
		Timestamp:    1703419200,
		Nonce:        "abc123",
	}

	result, err := service.ExecuteStrategy("test_hmac_strategy", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategy() error = %v", err)
	}

	// Check that signature was generated
	if result.Values["signature"] == nil {
		t.Error("Expected signature in result values")
	}

	// Check that header was set
	if result.Headers["X-Signature"] == "" {
		t.Error("Expected X-Signature header to be set")
	}

	t.Logf("Generated signature: %v", result.Values["signature"])
	t.Logf("Header: %v", result.Headers["X-Signature"])
}

func TestStrategyService_ExecuteBasicAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create Basic Auth strategy
	config := map[string]interface{}{
		"username_ref": "api_key",
		"password_ref": "api_secret",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "basic-1",
		Name:         "test_basic_auth",
		StrategyType: "BASIC_AUTH",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_key":    "username123",
			"api_secret": "password456",
		},
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	result, err := service.ExecuteStrategy("test_basic_auth", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategy() error = %v", err)
	}

	// Check Authorization header
	authHeader := result.Headers["Authorization"]
	if authHeader == "" {
		t.Error("Expected Authorization header to be set")
	}

	// Basic auth should start with "Basic "
	if len(authHeader) < 7 || authHeader[:6] != "Basic " {
		t.Errorf("Authorization header should start with 'Basic ', got: %s", authHeader)
	}

	t.Logf("Authorization header: %s", authHeader)
}

func TestStrategyService_ExecuteAPIKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create API Key strategy
	config := map[string]interface{}{
		"key_ref":       "api_key",
		"header_name":   "X-API-Key",
		"header_prefix": "",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "apikey-1",
		Name:         "test_api_key",
		StrategyType: "API_KEY",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_key": "sk-test-12345",
		},
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	result, err := service.ExecuteStrategy("test_api_key", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategy() error = %v", err)
	}

	// Check X-API-Key header
	if result.Headers["X-API-Key"] != "sk-test-12345" {
		t.Errorf("Expected X-API-Key header = 'sk-test-12345', got: %s", result.Headers["X-API-Key"])
	}
}

func TestStrategyService_ExecutePayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create Payload strategy
	config := map[string]interface{}{
		"mappings": map[string]interface{}{
			"userId":    "input:user_id",
			"amount":    "input:amount",
			"apiKey":    "credential:api_key",
			"timestamp": "timestamp:",
		},
		"include_fields": []interface{}{"currency"},
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "payload-1",
		Name:         "test_payload",
		StrategyType: "PAYLOAD",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_key": "key-123",
		},
		Input: map[string]interface{}{
			"user_id":  "user-456",
			"amount":   100.50,
			"currency": "USD",
		},
		Intermediate: make(map[string]interface{}),
		Timestamp:    1703419200,
	}

	result, err := service.ExecuteStrategy("test_payload", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategy() error = %v", err)
	}

	payload, ok := result.Values["payload"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected payload in result values")
	}

	// Check mapped fields
	if payload["userId"] != "user-456" {
		t.Errorf("Expected userId = 'user-456', got: %v", payload["userId"])
	}
	if payload["apiKey"] != "key-123" {
		t.Errorf("Expected apiKey = 'key-123', got: %v", payload["apiKey"])
	}

	// Check included field
	if payload["currency"] != "USD" {
		t.Errorf("Expected currency = 'USD', got: %v", payload["currency"])
	}
}

func TestStrategyService_ExecuteStrategy_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	ctx := &StrategyExecutionContext{
		Credentials:  make(map[string]string),
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	_, err := service.ExecuteStrategy("non_existent_strategy", ctx)
	if err == nil {
		t.Error("Expected error when strategy not found")
	}
}

func TestStrategyService_ExecuteStrategy_UnsupportedType(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	strategy := &models.Strategy{
		ID:           "unsupported-1",
		Name:         "unsupported_strategy",
		StrategyType: "UNSUPPORTED_TYPE",
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials:  make(map[string]string),
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	_, err := service.ExecuteStrategy("unsupported_strategy", ctx)
	if err == nil {
		t.Error("Expected error for unsupported strategy type")
	}
}

func TestStrategyService_ExecuteStrategy_MissingCredential(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create HMAC strategy that requires api_secret
	config := map[string]interface{}{
		"secret_key_ref": "api_secret",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "hmac-missing",
		Name:         "hmac_missing_cred",
		StrategyType: "HMAC",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials:  make(map[string]string), // No credentials
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	_, err := service.ExecuteStrategy("hmac_missing_cred", ctx)
	if err == nil {
		t.Error("Expected error when required credential is missing")
	}
}

func TestStrategyService_GetStrategiesByType(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create strategies of different types
	strategies := []*models.Strategy{
		{ID: "s1", Name: "hmac_1", StrategyType: "HMAC"},
		{ID: "s2", Name: "hmac_2", StrategyType: "HMAC"},
		{ID: "s3", Name: "basic_1", StrategyType: "BASIC_AUTH"},
	}

	for _, s := range strategies {
		_ = repo.Create(s)
	}

	result, err := service.GetStrategiesByType("HMAC")
	if err != nil {
		t.Fatalf("GetStrategiesByType() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetStrategiesByType() got %d strategies, want 2", len(result))
	}
}

func TestStrategyService_ExecuteSignature(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create Signature strategy with base64 encoding
	config := map[string]interface{}{
		"algorithm":        "SHA256",
		"secret_key_ref":   "api_secret",
		"encoding":         "base64",
		"output_key":       "sig",
		"message_template": "${input:data}",
		"header_name":      "X-Signature",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "sig-1",
		Name:         "test_signature",
		StrategyType: "SIGNATURE",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_secret": "secret-key-456",
		},
		Input: map[string]interface{}{
			"data": "test-payload-data",
		},
		Intermediate: make(map[string]interface{}),
	}

	result, err := service.ExecuteStrategy("test_signature", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategy() error = %v", err)
	}

	// Check that signature was generated
	sig := result.Values["sig"]
	if sig == nil || sig == "" {
		t.Error("Expected signature in result values")
	}

	// Base64 encoded signature should contain only valid characters
	sigStr, ok := sig.(string)
	if !ok {
		t.Error("Expected signature to be a string")
	}

	t.Logf("Generated base64 signature: %s", sigStr)
}

func TestStrategyService_ExecuteStrategyByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockStrategyRepo()
	service := NewStrategyService(repo, logger)

	// Create API Key strategy
	config := map[string]interface{}{
		"key_ref":     "api_key",
		"header_name": "Authorization",
	}
	configJSON, _ := json.Marshal(config)

	strategy := &models.Strategy{
		ID:           "id-test-123",
		Name:         "by_id_strategy",
		StrategyType: "API_KEY",
		Config:       datatypes.JSON(configJSON),
	}
	_ = repo.Create(strategy)

	ctx := &StrategyExecutionContext{
		Credentials: map[string]string{
			"api_key": "key-by-id",
		},
		Input:        make(map[string]interface{}),
		Intermediate: make(map[string]interface{}),
	}

	result, err := service.ExecuteStrategyByID("id-test-123", ctx)
	if err != nil {
		t.Fatalf("ExecuteStrategyByID() error = %v", err)
	}

	if result.Headers["Authorization"] != "key-by-id" {
		t.Errorf("Expected Authorization header = 'key-by-id', got: %s", result.Headers["Authorization"])
	}
}
