package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEndToEndTablePrefix simulates a complete real-world scenario
// where an application with existing tables integrates the orchestrator library
func TestEndToEndTablePrefix(t *testing.T) {
	// Scenario: An e-commerce app already has a "providers" table for shipping providers
	// They need to integrate the orchestrator library without conflicts

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Reset prefix for clean test
	ResetTablePrefix()

	t.Run("ExistingAppTables", func(t *testing.T) {
		// Simulate existing app tables
		type ShippingProvider struct {
			ID   uint   `gorm:"primaryKey"`
			Name string `gorm:"not null"`
			Type string
		}

		// Create existing app's provider table
		err := db.Table("providers").AutoMigrate(&ShippingProvider{})
		require.NoError(t, err)

		// Add existing data
		existingProvider := ShippingProvider{Name: "FedEx", Type: "shipping"}
		err = db.Table("providers").Create(&existingProvider).Error
		require.NoError(t, err)

		// Verify existing data is there
		var count int64
		db.Table("providers").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("IntegrateOrchestratorWithPrefix", func(t *testing.T) {
		// Now integrate orchestrator with prefix to avoid conflicts
		SetTablePrefix("orch_")

		// Migrate orchestrator tables
		err := db.AutoMigrate(
			&Provider{},
			&Credential{},
			&Endpoint{},
			&HeaderRule{},
			&Strategy{},
			&RequestSchema{},
			&RequestValue{},
			&ResponseMapping{},
		)
		require.NoError(t, err)

		// Create orchestrator provider (different from shipping provider)
		orchProvider := &Provider{
			ID:           uuid.New().String(),
			Name:         "banxa",
			DisplayName:  "Banxa",
			PipelineType: "generic",
			BaseURL:      "https://api.banxa.com",
			IsActive:     true,
		}
		err = db.Create(orchProvider).Error
		require.NoError(t, err)

		// Verify orchestrator provider is in orch_providers table
		var orchCount int64
		db.Table("orch_providers").Count(&orchCount)
		assert.Equal(t, int64(1), orchCount)

		// Verify existing providers table is untouched
		var existingCount int64
		db.Table("providers").Count(&existingCount)
		assert.Equal(t, int64(1), existingCount, "Existing providers table should still have 1 record")

		// Create orchestrator credential
		credential := &Credential{
			ID:         uuid.New().String(),
			ProviderID: orchProvider.ID,
			Key:        "API_KEY",
			Value:      "test_key_123",
		}
		err = db.Create(credential).Error
		require.NoError(t, err)

		// Create orchestrator endpoint
		endpoint := &Endpoint{
			ID:         uuid.New().String(),
			ProviderID: orchProvider.ID,
			Name:       "create_order",
			Method:     "POST",
			Path:       "/orders",
		}
		err = db.Create(endpoint).Error
		require.NoError(t, err)

		// Verify all tables exist with correct names
		var tables []string
		db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'orch_%' ORDER BY name").Scan(&tables)

		expectedTables := []string{
			"orch_credentials",
			"orch_endpoints",
			"orch_header_rules",
			"orch_providers",
			"orch_request_schemas",
			"orch_request_values",
			"orch_response_mappings",
			"orch_strategies",
		}

		assert.Equal(t, len(expectedTables), len(tables), "Should have all orchestrator tables with prefix")
		for _, expectedTable := range expectedTables {
			assert.Contains(t, tables, expectedTable, "Table %s should exist", expectedTable)
		}
	})

	t.Run("BothTablesCoexist", func(t *testing.T) {
		// Verify both tables coexist and work independently

		// Query existing providers table
		type ShippingProvider struct {
			ID   uint   `gorm:"primaryKey"`
			Name string `gorm:"not null"`
			Type string
		}
		var shippingProvider ShippingProvider
		err := db.Table("providers").First(&shippingProvider).Error
		require.NoError(t, err)
		assert.Equal(t, "FedEx", shippingProvider.Name)
		assert.Equal(t, "shipping", shippingProvider.Type)

		// Query orchestrator providers table
		var orchProvider Provider
		err = db.Where("name = ?", "banxa").First(&orchProvider).Error
		require.NoError(t, err)
		assert.Equal(t, "banxa", orchProvider.Name)
		assert.Equal(t, "Banxa", orchProvider.DisplayName)

		// Verify they are completely separate
		var providerCount, orchProviderCount int64
		db.Table("providers").Count(&providerCount)
		db.Table("orch_providers").Count(&orchProviderCount)

		assert.Equal(t, int64(1), providerCount, "Original providers table")
		assert.Equal(t, int64(1), orchProviderCount, "Orchestrator providers table")
	})
}

// TestTablePrefixWithConfigPackage tests integration with config package
func TestTablePrefixWithConfigPackage(t *testing.T) {
	// This test would require the config package to be available
	// For now, we'll test the models directly

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	ResetTablePrefix()
	SetTablePrefix("api_")

	// Simulate what config.AutoMigrateWithOptions would do
	err = db.AutoMigrate(
		&Provider{},
		&Credential{},
		&Endpoint{},
		&HeaderRule{},
		&Strategy{},
		&RequestSchema{},
		&RequestValue{},
		&ResponseMapping{},
	)
	require.NoError(t, err)

	// Verify tables were created with api_ prefix
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'api_%' ORDER BY name").Scan(&tables)

	assert.True(t, len(tables) == 8, "Should have 8 tables with api_ prefix")

	// Create test data
	provider := &Provider{
		ID:           uuid.New().String(),
		Name:         "test_api",
		DisplayName:  "Test API",
		PipelineType: "generic",
		IsActive:     true,
	}
	err = db.Create(provider).Error
	require.NoError(t, err)

	// Query it back
	var retrieved Provider
	err = db.First(&retrieved, "name = ?", "test_api").Error
	require.NoError(t, err)
	assert.Equal(t, "test_api", retrieved.Name)
}

// TestTablePrefixThreadSafety tests that prefix setting is thread-safe
func TestTablePrefixThreadSafety(t *testing.T) {
	ResetTablePrefix()

	// Set prefix
	SetTablePrefix("thread_test_")

	// Try to set it again (should be ignored due to sync.Once)
	SetTablePrefix("different_prefix_")

	// Verify original prefix is still set
	assert.Equal(t, "thread_test_", GetTablePrefix())

	// Verify table names use the original prefix
	provider := &Provider{}
	assert.Equal(t, "thread_test_providers", provider.TableName())

	credential := &Credential{}
	assert.Equal(t, "thread_test_credentials", credential.TableName())
}
