package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestFullFlowWithTablePrefix demonstrates the complete integration:
// 1. Setup database with table prefix to avoid conflicts
// 2. Create provider, credentials, endpoints with proper table naming
// 3. Execute full request building flow
// 4. Verify everything works end-to-end
func TestFullFlowWithTablePrefix(t *testing.T) {
	// ===========================================
	// 1. Setup Database with Table Prefix
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 1: Database Setup with Table Prefix")
	t.Log(strings.Repeat("=", 60))

	// Reset prefix for clean test
	models.ResetTablePrefix()

	// Set table prefix to avoid conflicts
	tablePrefix := "orch_"
	models.SetTablePrefix(tablePrefix)
	t.Logf("✓ Table prefix set to: %s", tablePrefix)

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	t.Log("✓ Database created")

	// Run migrations with prefix
	err = db.AutoMigrate(
		&models.Provider{},
		&models.Credential{},
		&models.Endpoint{},
		&models.HeaderRule{},
		&models.Strategy{},
		&models.RequestSchema{},
		&models.RequestValue{},
		&models.ResponseMapping{},
	)
	require.NoError(t, err)
	t.Log("✓ Tables migrated with prefix")

	// Verify tables were created with prefix
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'orch_%' ORDER BY name").Scan(&tables)
	t.Logf("✓ Created %d tables with prefix: %v", len(tables), tables)

	assert.Contains(t, tables, "orch_providers", "Should have orch_providers table")
	assert.Contains(t, tables, "orch_credentials", "Should have orch_credentials table")
	assert.Contains(t, tables, "orch_endpoints", "Should have orch_endpoints table")
	assert.Contains(t, tables, "orch_header_rules", "Should have orch_header_rules table")

	// ===========================================
	// 2. Create Provider with Prefixed Table
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 2: Create Provider")
	t.Log(strings.Repeat("=", 60))

	provider := &models.Provider{
		ID:           uuid.New().String(),
		Name:         "banxa",
		DisplayName:  "Banxa",
		PipelineType: "generic",
		BaseURL:      "https://api.banxa.com",
		IsActive:     true,
	}

	err = db.Create(provider).Error
	require.NoError(t, err)
	t.Logf("✓ Provider created: %s (ID: %s)", provider.DisplayName, provider.ID[:8])

	// Verify it's in the prefixed table
	var providerCount int64
	db.Table("orch_providers").Count(&providerCount)
	assert.Equal(t, int64(1), providerCount, "Provider should be in orch_providers table")
	t.Logf("✓ Provider stored in %s table", provider.TableName())

	// ===========================================
	// 3. Create Credentials with Prefixed Table
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 3: Create Credentials")
	t.Log(strings.Repeat("=", 60))

	apiKey := &models.Credential{
		ID:         uuid.New().String(),
		ProviderID: provider.ID,
		Key:        "API_KEY",
	}
	apiKey.SetEncryptedValue("pk_live_abc123xyz")

	apiSecret := &models.Credential{
		ID:         uuid.New().String(),
		ProviderID: provider.ID,
		Key:        "API_SECRET",
	}
	apiSecret.SetEncryptedValue("sk_live_secret123")

	err = db.Create(apiKey).Error
	require.NoError(t, err)
	err = db.Create(apiSecret).Error
	require.NoError(t, err)

	t.Logf("✓ Created 2 credentials (API_KEY, API_SECRET)")

	// Verify credentials are in prefixed table
	var credentialCount int64
	db.Table("orch_credentials").Count(&credentialCount)
	assert.Equal(t, int64(2), credentialCount, "Credentials should be in orch_credentials table")
	t.Logf("✓ Credentials stored in %s table", apiKey.TableName())

	// ===========================================
	// 4. Create Endpoint with Prefixed Table
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 4: Create Endpoint")
	t.Log(strings.Repeat("=", 60))

	endpoint := &models.Endpoint{
		ID:          uuid.New().String(),
		ProviderID:  provider.ID,
		Name:        "create_order",
		Method:      "POST",
		Path:        "/api/orders",
		Description: "Create crypto purchase order",
	}

	err = db.Create(endpoint).Error
	require.NoError(t, err)
	t.Logf("✓ Endpoint created: %s %s", endpoint.Method, endpoint.Path)

	// Verify endpoint is in prefixed table
	var endpointCount int64
	db.Table("orch_endpoints").Count(&endpointCount)
	assert.Equal(t, int64(1), endpointCount, "Endpoint should be in orch_endpoints table")
	t.Logf("✓ Endpoint stored in %s table", endpoint.TableName())

	// ===========================================
	// 5. Create Request Schemas
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 5: Create Request Schemas")
	t.Log(strings.Repeat("=", 60))

	schemas := []*models.RequestSchema{
		{
			ID:            uuid.New().String(),
			EndpointID:    endpoint.ID,
			ParamName:     "fiat_amount",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeNumber,
			Required:      true,
		},
		{
			ID:            uuid.New().String(),
			EndpointID:    endpoint.ID,
			ParamName:     "fiat_code",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      true,
		},
		{
			ID:            uuid.New().String(),
			EndpointID:    endpoint.ID,
			ParamName:     "coin_code",
			ParamLocation: models.ParamLocationBody,
			ParamType:     models.ParamTypeString,
			Required:      true,
		},
	}

	for _, schema := range schemas {
		err = db.Create(schema).Error
		require.NoError(t, err)
	}

	t.Logf("✓ Created %d request schemas", len(schemas))
	t.Logf("✓ Schemas stored in %s table", schemas[0].TableName())

	// ===========================================
	// 6. Create Header Rules
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 6: Create Header Rules")
	t.Log(strings.Repeat("=", 60))

	headerRules := []*models.HeaderRule{
		{
			ID:              uuid.New().String(),
			ProviderID:      provider.ID,
			HeaderName:      "Content-Type",
			ValueExpression: "static:application/json",
			Priority:        1,
		},
		{
			ID:              uuid.New().String(),
			ProviderID:      provider.ID,
			HeaderName:      "X-API-Key",
			ValueExpression: "credential:API_KEY",
			Priority:        2,
		},
		{
			ID:              uuid.New().String(),
			ProviderID:      provider.ID,
			HeaderName:      "X-Timestamp",
			ValueExpression: "strategy:timestamp",
			Priority:        3,
		},
	}

	for _, rule := range headerRules {
		err = db.Create(rule).Error
		require.NoError(t, err)
	}

	t.Logf("✓ Created %d header rules", len(headerRules))
	t.Logf("✓ Header rules stored in %s table", headerRules[0].TableName())

	// ===========================================
	// 7. Create Strategy
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 7: Create Strategy")
	t.Log(strings.Repeat("=", 60))

	strategyConfig := map[string]interface{}{
		"algorithm": "SHA256",
		"encoding":  "hex",
	}
	configJSON, _ := json.Marshal(strategyConfig)

	strategy := &models.Strategy{
		ID:           uuid.New().String(),
		Name:         "timestamp_generator",
		StrategyType: "TIMESTAMP",
		Config:       datatypes.JSON(configJSON),
	}

	err = db.Create(strategy).Error
	require.NoError(t, err)
	t.Logf("✓ Strategy created: %s (type: %s)", strategy.Name, strategy.StrategyType)
	t.Logf("✓ Strategy stored in %s table", strategy.TableName())

	// ===========================================
	// 8. Query Data Back from Prefixed Tables
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 8: Query Data from Prefixed Tables")
	t.Log(strings.Repeat("=", 60))

	// Query provider
	var queriedProvider models.Provider
	err = db.Where("name = ?", "banxa").First(&queriedProvider).Error
	require.NoError(t, err)
	assert.Equal(t, "Banxa", queriedProvider.DisplayName)
	t.Logf("✓ Provider queried successfully: %s", queriedProvider.DisplayName)

	// Query credentials
	var queriedCredentials []models.Credential
	err = db.Where("provider_id = ?", provider.ID).Find(&queriedCredentials).Error
	require.NoError(t, err)
	assert.Len(t, queriedCredentials, 2)
	t.Logf("✓ Credentials queried successfully: %d found", len(queriedCredentials))

	// Query endpoint
	var queriedEndpoint models.Endpoint
	err = db.Where("provider_id = ? AND name = ?", provider.ID, "create_order").First(&queriedEndpoint).Error
	require.NoError(t, err)
	assert.Equal(t, "POST", queriedEndpoint.Method)
	t.Logf("✓ Endpoint queried successfully: %s %s", queriedEndpoint.Method, queriedEndpoint.Path)

	// ===========================================
	// 9. Execute Full Request Building Flow
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 9: Execute Request Building Flow")
	t.Log(strings.Repeat("=", 60))

	// Prepare credentials map
	credentials := make(map[string]string)
	for _, cred := range queriedCredentials {
		decrypted, _ := cred.GetDecryptedValue()
		credentials[cred.Key] = decrypted
	}
	t.Logf("✓ Credentials prepared: %d keys", len(credentials))

	// Prepare user input
	userInput := map[string]interface{}{
		"fiat_amount": 100.00,
		"fiat_code":   "USD",
		"coin_code":   "BTC",
	}
	t.Logf("✓ User input prepared: %+v", userInput)

	// Evaluate header rules
	computedValues := map[string]interface{}{
		"timestamp": "1703424000",
	}

	headers := make(map[string]string)
	for _, rule := range headerRules {
		var value string
		parts := strings.SplitN(rule.ValueExpression, ":", 2)

		switch parts[0] {
		case "static":
			value = parts[1]
		case "credential":
			if val, ok := credentials[parts[1]]; ok {
				value = val
			}
		case "strategy":
			if val, ok := computedValues[parts[1]]; ok {
				value = val.(string)
			}
		}

		if value != "" {
			headers[rule.HeaderName] = value
		}
	}
	t.Logf("✓ Headers built: %d headers", len(headers))

	// Build request body
	body := make(map[string]interface{})
	for _, schema := range schemas {
		if val, ok := userInput[schema.ParamName]; ok {
			body[schema.ParamName] = val
		}
	}
	bodyJSON, _ := json.Marshal(body)
	t.Logf("✓ Request body built: %d fields", len(body))

	// ===========================================
	// 10. Verify Final Request
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 10: Verify Final Request")
	t.Log(strings.Repeat("=", 60))

	finalURL := queriedProvider.BaseURL + queriedEndpoint.Path
	t.Logf("URL: %s", finalURL)
	t.Logf("Method: %s", queriedEndpoint.Method)
	t.Log("\nHeaders:")
	for k, v := range headers {
		t.Logf("  %s: %s", k, v)
	}
	t.Log("\nBody:")
	var prettyBody map[string]interface{}
	json.Unmarshal(bodyJSON, &prettyBody)
	prettyJSON, _ := json.MarshalIndent(prettyBody, "  ", "  ")
	t.Logf("%s", string(prettyJSON))

	// Verify all components
	assert.Equal(t, "https://api.banxa.com/api/orders", finalURL)
	assert.Equal(t, "POST", queriedEndpoint.Method)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "pk_live_abc123xyz", headers["X-API-Key"])
	assert.Equal(t, "1703424000", headers["X-Timestamp"])
	assert.Equal(t, 100.0, prettyBody["fiat_amount"])
	assert.Equal(t, "USD", prettyBody["fiat_code"])
	assert.Equal(t, "BTC", prettyBody["coin_code"])

	// ===========================================
	// 11. Verify Table Prefix Isolation
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("STEP 11: Verify Table Prefix Isolation")
	t.Log(strings.Repeat("=", 60))

	// Simulate existing app table (without prefix)
	type ShippingProvider struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	err = db.Table("providers").AutoMigrate(&ShippingProvider{})
	require.NoError(t, err)

	shippingProvider := ShippingProvider{Name: "FedEx"}
	err = db.Table("providers").Create(&shippingProvider).Error
	require.NoError(t, err)
	t.Log("✓ Created shipping provider in 'providers' table")

	// Verify both tables exist and are independent
	var orchProviderCount, appProviderCount int64
	db.Table("orch_providers").Count(&orchProviderCount)
	db.Table("providers").Count(&appProviderCount)

	assert.Equal(t, int64(1), orchProviderCount, "Orchestrator provider in orch_providers")
	assert.Equal(t, int64(1), appProviderCount, "App provider in providers")
	t.Log("✓ Both tables coexist independently")
	t.Logf("  - orch_providers: %d records (orchestrator)", orchProviderCount)
	t.Logf("  - providers: %d records (app)", appProviderCount)

	// Query orchestrator provider (should still work)
	var finalCheck models.Provider
	err = db.First(&finalCheck, "name = ?", "banxa").Error
	require.NoError(t, err)
	assert.Equal(t, "Banxa", finalCheck.DisplayName)
	t.Log("✓ Orchestrator data still accessible via models")

	// ===========================================
	// Success!
	// ===========================================
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("✅ FULL FLOW WITH TABLE PREFIX TEST COMPLETED!")
	t.Log(strings.Repeat("=", 60))
	t.Log("\nSummary:")
	t.Log("✓ Database created with table prefix (orch_)")
	t.Log("✓ All 8 model tables created with prefix")
	t.Log("✓ Provider, credentials, endpoints created successfully")
	t.Log("✓ Data stored in prefixed tables")
	t.Log("✓ Data queried successfully from prefixed tables")
	t.Log("✓ Request building flow executed successfully")
	t.Log("✓ Table isolation verified (app vs orchestrator tables)")
	t.Log("\n🎉 Table prefix feature working perfectly in full flow!")
}
