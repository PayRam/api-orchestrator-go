package orchestrator

import (
	"encoding/json"
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"github.com/PayRam/api-orchestrator-go/internal/services"
	"gorm.io/datatypes"
)

// AdminAPI provides public methods for managing providers, credentials, endpoints, and configurations.
// This is the primary interface for external Go projects to interact with the orchestrator library.
type AdminAPI struct {
	providerService      services.ProviderService
	credentialService    services.CredentialService
	credentialRepo       repositories.CredentialRepo
	endpointService      services.EndpointService
	strategyService      services.StrategyService
	requestSchemaService services.RequestSchemaService
	requestValueService  services.RequestValueService
	headerRuleRepo       repositories.HeaderRuleRepo
	requestSchemaRepo    repositories.RequestSchemaRepo
	requestValueRepo     repositories.RequestValueRepo
	responseMappingRepo  repositories.ResponseMappingRepo
}

// ===============================
// Provider Configuration
// ===============================

// ProviderConfig represents the configuration for a provider
type ProviderConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	BaseURL     string `json:"base_url"`
	IsActive    bool   `json:"is_active"`
}

// CreateProvider creates a new provider using the provider service
func (a *AdminAPI) CreateProvider(cfg ProviderConfig) error {
	provider := &models.Provider{
		ID:          cfg.ID,
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		BaseURL:     cfg.BaseURL,
		IsActive:    cfg.IsActive,
	}
	return a.providerService.CreateProvider(provider)
}

// GetProvider retrieves a provider by ID
func (a *AdminAPI) GetProvider(id string) (*ProviderConfig, error) {
	provider, err := a.providerService.GetProviderByID(id)
	if err != nil {
		return nil, err
	}
	return &ProviderConfig{
		ID:          provider.ID,
		Name:        provider.Name,
		DisplayName: provider.DisplayName,
		BaseURL:     provider.BaseURL,
		IsActive:    provider.IsActive,
	}, nil
}

// GetProviderByName retrieves a provider by name
func (a *AdminAPI) GetProviderByName(name string) (*ProviderConfig, error) {
	provider, err := a.providerService.GetProviderByName(name)
	if err != nil {
		return nil, err
	}
	return &ProviderConfig{
		ID:          provider.ID,
		Name:        provider.Name,
		DisplayName: provider.DisplayName,
		BaseURL:     provider.BaseURL,
		IsActive:    provider.IsActive,
	}, nil
}

// ListProviders returns all providers
func (a *AdminAPI) ListProviders() ([]ProviderConfig, error) {
	providers, err := a.providerService.ListProviders()
	if err != nil {
		return nil, err
	}

	configs := make([]ProviderConfig, len(providers))
	for i, p := range providers {
		configs[i] = ProviderConfig{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
			BaseURL:     p.BaseURL,
			IsActive:    p.IsActive,
		}
	}
	return configs, nil
}

// ListActiveProviders returns only active providers
func (a *AdminAPI) ListActiveProviders() ([]ProviderConfig, error) {
	providers, err := a.providerService.GetAllActive()
	if err != nil {
		return nil, err
	}

	configs := make([]ProviderConfig, len(providers))
	for i, p := range providers {
		configs[i] = ProviderConfig{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
			BaseURL:     p.BaseURL,
			IsActive:    p.IsActive,
		}
	}
	return configs, nil
}

// UpdateProvider updates an existing provider
func (a *AdminAPI) UpdateProvider(cfg ProviderConfig) error {
	provider := &models.Provider{
		ID:          cfg.ID,
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		BaseURL:     cfg.BaseURL,
		IsActive:    cfg.IsActive,
	}
	return a.providerService.UpdateProvider(provider)
}

// DeleteProvider soft deletes a provider
func (a *AdminAPI) DeleteProvider(id string) error {
	return a.providerService.DeleteProvider(id)
}

// ===============================
// Credential Configuration
// ===============================

