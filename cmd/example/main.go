package main

import (
	"log"

	"github.com/PayRam/api-orchestrator-go/config"
	"github.com/PayRam/api-orchestrator-go/repository"
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
	logger.Info("Starting API Orchestrator Example")

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
	strategyRepo := repository.NewStrategyRepo(db)
	responseMappingRepo := repository.NewResponseMappingRepo(db)

	logger.Info("All repositories initialized successfully")

	// Example: List all providers
	providers, err := providerRepo.List()
	if err != nil {
		logger.Error("Failed to list providers", zap.Error(err))
	} else {
		logger.Info("Found providers", zap.Int("count", len(providers)))
	}

	// TODO: In Phase 2+, you'll initialize the orchestrator service here
	// orchestrator := orchestrator.New(db)
	// response, err := orchestrator.Call("banxa", "create_widget_url", params)

	logger.Info("API Orchestrator is running...")

	// Keep repositories in scope (avoid unused variable errors)
	_ = credentialRepo
	_ = headerRuleRepo
	_ = endpointRepo
	_ = requestValueRepo
	_ = strategyRepo
	_ = responseMappingRepo
}
