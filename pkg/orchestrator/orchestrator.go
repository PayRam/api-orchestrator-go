// Package orchestrator provides the public API for the api-orchestrator-go library.
// This is the main entrypoint for external consumers of the library.
package orchestrator

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/builder"
	"github.com/PayRam/api-orchestrator-go/internal/context"
	"github.com/PayRam/api-orchestrator-go/internal/executor"
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"github.com/PayRam/api-orchestrator-go/internal/response"
	"github.com/PayRam/api-orchestrator-go/internal/services"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Response represents the unified response from an API call.
// This is the public response type returned by the Call method.
type Response = response.UnifiedResponse

// SetEncryptionKey sets the global encryption key for credential encryption.
// This MUST be called before creating the Orchestrator instance.
// The key must be exactly 32 bytes (256 bits) for AES-256-GCM encryption.
func SetEncryptionKey(key []byte) error {
	return models.SetEncryptionKey(key)
}

// SetEncryptionKeyFromString sets the encryption key from a string.
// The string must be exactly 32 characters long.
func SetEncryptionKeyFromString(key string) error {
	return models.SetEncryptionKeyFromString(key)
}

// IsEncryptionKeySet returns true if the encryption key has been set.
func IsEncryptionKeySet() bool {
	return models.IsEncryptionKeySet()
}

// ClearEncryptionKey clears the encryption key (useful for testing).
func ClearEncryptionKey() {
	models.ClearEncryptionKey()
}

// Orchestrator is the main orchestrator for dynamic API calls.
// It provides a clean interface for executing API requests based on
// database-driven configurations.
type Orchestrator struct {
	db                   *gorm.DB
	providerService      services.ProviderService
	credentialService    services.CredentialService
	endpointService      services.EndpointService
	strategyService      services.StrategyService
	requestSchemaService services.RequestSchemaService
	headerRuleRepo       repositories.HeaderRuleRepo
	requestSchemaRepo    repositories.RequestSchemaRepo
	requestValueRepo     repositories.RequestValueRepo
	responseMappingRepo  repositories.ResponseMappingRepo
	credentialRepo       repositories.CredentialRepo
	headerBuilder        *builder.HeaderBuilder
	requestBuilder       *builder.RequestBuilder
	httpExecutor         *executor.HTTPExecutor
	responseMapper       *response.ResponseMapper
	adminAPI             *AdminAPI
	logger               *zap.Logger
}

// Config holds the configuration for creating a new Orchestrator instance.
type Config struct {
	DB          *gorm.DB    // Required: Database connection
	Logger      *zap.Logger // Optional: Logger instance
	TablePrefix string      // Optional: Prefix for all orchestrator tables (e.g., "orch_")
}

// New creates a new Orchestrator instance with the given configuration.
// It initializes all internal repositories, services, and components.
func New(cfg Config) (*Orchestrator, error) {
	if cfg.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	// Set table prefix if provided
	if cfg.TablePrefix != "" {
		models.SetTablePrefix(cfg.TablePrefix)
		logger.Info("Using table prefix", zap.String("prefix", cfg.TablePrefix))
	}

	// Initialize repositories
	providerRepo := repositories.NewProviderRepo(cfg.DB)
	credentialRepo := repositories.NewCredentialRepo(cfg.DB)
	endpointRepo := repositories.NewEndpointRepo(cfg.DB)
	headerRuleRepo := repositories.NewHeaderRuleRepo(cfg.DB)
	requestSchemaRepo := repositories.NewRequestSchemaRepo(cfg.DB)
	requestValueRepo := repositories.NewRequestValueRepo(cfg.DB)
	responseMappingRepo := repositories.NewResponseMappingRepo(cfg.DB)
	strategyRepo := repositories.NewStrategyRepo(cfg.DB)

	// Initialize services
	providerService := services.NewProviderService(providerRepo, logger)
	credentialService := services.NewCredentialService(credentialRepo, providerService, logger)
	endpointService := services.NewEndpointService(endpointRepo, providerService, logger)
	strategyService := services.NewStrategyService(strategyRepo, logger)
	requestSchemaService := services.NewRequestSchemaService(requestSchemaRepo, endpointService, logger)
	requestValueService := services.NewRequestValueService(requestValueRepo, requestSchemaRepo, logger)

	// Initialize AdminAPI
	adminAPI := &AdminAPI{
		providerService:      providerService,
		credentialService:    credentialService,
		credentialRepo:       credentialRepo,
		endpointService:      endpointService,
		strategyService:      strategyService,
		requestSchemaService: requestSchemaService,
		requestValueService:  requestValueService,
		headerRuleRepo:       headerRuleRepo,
		requestSchemaRepo:    requestSchemaRepo,
		requestValueRepo:     requestValueRepo,
		responseMappingRepo:  responseMappingRepo,
	}

	return &Orchestrator{
		db:                   cfg.DB,
		providerService:      providerService,
		credentialService:    credentialService,
		endpointService:      endpointService,
		strategyService:      strategyService,
		requestSchemaService: requestSchemaService,
		headerRuleRepo:       headerRuleRepo,
		requestSchemaRepo:    requestSchemaRepo,
		requestValueRepo:     requestValueRepo,
		responseMappingRepo:  responseMappingRepo,
		credentialRepo:       credentialRepo,
		headerBuilder:        builder.NewHeaderBuilder(credentialService, logger),
		requestBuilder:       builder.NewRequestBuilder(requestSchemaRepo, logger),
		httpExecutor:         executor.NewHTTPExecutor(logger),
		responseMapper:       response.NewResponseMapper(logger),
		adminAPI:             adminAPI,
		logger:               logger,
	}, nil
}

