package main

import (
	"log"

	"github.com/PayRam/api-orchestrator-go/config"
	"github.com/PayRam/api-orchestrator-go/orchestrator"
	"github.com/PayRam/api-orchestrator-go/repository"
	"github.com/PayRam/api-orchestrator-go/service"
	"github.com/PayRam/api-orchestrator-go/utils"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Initialize logger
	if err := utils.InitLogger(cfg.Logger.Environment, cfg.Logger.Level); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer utils.Sync()

	logger := utils.GetLogger()
	logger.Info("Starting API Orchestrator")

	// Connect to database
	dbConn, err := config.NewConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	// Run migrations
	if err := dbConn.AutoMigrate(); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize repositories
	db := dbConn.GetDB()
	providerRepo := repository.NewProviderRepo(db)
	credentialRepo := repository.NewCredentialRepo(db)
	headerRuleRepo := repository.NewHeaderRuleRepo(db)
	endpointRepo := repository.NewEndpointRepo(db)
	requestValueRepo := repository.NewRequestValueRepo(db)
	responseMappingRepo := repository.NewResponseMappingRepo(db)
	strategyRepo := repository.NewStrategyRepo(db)

	// Initialize services (services depend on other services, not repos directly)
	providerService := service.NewProviderService(providerRepo, logger)
	credentialService := service.NewCredentialService(credentialRepo, providerService, logger)
	endpointService := service.NewEndpointService(endpointRepo, providerService, logger)
	strategyService := service.NewStrategyService(strategyRepo, logger)

	logger.Info("All services initialized successfully")

	// Initialize orchestrator
	orch := orchestrator.New(
		providerService,
		credentialService,
		endpointService,
		headerRuleRepo,
		requestValueRepo,
		responseMappingRepo,
		logger,
	)

	logger.Info("Orchestrator initialized successfully")

	// Example usage (commented out - requires DB data):
	// params := map[string]interface{}{
	// 	"account_reference": "user123",
	// 	"source_currency":   "USD",
	// 	"target_currency":   "BTC",
	// 	"wallet_address":    "0x123...",
	// }
	//
	// response, err := orch.Call("banxa", "create_widget_url", params)
	// if err != nil {
	// 	logger.Error("API call failed", zap.Error(err))
	// 	return
	// }
	//
	// logger.Info("API call successful",
	// 	zap.Bool("success", response.Success),
	// 	zap.Int("status_code", response.StatusCode),
	// 	zap.Any("data", response.Data))

	// List all providers as an example
	providers, err := providerService.ListProviders()
	if err != nil {
		logger.Error("Failed to list providers", zap.Error(err))
	} else {
		logger.Info("Found providers", zap.Int("count", len(providers)))
	}

	// Keep variables in scope
	_ = orch
	_ = strategyService

	logger.Info("API Orchestrator is ready")
}
