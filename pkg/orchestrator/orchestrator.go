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

// Orchestrator is the main orchestrator for dynamic API calls.
// It provides a clean interface for executing API requests based on
// database-driven configurations.
type Orchestrator struct {
	providerService     services.ProviderService
	credentialService   services.CredentialService
	endpointService     services.EndpointService
	headerRuleRepo      repositories.HeaderRuleRepo
	requestValueRepo    repositories.RequestValueRepo
	responseMappingRepo repositories.ResponseMappingRepo
	headerBuilder       *builder.HeaderBuilder
	requestBuilder      *builder.RequestBuilder
	httpExecutor        *executor.HTTPExecutor
	responseMapper      *response.ResponseMapper
	logger              *zap.Logger
}

// Config holds the configuration for creating a new Orchestrator instance.
type Config struct {
	DB     *gorm.DB
	Logger *zap.Logger
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

	// Initialize repositories
	providerRepo := repositories.NewProviderRepo(cfg.DB)
	credentialRepo := repositories.NewCredentialRepo(cfg.DB)
	endpointRepo := repositories.NewEndpointRepo(cfg.DB)
	headerRuleRepo := repositories.NewHeaderRuleRepo(cfg.DB)
	requestValueRepo := repositories.NewRequestValueRepo(cfg.DB)
	responseMappingRepo := repositories.NewResponseMappingRepo(cfg.DB)

	// Initialize services
	providerService := services.NewProviderService(providerRepo, logger)
	credentialService := services.NewCredentialService(credentialRepo, providerService, logger)
	endpointService := services.NewEndpointService(endpointRepo, providerService, logger)

	return &Orchestrator{
		providerService:     providerService,
		credentialService:   credentialService,
		endpointService:     endpointService,
		headerRuleRepo:      headerRuleRepo,
		requestValueRepo:    requestValueRepo,
		responseMappingRepo: responseMappingRepo,
		headerBuilder:       builder.NewHeaderBuilder(credentialService, logger),
		requestBuilder:      builder.NewRequestBuilder(),
		httpExecutor:        executor.NewHTTPExecutor(logger),
		responseMapper:      response.NewResponseMapper(logger),
		logger:              logger,
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