// Call executes an API call to a provider with the given action and parameters.
// This is the main entrypoint for dynamic API orchestration.
//
// Parameters:
//   - provider: The name of the provider (e.g., "banxa", "transak")
//   - action: The action/endpoint to call (e.g., "create_widget_url")
//   - params: A map of input parameters for the request
//
// Returns:
//   - *Response: The unified response containing success status, data, and metadata
//   - error: Any error that occurred during the orchestration process
func (o *Orchestrator) Call(provider string, action string, params map[string]interface{}) (*Response, error) {
	o.logger.Info("Orchestrating API call",
		zap.String("provider", provider),
		zap.String("action", action))

	// Step 1: Get provider
	providerModel, err := o.providerService.GetProviderByName(provider)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Create orchestration context with the loaded provider
	ctx := context.NewContext(models.Provider{
		ID:           providerModel.ID,
		Name:         providerModel.Name,
		DisplayName:  providerModel.DisplayName,
		PipelineType: providerModel.PipelineType,
		IsActive:     providerModel.IsActive,
	}, action, params)

	// Step 2: Get endpoint
	endpoint, err := o.endpointService.GetEndpointByProviderAndName(provider, action)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Step 3: Load credentials
	credentials, err := o.credentialService.GetCredentialsByProviderID(providerModel.ID)
	if err != nil {
		o.logger.Warn("Failed to load credentials", zap.Error(err))
	} else {
		for _, cred := range credentials {
			ctx.SetCredential(cred.Key, cred.Value)
		}
	}

	// Step 4: Build headers
	headerRules, err := o.headerRuleRepo.FindByProviderID(providerModel.ID)
	if err != nil {
		o.logger.Warn("Failed to load header rules", zap.Error(err))
	} else {
		err = o.headerBuilder.BuildHeaders(ctx, headerRules, providerModel.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to build headers: %w", err)
		}
	}

	// Step 5: Build request
	requestValues := []*models.RequestValue{}
	finalRequest, err := o.requestBuilder.BuildRequest(ctx, providerModel, endpoint, requestValues)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	o.logger.Info("Request built successfully",
		zap.String("method", finalRequest.Method),
		zap.String("url", finalRequest.URL))
	o.logger.Debug("Curl command", zap.String("curl", finalRequest.CurlCommand))

	// Step 6: Execute request
	executionResult, err := o.httpExecutor.Execute(finalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Step 7: Map response
	responseMappings, err := o.responseMappingRepo.FindByProviderIDAndAction(providerModel.ID, action)
	if err != nil {
		o.logger.Warn("Failed to load response mappings", zap.Error(err))
		responseMappings = []*models.ResponseMapping{}
	}

	unifiedResponse, err := o.responseMapper.MapResponse(
		executionResult,
		responseMappings,
		provider,
		action,
		ctx.RequestID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to map response: %w", err)
	}

	o.logger.Info("API call completed successfully",
		zap.Bool("success", unifiedResponse.Success),
		zap.Int("status_code", unifiedResponse.StatusCode))

	return unifiedResponse, nil
}

// Admin returns the AdminAPI for managing providers, credentials, endpoints, and configurations.
// This provides a clean interface for external Go projects to configure the orchestrator.
func (o *Orchestrator) Admin() *AdminAPI {
	return o.adminAPI
}
