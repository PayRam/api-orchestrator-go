package main

import (
	"log"

	"github.com/PayRam/api-orchestrator-go/config"
	"github.com/PayRam/api-orchestrator-go/internal/utils"
	"github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
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

	// Initialize orchestrator using the new public API
	db := dbConn.GetDB()
	orch, err := orchestrator.New(orchestrator.Config{
		DB:     db,
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("Failed to initialize orchestrator", zap.Error(err))
	}

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

	// Keep orchestrator in scope
	_ = orch

	logger.Info("API Orchestrator is ready")
}
