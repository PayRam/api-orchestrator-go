package config

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/model"
	"github.com/PayRam/api-orchestrator-go/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnection holds the database connection
type DBConnection struct {
	DB *gorm.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg *DatabaseConfig) (*DBConnection, error) {
	dsn := cfg.GetDSN()
	log := utils.GetLogger()

	log.Info("Connecting to database", zap.String("host", cfg.Host), zap.Int("port", cfg.Port))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection established successfully")

	return &DBConnection{DB: db}, nil
}

// AutoMigrate runs database migrations
func (conn *DBConnection) AutoMigrate() error {
	log := utils.GetLogger()
	log.Info("Running database migrations...")

	err := conn.DB.AutoMigrate(
		&model.Provider{},
		&model.ProviderCredential{},
		&model.ProviderHeaderRule{},
		&model.ProviderEndpoint{},
		&model.ProviderRequestSchema{},
		&model.ProviderRequestValue{},
		&model.Strategy{},
		&model.ProviderResponseMapping{},
	)

	if err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (conn *DBConnection) Close() error {
	sqlDB, err := conn.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM DB instance
func (conn *DBConnection) GetDB() *gorm.DB {
	return conn.DB
}