// CredentialConfig represents the configuration for a credential
type CredentialConfig struct {
	ProviderID  string `json:"provider_id"`
	Key         string `json:"key"`
	Value       string `json:"value"` // Plain text value (will be encrypted)
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// CreateCredential creates a new credential with plain text value (will be encrypted)
func (a *AdminAPI) CreateCredential(cfg CredentialConfig) error {
	_, err := a.credentialService.CreateCredentialWithPlainValue(
		cfg.ProviderID,
		cfg.Key,
		cfg.Value,
		cfg.Description,
		cfg.Required,
	)
	return err
}

// GetCredentials retrieves all credentials for a provider (returns decrypted key-value map)
func (a *AdminAPI) GetCredentials(providerID string) (map[string]string, error) {
	return a.credentialService.GetDecryptedCredentialsByProviderID(providerID)
}

// GetCredentialValue retrieves a specific decrypted credential value
func (a *AdminAPI) GetCredentialValue(providerID, key string) (string, error) {
	return a.credentialService.GetCredentialValue(providerID, key)
}

// UpdateCredentialValue updates an existing credential value (finds by provider+key, then updates by ID)
func (a *AdminAPI) UpdateCredentialValue(providerID, key, newPlainValue string) error {
	// Find the credential first
	cred, err := a.credentialRepo.FindByProviderIDAndKey(providerID, key)
	if err != nil {
		return fmt.Errorf("failed to find credential: %w", err)
	}
	// Update using the service method
	return a.credentialService.UpdateCredentialValue(cred.ID, newPlainValue)
}

// DeleteCredentialByKey deletes a credential by provider ID and key
func (a *AdminAPI) DeleteCredentialByKey(providerID, key string) error {
	// Find the credential first
	cred, err := a.credentialRepo.FindByProviderIDAndKey(providerID, key)
	if err != nil {
		return fmt.Errorf("failed to find credential: %w", err)
	}
	// Delete using the service method
	return a.credentialService.DeleteCredential(cred.ID)
}

// ===============================
// Endpoint Configuration
// ===============================

// EndpointConfig represents the configuration for an endpoint
type EndpointConfig struct {
	ID          string `json:"id"`
	ProviderID  string `json:"provider_id"`
	Name        string `json:"name"`
	Method      string `json:"method"`   // HTTP method: GET, POST, etc.
	Path        string `json:"path"`     // URL path: /api/v1/orders
	BaseURL     string `json:"base_url"` // Optional override for provider base URL
	Description string `json:"description"`
}

// CreateEndpoint creates a new endpoint
func (a *AdminAPI) CreateEndpoint(cfg EndpointConfig) error {
	endpoint := &models.Endpoint{
		ID:          cfg.ID,
		ProviderID:  cfg.ProviderID,
		Name:        cfg.Name,
		Method:      cfg.Method,
		Path:        cfg.Path,
		BaseURL:     cfg.BaseURL,
		Description: cfg.Description,
	}
	return a.endpointService.CreateEndpoint(endpoint)
}

// GetEndpoint retrieves an endpoint by ID
func (a *AdminAPI) GetEndpoint(id string) (*EndpointConfig, error) {
	endpoint, err := a.endpointService.GetEndpointByID(id)
	if err != nil {
		return nil, err
	}

	return &EndpointConfig{
		ID:          endpoint.ID,
		ProviderID:  endpoint.ProviderID,
		Name:        endpoint.Name,
		Method:      endpoint.Method,
		Path:        endpoint.Path,
		BaseURL:     endpoint.BaseURL,
		Description: endpoint.Description,
	}, nil
}

// GetEndpointByProviderAndName retrieves an endpoint by provider name and endpoint name
func (a *AdminAPI) GetEndpointByProviderAndName(providerName, name string) (*EndpointConfig, error) {
	endpoint, err := a.endpointService.GetEndpointByProviderAndName(providerName, name)
	if err != nil {
		return nil, err
	}

	return &EndpointConfig{
		ID:          endpoint.ID,
		ProviderID:  endpoint.ProviderID,
		Name:        endpoint.Name,
		Method:      endpoint.Method,
		Path:        endpoint.Path,
		BaseURL:     endpoint.BaseURL,
		Description: endpoint.Description,
	}, nil
}

// ListEndpointsByProvider retrieves all endpoints for a provider
func (a *AdminAPI) ListEndpointsByProvider(providerID string) ([]EndpointConfig, error) {
	endpoints, err := a.endpointService.GetEndpointsByProviderID(providerID)
	if err != nil {
		return nil, err
	}

	configs := make([]EndpointConfig, len(endpoints))
	for i, e := range endpoints {
		configs[i] = EndpointConfig{
			ID:          e.ID,
			ProviderID:  e.ProviderID,
			Name:        e.Name,
			Method:      e.Method,
			Path:        e.Path,
			BaseURL:     e.BaseURL,
			Description: e.Description,
		}
	}
	return configs, nil
}

// UpdateEndpoint updates an existing endpoint
func (a *AdminAPI) UpdateEndpoint(cfg EndpointConfig) error {
	endpoint := &models.Endpoint{
		ID:          cfg.ID,
		ProviderID:  cfg.ProviderID,
		Name:        cfg.Name,
		Method:      cfg.Method,
		Path:        cfg.Path,
		BaseURL:     cfg.BaseURL,
		Description: cfg.Description,
	}
	return a.endpointService.UpdateEndpoint(endpoint)
}

// DeleteEndpoint soft deletes an endpoint
func (a *AdminAPI) DeleteEndpoint(id string) error {
	return a.endpointService.DeleteEndpoint(id)
}

// ===============================
// Request Schema Configuration
// ===============================

// RequestSchemaConfig represents the configuration for a request schema parameter
type RequestSchemaConfig struct {
	ID           string `json:"id"`
	EndpointID   string `json:"endpoint_id"`
	FieldName    string `json:"field_name"` // Maps to ParamName in model
	FieldType    string `json:"field_type"` // Maps to ParamType in model
	Required     bool   `json:"required"`
	Location     string `json:"location"` // Maps to ParamLocation in model
	DefaultValue string `json:"default_value"`
	Description  string `json:"description"`
}

// CreateRequestSchema creates a new request schema parameter for an endpoint
func (a *AdminAPI) CreateRequestSchema(cfg RequestSchemaConfig) error {
	schema := &models.RequestSchema{
		ID:            cfg.ID,
		EndpointID:    cfg.EndpointID,
		ParamName:     cfg.FieldName,
		ParamLocation: cfg.Location,
		ParamType:     cfg.FieldType,
		Required:      cfg.Required,
		DefaultValue:  cfg.DefaultValue,
		Description:   cfg.Description,
	}
	return a.requestSchemaService.CreateRequestSchema(schema)
}

// GetRequestSchema retrieves a request schema by ID
func (a *AdminAPI) GetRequestSchema(id string) (*RequestSchemaConfig, error) {
	schema, err := a.requestSchemaService.GetRequestSchemaByID(id)
	if err != nil {
		return nil, err
	}
	return &RequestSchemaConfig{
		ID:           schema.ID,
		EndpointID:   schema.EndpointID,
		FieldName:    schema.ParamName,
		FieldType:    schema.ParamType,
		Required:     schema.Required,
		Location:     schema.ParamLocation,
		DefaultValue: schema.DefaultValue,
		Description:  schema.Description,
	}, nil
}

// GetRequestSchemasByEndpoint retrieves all request schemas for an endpoint
func (a *AdminAPI) GetRequestSchemasByEndpoint(endpointID string) ([]*RequestSchemaConfig, error) {
	schemas, err := a.requestSchemaService.GetRequestSchemasByEndpointID(endpointID)
	if err != nil {
		return nil, err
	}

	configs := make([]*RequestSchemaConfig, len(schemas))
	for i, schema := range schemas {
		configs[i] = &RequestSchemaConfig{
			ID:           schema.ID,
			EndpointID:   schema.EndpointID,
			FieldName:    schema.ParamName,
			FieldType:    schema.ParamType,
			Required:     schema.Required,
			Location:     schema.ParamLocation,
			DefaultValue: schema.DefaultValue,
			Description:  schema.Description,
		}
	}
	return configs, nil
}

// UpdateRequestSchema updates an existing request schema
func (a *AdminAPI) UpdateRequestSchema(cfg RequestSchemaConfig) error {
	schema := &models.RequestSchema{
		ID:            cfg.ID,
		EndpointID:    cfg.EndpointID,
		ParamName:     cfg.FieldName,
		ParamLocation: cfg.Location,
		ParamType:     cfg.FieldType,
		Required:      cfg.Required,
		DefaultValue:  cfg.DefaultValue,
		Description:   cfg.Description,
	}
	return a.requestSchemaService.UpdateRequestSchema(schema)
}

// DeleteRequestSchema deletes a request schema by ID
func (a *AdminAPI) DeleteRequestSchema(id string) error {
	return a.requestSchemaService.DeleteRequestSchema(id)
}

// ===============================
// Request Value Configuration
// ===============================

// RequestValueConfig represents the configuration for a request value
type RequestValueConfig struct {
	ID         string          `json:"id"`
	SchemaID   string          `json:"schema_id"`
	Value      json.RawMessage `json:"value"`
	SourceType string          `json:"source_type"` // "static", "input", "credential", "computed"
	SourceKey  string          `json:"source_key"`
}

// CreateRequestValue creates a new request value for a schema
func (a *AdminAPI) CreateRequestValue(cfg RequestValueConfig) error {
	value := &models.RequestValue{
		ID:         cfg.ID,
		SchemaID:   cfg.SchemaID,
		Value:      datatypes.JSON(cfg.Value),
		SourceType: cfg.SourceType,
		SourceKey:  cfg.SourceKey,
	}
	return a.requestValueService.CreateRequestValue(value)
}

// GetRequestValue retrieves a request value by ID
func (a *AdminAPI) GetRequestValue(id string) (*RequestValueConfig, error) {
	value, err := a.requestValueService.GetRequestValueByID(id)
	if err != nil {
		return nil, err
	}
	return &RequestValueConfig{
		ID:         value.ID,
		SchemaID:   value.SchemaID,
		Value:      json.RawMessage(value.Value),
		SourceType: value.SourceType,
		SourceKey:  value.SourceKey,
	}, nil
}

// GetRequestValuesBySchema retrieves all request values for a schema
func (a *AdminAPI) GetRequestValuesBySchema(schemaID string) ([]*RequestValueConfig, error) {
	values, err := a.requestValueService.GetRequestValuesBySchemaID(schemaID)
	if err != nil {
		return nil, err
	}

	configs := make([]*RequestValueConfig, len(values))
	for i, value := range values {
		configs[i] = &RequestValueConfig{
			ID:         value.ID,
			SchemaID:   value.SchemaID,
			Value:      json.RawMessage(value.Value),
			SourceType: value.SourceType,
			SourceKey:  value.SourceKey,
		}
	}
	return configs, nil
}

// UpdateRequestValue updates an existing request value
func (a *AdminAPI) UpdateRequestValue(cfg RequestValueConfig) error {
	value := &models.RequestValue{
		ID:         cfg.ID,
		SchemaID:   cfg.SchemaID,
		Value:      datatypes.JSON(cfg.Value),
		SourceType: cfg.SourceType,
		SourceKey:  cfg.SourceKey,
	}
	return a.requestValueService.UpdateRequestValue(value)
}

// DeleteRequestValue deletes a request value by ID
func (a *AdminAPI) DeleteRequestValue(id string) error {
	return a.requestValueService.DeleteRequestValue(id)
}

// ===============================
// Header Rule Configuration (Direct Repository Access)
// ===============================

// HeaderRuleConfig represents the configuration for a header rule
type HeaderRuleConfig struct {
	ID              string `json:"id"`
	ProviderID      string `json:"provider_id"`
	HeaderName      string `json:"header_name"`
	ValueExpression string `json:"value_expression"`
	Priority        int    `json:"priority"`
}

// CreateHeaderRule creates a new header rule
func (a *AdminAPI) CreateHeaderRule(cfg HeaderRuleConfig) error {
	rule := &models.HeaderRule{
		ID:              cfg.ID,
		ProviderID:      cfg.ProviderID,
		HeaderName:      cfg.HeaderName,
		ValueExpression: cfg.ValueExpression,
		Priority:        cfg.Priority,
	}
	return a.headerRuleRepo.Create(rule)
}

// GetHeaderRule retrieves a header rule by ID
func (a *AdminAPI) GetHeaderRule(id string) (*HeaderRuleConfig, error) {
	rule, err := a.headerRuleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &HeaderRuleConfig{
		ID:              rule.ID,
		ProviderID:      rule.ProviderID,
		HeaderName:      rule.HeaderName,
		ValueExpression: rule.ValueExpression,
		Priority:        rule.Priority,
	}, nil
}

// ListHeaderRulesByProvider retrieves all header rules for a provider
func (a *AdminAPI) ListHeaderRulesByProvider(providerID string) ([]HeaderRuleConfig, error) {
	rules, err := a.headerRuleRepo.FindByProviderID(providerID)
	if err != nil {
		return nil, err
	}

	configs := make([]HeaderRuleConfig, len(rules))
	for i, r := range rules {
		configs[i] = HeaderRuleConfig{
			ID:              r.ID,
			ProviderID:      r.ProviderID,
			HeaderName:      r.HeaderName,
			ValueExpression: r.ValueExpression,
			Priority:        r.Priority,
		}
	}
	return configs, nil
}

// UpdateHeaderRule updates an existing header rule
func (a *AdminAPI) UpdateHeaderRule(cfg HeaderRuleConfig) error {
	rule := &models.HeaderRule{
		ID:              cfg.ID,
		ProviderID:      cfg.ProviderID,
		HeaderName:      cfg.HeaderName,
		ValueExpression: cfg.ValueExpression,
		Priority:        cfg.Priority,
	}
	return a.headerRuleRepo.Update(rule)
}

// DeleteHeaderRule deletes a header rule
func (a *AdminAPI) DeleteHeaderRule(id string) error {
	return a.headerRuleRepo.Delete(id)
}

// ===============================
// Strategy Configuration
// ===============================

// StrategyConfig represents the configuration for a strategy
type StrategyConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	StrategyType string                 `json:"strategy_type"`
	Config       map[string]interface{} `json:"config"`
}

