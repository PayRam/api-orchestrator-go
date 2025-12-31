package config

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/utils"
	"github.com/PayRam/api-orchestrator-go/model"
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

// AutoMigrateOptions holds options for auto-migration
type AutoMigrateOptions struct {
	// TablePrefix sets a prefix for all orchestrator tables
	// Example: "orch_" will create tables like "orch_providers", "orch_credentials"
	// Leave empty for no prefix (default table names)
	TablePrefix string
}

// AutoMigrate runs database migrations with default options
func (conn *DBConnection) AutoMigrate() error {
	return conn.AutoMigrateWithOptions(AutoMigrateOptions{})
}

// AutoMigrateWithOptions runs database migrations with custom options
func (conn *DBConnection) AutoMigrateWithOptions(opts AutoMigrateOptions) error {
	log := utils.GetLogger()
	log.Info("Running database migrations...")

	// Set table prefix if provided
	if opts.TablePrefix != "" {
		models.SetTablePrefix(opts.TablePrefix)
		log.Info("Using table prefix", zap.String("prefix", opts.TablePrefix))
	}

	err := conn.DB.AutoMigrate(
		&model.Provider{},
		&model.ProviderCredential{},
		&model.ProviderHeaderRule{},
		&model.ProviderEndpoint{},
		&model.ProviderRequestSchema{},
		&model.ProviderRequestValue{},
		&model.Strategy{},
		&model.ProviderResponseMapping{},
		&models.Credential{},
		&models.HeaderRule{},
		&models.Provider{},
		&models.Strategy{},
		&models.Endpoint{},
		&models.RequestSchema{},
		&models.RequestValue{},
		&models.ResponseMapping{},
	)

	if err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}

// AutoMigrate runs database migrations on a raw GORM DB (standalone helper)
func AutoMigrate(db *gorm.DB) error {
	return AutoMigrateWithOptions(db, AutoMigrateOptions{})
}

// AutoMigrateWithOptions runs database migrations on a raw GORM DB with custom options (standalone helper)
func AutoMigrateWithOptions(db *gorm.DB, opts AutoMigrateOptions) error {
	log := utils.GetLogger()
	log.Info("Running database migrations...")

	// Set table prefix if provided
	if opts.TablePrefix != "" {
		models.SetTablePrefix(opts.TablePrefix)
		log.Info("Using table prefix", zap.String("prefix", opts.TablePrefix))
	}

	err := db.AutoMigrate(
		&model.Provider{},
		&model.ProviderCredential{},
		&model.ProviderHeaderRule{},
		&model.ProviderEndpoint{},
		&model.ProviderRequestSchema{},
		&model.ProviderRequestValue{},
		&model.Strategy{},
		&model.ProviderResponseMapping{},
		&models.Credential{},
		&models.HeaderRule{},
		&models.Provider{},
		&models.Strategy{},
		&models.Endpoint{},
		&models.RequestSchema{},
		&models.RequestValue{},
		&models.ResponseMapping{},
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
