package orchestrator

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/orchestrator/builder"
	"github.com/PayRam/api-orchestrator-go/orchestrator/context"
	"github.com/PayRam/api-orchestrator-go/orchestrator/executor"
	"github.com/PayRam/api-orchestrator-go/orchestrator/response"
	"github.com/PayRam/api-orchestrator-go/repository"
	"github.com/PayRam/api-orchestrator-go/service"
	"go.uber.org/zap"
)

// Orchestrator is the main orchestrator for API calls
type Orchestrator struct {
	providerService     service.ProviderService
	credentialService   service.CredentialService
	endpointService     service.EndpointService
	headerRuleRepo      repository.HeaderRuleRepo
	requestValueRepo    repository.RequestValueRepo
	responseMappingRepo repository.ResponseMappingRepo
	headerBuilder       *builder.HeaderBuilder
	requestBuilder      *builder.RequestBuilder
	httpExecutor        *executor.HTTPExecutor
	responseMapper      *response.ResponseMapper
	logger              *zap.Logger
}

// New creates a new orchestrator instance
func New(
	providerService service.ProviderService,
	credentialService service.CredentialService,
	endpointService service.EndpointService,
	headerRuleRepo repository.HeaderRuleRepo,
	requestValueRepo repository.RequestValueRepo,
	responseMappingRepo repository.ResponseMappingRepo,
	logger *zap.Logger,
) *Orchestrator {
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
	}
}

// Call executes an API call to a provider
func (o *Orchestrator) Call(providerName string, action string, params map[string]interface{}) (*response.UnifiedResponse, error) {
	o.logger.Info("Orchestrating API call",
		zap.String("provider", providerName),
		zap.String("action", action))

	// Create orchestration context
	ctx := context.NewContext(providerName, action, params)

	// Step 1: Get provider
	provider, err := o.providerService.GetProviderByName(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Step 2: Get endpoint
	endpoint, err := o.endpointService.GetEndpointByProviderAndName(providerName, action)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Step 3: Load credentials
	credentials, err := o.credentialService.GetCredentialsByProviderID(provider.ID)
	if err != nil {
		o.logger.Warn("Failed to load credentials", zap.Error(err))
	} else {
		for _, cred := range credentials {
			ctx.SetCredential(cred.Key, cred.Value)
		}
	}

	// Step 4: Build headers
	headerRules, err := o.headerRuleRepo.FindByProviderID(provider.ID)
	if err != nil {
		o.logger.Warn("Failed to load header rules", zap.Error(err))
	} else {
		err = o.headerBuilder.BuildHeaders(ctx, headerRules, provider.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to build headers: %w", err)
		}
	}

	// Step 5: Build request
	// TODO: Load request schema and values for this endpoint
	// For now, use empty request values
	requestValues := []*model.ProviderRequestValue{}

	finalRequest, err := o.requestBuilder.BuildRequest(ctx, provider, endpoint, requestValues)
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
	responseMappings, err := o.responseMappingRepo.FindByProviderIDAndAction(provider.ID, action)
	if err != nil {
		o.logger.Warn("Failed to load response mappings", zap.Error(err))
		responseMappings = []*model.ProviderResponseMapping{} // Continue without mappings
	}

	unifiedResponse, err := o.responseMapper.MapResponse(
		executionResult,
		responseMappings,
		providerName,
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
