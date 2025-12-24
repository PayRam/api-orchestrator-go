package services

import (
	"errors"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
)

// mockHeaderRuleRepo is a mock implementation of HeaderRuleRepo for testing
type mockHeaderRuleRepo struct {
	rules      map[string]*models.HeaderRule
	byProvider map[string][]*models.HeaderRule
}

func newMockHeaderRuleRepo() *mockHeaderRuleRepo {
	return &mockHeaderRuleRepo{
		rules:      make(map[string]*models.HeaderRule),
		byProvider: make(map[string][]*models.HeaderRule),
	}
}

func (m *mockHeaderRuleRepo) Create(rule *models.HeaderRule) error {
	if _, exists := m.rules[rule.ID]; exists {
		return errors.New("rule already exists")
	}
	m.rules[rule.ID] = rule
	m.byProvider[rule.ProviderID] = append(m.byProvider[rule.ProviderID], rule)
	return nil
}

func (m *mockHeaderRuleRepo) FindByID(id string) (*models.HeaderRule, error) {
	if rule, exists := m.rules[id]; exists {
		return rule, nil
	}
	return nil, errors.New("rule not found")
}

func (m *mockHeaderRuleRepo) FindByProviderID(providerID string) ([]*models.HeaderRule, error) {
	rules := m.byProvider[providerID]
	// Sort by priority (simple bubble sort for testing)
	for i := 0; i < len(rules)-1; i++ {
		for j := 0; j < len(rules)-i-1; j++ {
			if rules[j].Priority > rules[j+1].Priority {
				rules[j], rules[j+1] = rules[j+1], rules[j]
			}
		}
	}
	return rules, nil
}

func (m *mockHeaderRuleRepo) Update(rule *models.HeaderRule) error {
	if _, exists := m.rules[rule.ID]; !exists {
		return errors.New("rule not found")
	}
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockHeaderRuleRepo) Delete(id string) error {
	if rule, exists := m.rules[id]; exists {
		delete(m.rules, id)
		// Remove from byProvider
		rules := m.byProvider[rule.ProviderID]
		for i, r := range rules {
			if r.ID == id {
				m.byProvider[rule.ProviderID] = append(rules[:i], rules[i+1:]...)
				break
			}
		}
		return nil
	}
	return errors.New("rule not found")
}

