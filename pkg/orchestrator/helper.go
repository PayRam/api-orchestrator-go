package orchestrator

import (
	"encoding/json"
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Helper provides simple methods to add providers, endpoints, and all configuration.
// This is a simplified API for setting up your orchestrator.
type Helper struct {
	db *gorm.DB
}

// Helper returns a helper API for easy configuration
func (o *Orchestrator) Helper() *Helper {
	return &Helper{db: o.db}
}

// ===================================
// Provider Methods
// ===================================

// AddProvider creates a new provider
func (h *Helper) AddProvider(id, name, displayName, baseURL string, isActive bool) error {
	provider := &models.Provider{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
		BaseURL:     baseURL,
		IsActive:    isActive,
	}
	return h.db.Create(provider).Error
}

// GetProvider retrieves a provider by ID
func (h *Helper) GetProvider(id string) (*models.Provider, error) {
	var provider models.Provider
	err := h.db.Where("id = ?", id).First(&provider).Error
	return &provider, err
}

// ListProviders returns all active providers
func (h *Helper) ListProviders() ([]*models.Provider, error) {
	var providers []*models.Provider
	err := h.db.Where("is_active = ?", true).Find(&providers).Error
	return providers, err
}

// ===================================
// Credential Methods
// ===================================

// AddCredential creates a new credential with automatic encryption
func (h *Helper) AddCredential(providerID, key, value string, required bool) error {
	cred := &models.Credential{
		ID:         fmt.Sprintf("cred-%s-%s", providerID, key),
		ProviderID: providerID,
		Key:        key,
		Required:   required,
	}

	// Encrypt the value
	if err := cred.SetEncryptedValue(value); err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	return h.db.Create(cred).Error
}

// GetCredentials retrieves all credentials for a provider (decrypted)
func (h *Helper) GetCredentials(providerID string) (map[string]string, error) {
	var creds []*models.Credential
	if err := h.db.Where("provider_id = ?", providerID).Find(&creds).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cred := range creds {
		decrypted, err := cred.GetDecryptedValue()
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential %s: %w", cred.Key, err)
		}
		result[cred.Key] = decrypted
	}

	return result, nil
}

// ===================================
// Endpoint Methods
// ===================================

// AddEndpoint creates a new endpoint
func (h *Helper) AddEndpoint(id, providerID, name, method, path string) error {
	endpoint := &models.Endpoint{
		ID:         id,
		ProviderID: providerID,
		Name:       name,
		Method:     method,
		Path:       path,
	}
	return h.db.Create(endpoint).Error
}

// GetEndpoint retrieves an endpoint by ID
func (h *Helper) GetEndpoint(id string) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	err := h.db.Where("id = ?", id).First(&endpoint).Error
	return &endpoint, err
}

// ListEndpoints returns all endpoints for a provider
func (h *Helper) ListEndpoints(providerID string) ([]*models.Endpoint, error) {
	var endpoints []*models.Endpoint
	err := h.db.Where("provider_id = ?", providerID).Find(&endpoints).Error
	return endpoints, err
}

// ===================================
// Request Parameter Methods
// ===================================

// AddParam adds a request parameter to an endpoint
func (h *Helper) AddParam(endpointID, paramName, paramLocation, paramType string, required bool, defaultValue string) error {
	schema := &models.RequestSchema{
		ID:            fmt.Sprintf("schema-%s-%s", endpointID, paramName),
		EndpointID:    endpointID,
		ParamName:     paramName,
		ParamLocation: paramLocation,
		ParamType:     paramType,
		Required:      required,
		DefaultValue:  defaultValue,
	}
	return h.db.Create(schema).Error
}

// GetParams retrieves all parameters for an endpoint
func (h *Helper) GetParams(endpointID string) ([]*models.RequestSchema, error) {
	var schemas []*models.RequestSchema
	err := h.db.Where("endpoint_id = ?", endpointID).Find(&schemas).Error
	return schemas, err
}

// ===================================
// Header Rule Methods
// ===================================

// AddHeaderRule creates a new header rule
func (h *Helper) AddHeaderRule(providerID, headerName, valueExpression string, priority int) error {
	rule := &models.HeaderRule{
		ID:              fmt.Sprintf("rule-%s-%s", providerID, headerName),
		ProviderID:      providerID,
		HeaderName:      headerName,
		ValueExpression: valueExpression,
		Priority:        priority,
	}
	return h.db.Create(rule).Error
}

// GetHeaderRules retrieves all header rules for a provider
func (h *Helper) GetHeaderRules(providerID string) ([]*models.HeaderRule, error) {
	var rules []*models.HeaderRule
	err := h.db.Where("provider_id = ?", providerID).Order("priority ASC").Find(&rules).Error
	return rules, err
}

// ===================================
// Strategy Methods
// ===================================

// AddStrategy creates a new strategy
func (h *Helper) AddStrategy(id, name, strategyType string, config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy config: %w", err)
	}

	strategy := &models.Strategy{
		ID:           id,
		Name:         name,
		StrategyType: strategyType,
		Config:       datatypes.JSON(configJSON),
	}
	return h.db.Create(strategy).Error
}

// GetStrategy retrieves a strategy by name
func (h *Helper) GetStrategy(name string) (*models.Strategy, error) {
	var strategy models.Strategy
	err := h.db.Where("name = ?", name).First(&strategy).Error
	return &strategy, err
}

