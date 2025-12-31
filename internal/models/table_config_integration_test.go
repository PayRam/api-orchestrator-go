package models

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestTablePrefixIntegration tests the full flow with a real database
func TestTablePrefixIntegration(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Reset prefix for clean test
	ResetTablePrefix()

	t.Run("WithPrefix", func(t *testing.T) {
		// Set prefix
		SetTablePrefix("test_")

		// Create tables
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
		assert.NoError(t, err)

		// Verify tables were created with prefix
		var tables []string
		db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Scan(&tables)

		expectedTables := []string{
			"test_credentials",
			"test_endpoints",
			"test_header_rules",
			"test_providers",
			"test_request_schemas",
			"test_request_values",
			"test_response_mappings",
			"test_strategies",
		}

		for _, expectedTable := range expectedTables {
			assert.Contains(t, tables, expectedTable, "Table %s should exist", expectedTable)
		}

		// Test inserting and retrieving data
		provider := &Provider{
			ID:           uuid.New().String(),
			Name:         "test_provider",
			DisplayName:  "Test Provider",
			PipelineType: "generic",
			BaseURL:      "https://api.test.com",
			IsActive:     true,
		}

		err = db.Create(provider).Error
		assert.NoError(t, err)
		assert.NotEmpty(t, provider.ID)

		// Retrieve the provider
		var retrieved Provider
		err = db.Where("name = ?", "test_provider").First(&retrieved).Error
		assert.NoError(t, err)
		assert.Equal(t, "test_provider", retrieved.Name)
		assert.Equal(t, "Test Provider", retrieved.DisplayName)
	})
}

// TestTablePrefixWithEnvironment tests using environment variable
func TestTablePrefixWithEnvironment(t *testing.T) {
	// Reset prefix for clean test
	ResetTablePrefix()

	// Set environment variable
	os.Setenv("ORCHESTRATOR_TABLE_PREFIX", "env_")
	defer os.Unsetenv("ORCHESTRATOR_TABLE_PREFIX")

	// Set prefix from environment
	prefix := os.Getenv("ORCHESTRATOR_TABLE_PREFIX")
	SetTablePrefix(prefix)

	// Verify prefix is set
	assert.Equal(t, "env_", GetTablePrefix())

	// Create a provider to verify table name
	provider := &Provider{Name: "env_test"}
	tableName := provider.TableName()
	assert.Equal(t, "env_providers", tableName)
}

// TestMultipleModelsWithPrefix tests all models use the prefix correctly
func TestMultipleModelsWithPrefix(t *testing.T) {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Reset prefix for clean test
	ResetTablePrefix()
	SetTablePrefix("multi_")

	// Create all tables
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
	assert.NoError(t, err)

	// Create a provider
	provider := &Provider{
		ID:           uuid.New().String(),
		Name:         "multi_provider",
		DisplayName:  "Multi Provider",
		PipelineType: "generic",
		BaseURL:      "https://api.multi.com",
		IsActive:     true,
	}
	err = db.Create(provider).Error
	assert.NoError(t, err)

	// Create a credential linked to the provider
	credential := &Credential{
		ID:         uuid.New().String(),
		ProviderID: provider.ID,
		Key:        "test_key",
		Value:      "test_value",
	}
	err = db.Create(credential).Error
	assert.NoError(t, err)

	// Create an endpoint linked to the provider
	endpoint := &Endpoint{
		ID:         uuid.New().String(),
		ProviderID: provider.ID,
		Name:       "test_endpoint",
		Method:     "GET",
		Path:       "/test",
	}
	err = db.Create(endpoint).Error
	assert.NoError(t, err)

	// Create a strategy
	strategy := &Strategy{
		ID:           uuid.New().String(),
		Name:         "test_strategy",
		StrategyType: "SIGNATURE",
	}
	err = db.Create(strategy).Error
	assert.NoError(t, err)

	// Verify all records can be retrieved
	var retrievedProvider Provider
	err = db.Where("name = ?", "multi_provider").First(&retrievedProvider).Error
	assert.NoError(t, err)
	assert.Equal(t, "multi_provider", retrievedProvider.Name)

	var retrievedCredential Credential
	err = db.Where("key = ?", "test_key").First(&retrievedCredential).Error
	assert.NoError(t, err)
	assert.Equal(t, "test_key", retrievedCredential.Key)

	var retrievedEndpoint Endpoint
	err = db.Where("name = ?", "test_endpoint").First(&retrievedEndpoint).Error
	assert.NoError(t, err)
	assert.Equal(t, "test_endpoint", retrievedEndpoint.Name)

	var retrievedStrategy Strategy
	err = db.Where("name = ?", "test_strategy").First(&retrievedStrategy).Error
	assert.NoError(t, err)
	assert.Equal(t, "test_strategy", retrievedStrategy.Name)

	// Verify relationships work by querying with provider ID
	var credentials []Credential
	err = db.Where("provider_id = ?", provider.ID).Find(&credentials).Error
	assert.NoError(t, err)
	assert.Len(t, credentials, 1)
	assert.Equal(t, "test_key", credentials[0].Key)

	var endpoints []Endpoint
	err = db.Where("provider_id = ?", provider.ID).Find(&endpoints).Error
	assert.NoError(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "test_endpoint", endpoints[0].Name)
}