func TestHeaderRuleService_CreateHeaderRule(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)

	// Create a provider first
	provider := &models.Provider{ID: "provider-1", Name: "test-provider", IsActive: true}
	_ = providerRepo.Create(provider)

	service := NewHeaderRuleService(repo, providerService, logger)

	tests := []struct {
		name    string
		rule    *models.HeaderRule
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid header rule",
			rule: &models.HeaderRule{
				ID:              "rule-1",
				ProviderID:      "provider-1",
				HeaderName:      "Authorization",
				ValueExpression: "credential:api_key",
				Priority:        1,
			},
			wantErr: false,
		},
		{
			name: "missing header name",
			rule: &models.HeaderRule{
				ID:              "rule-2",
				ProviderID:      "provider-1",
				HeaderName:      "",
				ValueExpression: "static:value",
				Priority:        0,
			},
			wantErr: true,
			errMsg:  "header name cannot be empty",
		},
		{
			name: "missing value expression",
			rule: &models.HeaderRule{
				ID:              "rule-3",
				ProviderID:      "provider-1",
				HeaderName:      "Content-Type",
				ValueExpression: "",
				Priority:        0,
			},
			wantErr: true,
			errMsg:  "value expression cannot be empty",
		},
		{
			name: "non-existent provider",
			rule: &models.HeaderRule{
				ID:              "rule-4",
				ProviderID:      "non-existent-provider",
				HeaderName:      "X-API-Key",
				ValueExpression: "static:key",
				Priority:        0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateHeaderRule(tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHeaderRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHeaderRuleService_EvaluateExpression(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)
	service := NewHeaderRuleService(repo, providerService, logger)

	credentials := map[string]string{
		"api_key":    "sk-123456",
		"api_secret": "secret-abc",
	}
	strategyValues := map[string]interface{}{
		"signature": "sig-xyz",
		"nonce":     123456,
	}
	inputParams := map[string]interface{}{
		"user_id": "user-1",
		"amount":  100.50,
	}

	tests := []struct {
		name       string
		expression string
		want       string
		wantErr    bool
	}{
		// Static expressions
		{
			name:       "static value",
			expression: "static:application/json",
			want:       "application/json",
			wantErr:    false,
		},
		{
			name:       "plain value without prefix",
			expression: "plain-value",
			want:       "plain-value",
			wantErr:    false,
		},
		// Credential expressions
		{
			name:       "credential lookup",
			expression: "credential:api_key",
			want:       "sk-123456",
			wantErr:    false,
		},
		{
			name:       "credential not found",
			expression: "credential:non_existent",
			want:       "",
			wantErr:    true,
		},
		// Strategy expressions
		{
			name:       "strategy lookup string",
			expression: "strategy:signature",
			want:       "sig-xyz",
			wantErr:    false,
		},
		{
			name:       "strategy lookup number",
			expression: "strategy:nonce",
			want:       "123456",
			wantErr:    false,
		},
		{
			name:       "strategy not found",
			expression: "strategy:non_existent",
			want:       "",
			wantErr:    true,
		},
		// Param expressions
		{
			name:       "param lookup string",
			expression: "param:user_id",
			want:       "user-1",
			wantErr:    false,
		},
		{
			name:       "param lookup number",
			expression: "param:amount",
			want:       "100.5",
			wantErr:    false,
		},
		{
			name:       "param not found",
			expression: "param:non_existent",
			want:       "",
			wantErr:    true,
		},
		// Template expressions
		{
			name:       "template with credential",
			expression: "template:Bearer ${credential:api_key}",
			want:       "Bearer sk-123456",
			wantErr:    false,
		},
		{
			name:       "template with multiple placeholders",
			expression: "template:${credential:api_key}:${strategy:signature}",
			want:       "sk-123456:sig-xyz",
			wantErr:    false,
		},
		{
			name:       "template with missing placeholder",
			expression: "template:Bearer ${credential:missing}",
			want:       "",
			wantErr:    true,
		},
		// Unknown prefix
		{
			name:       "unknown prefix returns as-is",
			expression: "unknown:value",
			want:       "unknown:value",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.EvaluateExpression(tt.expression, credentials, strategyValues, inputParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EvaluateExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeaderRuleService_EvaluateHeaderRules(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)
	service := NewHeaderRuleService(repo, providerService, logger)

	// Create test rules with different priorities
	rules := []*models.HeaderRule{
		{ID: "rule-1", ProviderID: "provider-1", HeaderName: "Authorization", ValueExpression: "template:Bearer ${credential:api_key}", Priority: 1},
		{ID: "rule-2", ProviderID: "provider-1", HeaderName: "Content-Type", ValueExpression: "static:application/json", Priority: 0},
		{ID: "rule-3", ProviderID: "provider-1", HeaderName: "X-Signature", ValueExpression: "strategy:signature", Priority: 2},
		{ID: "rule-4", ProviderID: "provider-1", HeaderName: "X-User-ID", ValueExpression: "param:user_id", Priority: 3},
	}

	for _, r := range rules {
		_ = repo.Create(r)
	}

	credentials := map[string]string{"api_key": "sk-test-123"}
	strategyValues := map[string]interface{}{"signature": "sig-abc"}
	inputParams := map[string]interface{}{"user_id": "user-42"}

	headers, err := service.EvaluateHeaderRules("provider-1", credentials, strategyValues, inputParams)
	if err != nil {
		t.Fatalf("EvaluateHeaderRules() error = %v", err)
	}

	// Verify all headers were set
	expectedHeaders := map[string]string{
		"Authorization": "Bearer sk-test-123",
		"Content-Type":  "application/json",
		"X-Signature":   "sig-abc",
		"X-User-ID":     "user-42",
	}

	for name, expectedValue := range expectedHeaders {
		if headers[name] != expectedValue {
			t.Errorf("Header %s = %v, want %v", name, headers[name], expectedValue)
		}
	}
}

func TestHeaderRuleService_EvaluateHeaderRules_SkipsFailedExpressions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)
	service := NewHeaderRuleService(repo, providerService, logger)

	// Create rules where some will fail
	rules := []*models.HeaderRule{
		{ID: "rule-1", ProviderID: "provider-1", HeaderName: "Good-Header", ValueExpression: "static:good-value", Priority: 0},
		{ID: "rule-2", ProviderID: "provider-1", HeaderName: "Bad-Header", ValueExpression: "credential:missing", Priority: 1},
		{ID: "rule-3", ProviderID: "provider-1", HeaderName: "Another-Good", ValueExpression: "static:another-value", Priority: 2},
	}

	for _, r := range rules {
		_ = repo.Create(r)
	}

	headers, err := service.EvaluateHeaderRules("provider-1", map[string]string{}, nil, nil)
	if err != nil {
		t.Fatalf("EvaluateHeaderRules() error = %v", err)
	}

	// Should have 2 headers (skipped the failed one)
	if len(headers) != 2 {
		t.Errorf("EvaluateHeaderRules() got %d headers, want 2", len(headers))
	}

	if headers["Good-Header"] != "good-value" {
		t.Errorf("Good-Header = %v, want 'good-value'", headers["Good-Header"])
	}
	if headers["Another-Good"] != "another-value" {
		t.Errorf("Another-Good = %v, want 'another-value'", headers["Another-Good"])
	}
	if _, exists := headers["Bad-Header"]; exists {
		t.Error("Bad-Header should not exist in headers")
	}
}

func TestHeaderRuleService_EvaluateExpression_NilMaps(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)
	service := NewHeaderRuleService(repo, providerService, logger)

	// Test with nil credentials
	_, err := service.EvaluateExpression("credential:key", nil, nil, nil)
	if err == nil {
		t.Error("Expected error when credentials map is nil")
	}

	// Test with nil strategy values
	_, err = service.EvaluateExpression("strategy:key", map[string]string{}, nil, nil)
	if err == nil {
		t.Error("Expected error when strategy values map is nil")
	}

	// Test with nil input params
	_, err = service.EvaluateExpression("param:key", map[string]string{}, map[string]interface{}{}, nil)
	if err == nil {
		t.Error("Expected error when input params map is nil")
	}

	// Static should work with nil maps
	val, err := service.EvaluateExpression("static:test", nil, nil, nil)
	if err != nil {
		t.Errorf("Static expression should work with nil maps: %v", err)
	}
	if val != "test" {
		t.Errorf("Static expression = %v, want 'test'", val)
	}
}

func TestHeaderRuleService_GetHeaderRulesByProviderID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockHeaderRuleRepo()
	providerRepo := newMockProviderRepo()
	providerService := NewProviderService(providerRepo, logger)
	service := NewHeaderRuleService(repo, providerService, logger)

	// Create rules with different providers and priorities
	rules := []*models.HeaderRule{
		{ID: "rule-1", ProviderID: "provider-1", HeaderName: "H1", ValueExpression: "s:1", Priority: 3},
		{ID: "rule-2", ProviderID: "provider-1", HeaderName: "H2", ValueExpression: "s:2", Priority: 1},
		{ID: "rule-3", ProviderID: "provider-2", HeaderName: "H3", ValueExpression: "s:3", Priority: 0},
	}

	for _, r := range rules {
		_ = repo.Create(r)
	}

	// Fetch rules for provider-1
	result, err := service.GetHeaderRulesByProviderID("provider-1")
	if err != nil {
		t.Fatalf("GetHeaderRulesByProviderID() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("GetHeaderRulesByProviderID() got %d rules, want 2", len(result))
	}

	// Verify priority ordering
	if result[0].Priority > result[1].Priority {
		t.Error("Rules should be ordered by priority ASC")
	}
}