// ListStrategies returns all strategies
func (h *Helper) ListStrategies() ([]*models.Strategy, error) {
	var strategies []*models.Strategy
	err := h.db.Find(&strategies).Error
	return strategies, err
}

// ===================================
// Response Mapping Methods
// ===================================

// AddResponseMapping creates a new response mapping
func (h *Helper) AddResponseMapping(providerID, action, targetField, sourceJSONPath string, required bool, transform, defaultValue string, priority int) error {
	mapping := &models.ResponseMapping{
		ID:             fmt.Sprintf("mapping-%s-%s-%s", providerID, action, targetField),
		ProviderID:     providerID,
		Action:         action,
		TargetField:    targetField,
		SourceJSONPath: sourceJSONPath,
		Transform:      transform,
		DefaultValue:   defaultValue,
		IsRequired:     required,
		Priority:       priority,
	}
	return h.db.Create(mapping).Error
}

// GetResponseMappings retrieves all mappings for a provider and action
func (h *Helper) GetResponseMappings(providerID, action string) ([]*models.ResponseMapping, error) {
	var mappings []*models.ResponseMapping
	err := h.db.Where("provider_id = ? AND action = ?", providerID, action).
		Order("priority ASC").
		Find(&mappings).Error
	return mappings, err
}

// ===================================
// Bulk Setup Methods
// ===================================

// SetupProvider is a convenience method to set up a complete provider in one call
func (h *Helper) SetupProvider(cfg ProviderSetup) error {
	// Create provider
	if err := h.AddProvider(cfg.ID, cfg.Name, cfg.DisplayName, cfg.BaseURL, true); err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Add credentials
	for _, cred := range cfg.Credentials {
		if err := h.AddCredential(cfg.ID, cred.Key, cred.Value, cred.Required); err != nil {
			return fmt.Errorf("failed to add credential %s: %w", cred.Key, err)
		}
	}

	// Add endpoints
	for _, ep := range cfg.Endpoints {
		endpointID := fmt.Sprintf("endpoint-%s-%s", cfg.ID, ep.Name)
		if err := h.AddEndpoint(endpointID, cfg.ID, ep.Name, ep.Method, ep.Path); err != nil {
			return fmt.Errorf("failed to add endpoint %s: %w", ep.Name, err)
		}

		// Add parameters
		for _, param := range ep.Params {
			if err := h.AddParam(endpointID, param.Name, param.Location, param.Type, param.Required, param.DefaultValue); err != nil {
				return fmt.Errorf("failed to add param %s: %w", param.Name, err)
			}
		}
	}

	// Add header rules
	for i, rule := range cfg.HeaderRules {
		if err := h.AddHeaderRule(cfg.ID, rule.HeaderName, rule.ValueExpression, i+1); err != nil {
			return fmt.Errorf("failed to add header rule %s: %w", rule.HeaderName, err)
		}
	}

	// Add strategies
	for _, strat := range cfg.Strategies {
		if err := h.AddStrategy(strat.ID, strat.Name, strat.Type, strat.Config); err != nil {
			return fmt.Errorf("failed to add strategy %s: %w", strat.Name, err)
		}
	}

	// Add response mappings
	for _, mapping := range cfg.ResponseMappings {
		if err := h.AddResponseMapping(
			cfg.ID,
			mapping.Action,
			mapping.TargetField,
			mapping.SourceJSONPath,
			mapping.Required,
			mapping.Transform,
			mapping.DefaultValue,
			mapping.Priority,
		); err != nil {
			return fmt.Errorf("failed to add response mapping %s: %w", mapping.TargetField, err)
		}
	}

	return nil
}

// ProviderSetup represents a complete provider configuration
type ProviderSetup struct {
	ID               string
	Name             string
	DisplayName      string
	BaseURL          string
	Credentials      []CredentialSetup
	Endpoints        []EndpointSetup
	HeaderRules      []HeaderRuleSetup
	Strategies       []StrategySetup
	ResponseMappings []ResponseMappingSetup
}

// CredentialSetup represents credential configuration
type CredentialSetup struct {
	Key      string
	Value    string
	Required bool
}

// EndpointSetup represents endpoint configuration
type EndpointSetup struct {
	Name   string
	Method string
	Path   string
	Params []ParamSetup
}

// ParamSetup represents parameter configuration
type ParamSetup struct {
	Name         string
	Location     string // "body", "query", "path", "header"
	Type         string // "string", "number", "integer", "boolean", "object", "array"
	Required     bool
	DefaultValue string
}

// HeaderRuleSetup represents header rule configuration
type HeaderRuleSetup struct {
	HeaderName      string
	ValueExpression string // "static:value", "credential:KEY", "template:...", "strategy:key", "param:name"
}

// StrategySetup represents strategy configuration
type StrategySetup struct {
	ID     string
	Name   string
	Type   string                 // "HMAC", "SIGNATURE", "BASIC_AUTH", etc.
	Config map[string]interface{} // Strategy-specific configuration
}

// ResponseMappingSetup represents response mapping configuration
type ResponseMappingSetup struct {
	Action         string
	TargetField    string
	SourceJSONPath string
	Required       bool
	Transform      string
	DefaultValue   string
	Priority       int
}
