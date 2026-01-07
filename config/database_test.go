package config

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAutoMigrateWithPrefixDoesNotCreateDuplicateTables tests that AutoMigrateWithOptions
// creates only prefixed tables when prefix is specified, not both prefixed and non-prefixed
func TestAutoMigrateWithPrefixDoesNotCreateDuplicateTables(t *testing.T) {
	// Create in-memory database with a unique name to avoid any caching issues
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	assert.NoError(t, err)

	// Run migration with prefix (this will call models.SetTablePrefix internally)
	err = AutoMigrateWithOptions(db, AutoMigrateOptions{
		TablePrefix: "test_",
	})
	assert.NoError(t, err)

	// Get all table names
	var tableNames []string
	rows, err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Rows()
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		assert.NoError(t, err)
		tableNames = append(tableNames, tableName)
	}

	t.Logf("Created tables: %v", tableNames)

	// Verify only prefixed tables exist
	expectedTables := []string{
		"test_providers",
		"test_credentials",
		"test_endpoints",
		"test_header_rules",
		"test_strategies",
		"test_request_schemas",
		"test_request_values",
		"test_response_mappings",
	}

	// Check that all expected prefixed tables exist
	for _, expectedTable := range expectedTables {
		assert.Contains(t, tableNames, expectedTable, "Expected prefixed table %s not found", expectedTable)
	}

	// Check that NO non-prefixed tables exist (the bug was creating both)
	unprefixedTables := []string{
		"providers",
		"credentials",
		"endpoints",
		"header_rules",
		"strategies",
		"request_schemas",
		"request_values",
		"response_mappings",
		"provider_credentials", // old model names
		"provider_header_rules",
		"provider_endpoints",
		"provider_request_schemas",
		"provider_request_values",
		"provider_response_mappings",
	}

	for _, unprefixedTable := range unprefixedTables {
		assert.NotContains(t, tableNames, unprefixedTable, "Found non-prefixed table %s - this indicates the bug still exists", unprefixedTable)
	}

	// Verify we have exactly the expected number of tables (plus sqlite_sequence)
	// SQLite creates sqlite_sequence automatically for autoincrement
	expectedTableCount := len(expectedTables)
	actualOrchestratorTables := 0
	for _, table := range tableNames {
		if table != "sqlite_sequence" {
			actualOrchestratorTables++
		}
	}

	assert.Equal(t, expectedTableCount, actualOrchestratorTables,
		"Expected exactly %d orchestrator tables, got %d. This suggests duplicate tables were created.",
		expectedTableCount, actualOrchestratorTables)

	t.Logf("✅ Bug fix verified: Only %d prefixed tables created, no duplicates", actualOrchestratorTables)
}

// TestAutoMigrateWithoutPrefixCreatesDefaultTables tests that without prefix,
// tables are created with default names
func TestAutoMigrateWithoutPrefixCreatesDefaultTables(t *testing.T) {
	// Reset any previous prefix from other tests
	models.ResetTablePrefix()

	// Create in-memory database with a unique name to avoid any caching issues
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	assert.NoError(t, err)

	// Run migration without prefix
	err = AutoMigrateWithOptions(db, AutoMigrateOptions{})
	assert.NoError(t, err)

	// Get all table names
	var tableNames []string
	rows, err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").Rows()
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		assert.NoError(t, err)
		tableNames = append(tableNames, tableName)
	}

	t.Logf("Created tables: %v", tableNames)

	// Verify default table names exist
	expectedTables := []string{
		"providers",
		"credentials",
		"endpoints",
		"header_rules",
		"strategies",
		"request_schemas",
		"request_values",
		"response_mappings",
	}

	for _, expectedTable := range expectedTables {
		assert.Contains(t, tableNames, expectedTable, "Expected default table %s not found", expectedTable)
	}

	t.Logf("✅ Default tables created correctly: %d tables", len(expectedTables))
}