// CreateStrategy creates a new strategy
func (a *AdminAPI) CreateStrategy(cfg StrategyConfig) error {
	configJSON, err := json.Marshal(cfg.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	strategy := &models.Strategy{
		ID:           cfg.ID,
		Name:         cfg.Name,
		StrategyType: cfg.StrategyType,
		Config:       datatypes.JSON(configJSON),
	}
	return a.strategyService.CreateStrategy(strategy)
}

// GetStrategy retrieves a strategy by ID
func (a *AdminAPI) GetStrategy(id string) (*StrategyConfig, error) {
	strategy, err := a.strategyService.GetStrategyByID(id)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(strategy.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &StrategyConfig{
		ID:           strategy.ID,
		Name:         strategy.Name,
		StrategyType: strategy.StrategyType,
		Config:       config,
	}, nil
}

// GetStrategyByName retrieves a strategy by name
func (a *AdminAPI) GetStrategyByName(name string) (*StrategyConfig, error) {
	strategy, err := a.strategyService.GetStrategyByName(name)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(strategy.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &StrategyConfig{
		ID:           strategy.ID,
		Name:         strategy.Name,
		StrategyType: strategy.StrategyType,
		Config:       config,
	}, nil
}

// ListStrategies retrieves all strategies
func (a *AdminAPI) ListStrategies() ([]StrategyConfig, error) {
	strategies, err := a.strategyService.ListStrategies()
	if err != nil {
		return nil, err
	}

	configs := make([]StrategyConfig, len(strategies))
	for i, s := range strategies {
		var config map[string]interface{}
		if err := json.Unmarshal(s.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config for strategy %s: %w", s.ID, err)
		}

		configs[i] = StrategyConfig{
			ID:           s.ID,
			Name:         s.Name,
			StrategyType: s.StrategyType,
			Config:       config,
		}
	}
	return configs, nil
}

// ListStrategiesByType retrieves all strategies of a specific type
func (a *AdminAPI) ListStrategiesByType(strategyType string) ([]StrategyConfig, error) {
	strategies, err := a.strategyService.GetStrategiesByType(strategyType)
	if err != nil {
		return nil, err
	}

	configs := make([]StrategyConfig, len(strategies))
	for i, s := range strategies {
		var config map[string]interface{}
		if err := json.Unmarshal(s.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config for strategy %s: %w", s.ID, err)
		}

		configs[i] = StrategyConfig{
			ID:           s.ID,
			Name:         s.Name,
			StrategyType: s.StrategyType,
			Config:       config,
		}
	}
	return configs, nil
}

// UpdateStrategy updates an existing strategy
func (a *AdminAPI) UpdateStrategy(cfg StrategyConfig) error {
	configJSON, err := json.Marshal(cfg.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	strategy := &models.Strategy{
		ID:           cfg.ID,
		Name:         cfg.Name,
		StrategyType: cfg.StrategyType,
		Config:       datatypes.JSON(configJSON),
	}
	return a.strategyService.UpdateStrategy(strategy)
}

// DeleteStrategy soft deletes a strategy
func (a *AdminAPI) DeleteStrategy(id string) error {
	return a.strategyService.DeleteStrategy(id)
}

// ===============================
// Response Mapping Configuration (Direct Repository Access)
// ===============================

// ResponseMappingConfig represents the configuration for a response mapping
type ResponseMappingConfig struct {
	ID             string `json:"id"`
	ProviderID     string `json:"provider_id"`
	Action         string `json:"action"`
	SourceJSONPath string `json:"source_json_path"`
	TargetField    string `json:"target_field"`
	Transform      string `json:"transform,omitempty"`
	DefaultValue   string `json:"default_value,omitempty"`
	IsRequired     bool   `json:"is_required"`
	Priority       int    `json:"priority"`
	Description    string `json:"description,omitempty"`
}

// CreateResponseMapping creates a new response mapping
func (a *AdminAPI) CreateResponseMapping(cfg ResponseMappingConfig) error {
	mapping := &models.ResponseMapping{
		ID:             cfg.ID,
		ProviderID:     cfg.ProviderID,
		Action:         cfg.Action,
		SourceJSONPath: cfg.SourceJSONPath,
		TargetField:    cfg.TargetField,
		Transform:      cfg.Transform,
		DefaultValue:   cfg.DefaultValue,
		IsRequired:     cfg.IsRequired,
		Priority:       cfg.Priority,
		Description:    cfg.Description,
	}
	return a.responseMappingRepo.Create(mapping)
}

// GetResponseMapping retrieves a response mapping by ID
func (a *AdminAPI) GetResponseMapping(id string) (*ResponseMappingConfig, error) {
	mapping, err := a.responseMappingRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &ResponseMappingConfig{
		ID:             mapping.ID,
		ProviderID:     mapping.ProviderID,
		Action:         mapping.Action,
		SourceJSONPath: mapping.SourceJSONPath,
		TargetField:    mapping.TargetField,
		Transform:      mapping.Transform,
		DefaultValue:   mapping.DefaultValue,
		IsRequired:     mapping.IsRequired,
		Priority:       mapping.Priority,
		Description:    mapping.Description,
	}, nil
}

// ListResponseMappingsByProviderAndAction retrieves all response mappings for a provider and action
func (a *AdminAPI) ListResponseMappingsByProviderAndAction(providerID, action string) ([]ResponseMappingConfig, error) {
	mappings, err := a.responseMappingRepo.FindByProviderIDAndAction(providerID, action)
	if err != nil {
		return nil, err
	}

	configs := make([]ResponseMappingConfig, len(mappings))
	for i, m := range mappings {
		configs[i] = ResponseMappingConfig{
			ID:             m.ID,
			ProviderID:     m.ProviderID,
			Action:         m.Action,
			SourceJSONPath: m.SourceJSONPath,
			TargetField:    m.TargetField,
			Transform:      m.Transform,
			DefaultValue:   m.DefaultValue,
			IsRequired:     m.IsRequired,
			Priority:       m.Priority,
			Description:    m.Description,
		}
	}
	return configs, nil
}

// UpdateResponseMapping updates an existing response mapping
func (a *AdminAPI) UpdateResponseMapping(cfg ResponseMappingConfig) error {
	mapping := &models.ResponseMapping{
		ID:             cfg.ID,
		ProviderID:     cfg.ProviderID,
		Action:         cfg.Action,
		SourceJSONPath: cfg.SourceJSONPath,
		TargetField:    cfg.TargetField,
		Transform:      cfg.Transform,
		DefaultValue:   cfg.DefaultValue,
		IsRequired:     cfg.IsRequired,
		Priority:       cfg.Priority,
		Description:    cfg.Description,
	}
	return a.responseMappingRepo.Update(mapping)
}

// DeleteResponseMapping deletes a response mapping
func (a *AdminAPI) DeleteResponseMapping(id string) error {
	return a.responseMappingRepo.Delete(id)
}
