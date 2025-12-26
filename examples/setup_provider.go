package main

import (
	"fmt"
	"log"

	"github.com/PayRam/api-orchestrator-go/config"
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
	"go.uber.org/zap"
)

func main() {
	// Step 1: Set encryption key (REQUIRED before working with credentials)
	encryptionKey := "my-32-byte-encryption-key-here!!" // Must be 32 bytes for AES-256
	if err := models.SetEncryptionKeyFromString(encryptionKey); err != nil {
		log.Fatalf("Failed to set encryption key: %v", err)
	}

	// Step 2: Initialize database
	dbConfig := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Database: "orchestrator_test",
		SSLMode:  "disable",
	}

	conn, err := config.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := conn.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db := conn.DB

	// Step 3: Create orchestrator
	logger, _ := zap.NewDevelopment()
	orch, err := orchestrator.New(orchestrator.Config{
		DB:     db,
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Step 4: Get helper
	helper := orch.Helper()

	// ============================================
	// Example 1: Add provider piece by piece
	// ============================================
	fmt.Println("=== Example 1: Add provider piece by piece ===")

	// Add provider
	if err := helper.AddProvider("banxa", "Banxa", "Banxa Payment Provider", "https://api.banxa.com", true); err != nil {
		log.Printf("Error adding provider: %v", err)
	}

	// Add credentials
	if err := helper.AddCredential("banxa", "api_key", "your-api-key-here", true); err != nil {
		log.Printf("Error adding credential: %v", err)
	}
	if err := helper.AddCredential("banxa", "api_secret", "your-api-secret-here", true); err != nil {
		log.Printf("Error adding credential: %v", err)
	}

	// Add endpoint
	if err := helper.AddEndpoint("endpoint-banxa-create_widget", "banxa", "create_widget_url", "POST", "/api/orders"); err != nil {
		log.Printf("Error adding endpoint: %v", err)
	}

	// Add parameters
	if err := helper.AddParam("endpoint-banxa-create_widget", "account_reference", "body", "string", true, ""); err != nil {
		log.Printf("Error adding param: %v", err)
	}
	if err := helper.AddParam("endpoint-banxa-create_widget", "fiat_code", "body", "string", true, "USD"); err != nil {
		log.Printf("Error adding param: %v", err)
	}
	if err := helper.AddParam("endpoint-banxa-create_widget", "coin_code", "body", "string", true, "BTC"); err != nil {
		log.Printf("Error adding param: %v", err)
	}
	if err := helper.AddParam("endpoint-banxa-create_widget", "source", "body", "string", false, "web"); err != nil {
		log.Printf("Error adding param: %v", err)
	}

	// Add header rules
	if err := helper.AddHeaderRule("banxa", "Content-Type", "static:application/json", 1); err != nil {
		log.Printf("Error adding header rule: %v", err)
	}
	if err := helper.AddHeaderRule("banxa", "Authorization", "strategy:banxa_hmac", 2); err != nil {
		log.Printf("Error adding header rule: %v", err)
	}

	// Add strategy
	strategyConfig := map[string]interface{}{
		"algorithm":     "sha256",
		"key_from":      "credential:api_key",
		"secret_from":   "credential:api_secret",
		"include_body":  true,
		"include_query": true,
	}
	if err := helper.AddStrategy("strategy-banxa-hmac", "banxa_hmac", "HMAC", strategyConfig); err != nil {
		log.Printf("Error adding strategy: %v", err)
	}

	// Add response mappings
	if err := helper.AddResponseMapping("banxa", "create_widget_url", "order_id", "order.id", true, "", "", 1); err != nil {
		log.Printf("Error adding response mapping: %v", err)
	}
	if err := helper.AddResponseMapping("banxa", "create_widget_url", "widget_url", "order.checkout_url", true, "", "", 2); err != nil {
		log.Printf("Error adding response mapping: %v", err)
	}
	if err := helper.AddResponseMapping("banxa", "create_widget_url", "status", "order.status", true, "to_upper", "", 3); err != nil {
		log.Printf("Error adding response mapping: %v", err)
	}

	fmt.Println("✓ Provider added successfully!")

	// ============================================
	// Example 2: Add complete provider in one call
	// ============================================
	fmt.Println("\n=== Example 2: Add complete provider in one call ===")

	transakSetup := orchestrator.ProviderSetup{
		ID:          "transak",
		Name:        "Transak",
		DisplayName: "Transak Payment Provider",
		BaseURL:     "https://api.transak.com",
		Credentials: []orchestrator.CredentialSetup{
			{Key: "api_key", Value: "your-transak-api-key", Required: true},
			{Key: "api_secret", Value: "your-transak-secret", Required: true},
		},
		Endpoints: []orchestrator.EndpointSetup{
			{
				Name:   "create_order",
				Method: "POST",
				Path:   "/api/v2/orders",
				Params: []orchestrator.ParamSetup{
					{Name: "walletAddress", Location: "body", Type: "string", Required: true},
					{Name: "cryptoCurrency", Location: "body", Type: "string", Required: true},
					{Name: "fiatCurrency", Location: "body", Type: "string", Required: true},
					{Name: "network", Location: "body", Type: "string", Required: false, DefaultValue: "ethereum"},
				},
			},
		},
		HeaderRules: []orchestrator.HeaderRuleSetup{
			{HeaderName: "Content-Type", ValueExpression: "static:application/json"},
			{HeaderName: "api-key", ValueExpression: "credential:api_key"},
			{HeaderName: "api-secret", ValueExpression: "credential:api_secret"},
		},
		Strategies: []orchestrator.StrategySetup{
			{
				ID:   "strategy-transak-basic",
				Name: "transak_basic",
				Type: "BASIC_AUTH",
				Config: map[string]interface{}{
					"username_from": "credential:api_key",
					"password_from": "credential:api_secret",
				},
			},
		},
		ResponseMappings: []orchestrator.ResponseMappingSetup{
			{Action: "create_order", TargetField: "order_id", SourceJSONPath: "response.id", Required: true, Priority: 1},
			{Action: "create_order", TargetField: "status", SourceJSONPath: "response.status", Required: true, Transform: "to_lower", Priority: 2},
			{Action: "create_order", TargetField: "payment_url", SourceJSONPath: "response.redirectUrl", Required: false, Priority: 3},
		},
	}

	if err := helper.SetupProvider(transakSetup); err != nil {
		log.Printf("Error setting up Transak: %v", err)
	} else {
		fmt.Println("✓ Transak provider setup complete!")
	}

	// ============================================
	// Example 3: Query the data
	// ============================================
	fmt.Println("\n=== Example 3: Query the data ===")

	// List all providers
	providers, err := helper.ListProviders()
	if err != nil {
		log.Printf("Error listing providers: %v", err)
	} else {
		fmt.Printf("Found %d providers:\n", len(providers))
		for _, p := range providers {
			fmt.Printf("  - %s (%s)\n", p.DisplayName, p.BaseURL)
		}
	}

	// Get credentials for a provider
	creds, err := helper.GetCredentials("banxa")
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
	} else {
		fmt.Printf("\nBanxa credentials: %d keys\n", len(creds))
		for key := range creds {
			fmt.Printf("  - %s: ***\n", key)
		}
	}

	// Get endpoints for a provider
	endpoints, err := helper.ListEndpoints("banxa")
	if err != nil {
		log.Printf("Error listing endpoints: %v", err)
	} else {
		fmt.Printf("\nBanxa endpoints:\n")
		for _, ep := range endpoints {
			fmt.Printf("  - %s: %s %s\n", ep.Name, ep.Method, ep.Path)
		}
	}

	// Get parameters for an endpoint
	params, err := helper.GetParams("endpoint-banxa-create_widget")
	if err != nil {
		log.Printf("Error getting params: %v", err)
	} else {
		fmt.Printf("\nParameters for create_widget_url:\n")
		for _, p := range params {
			required := ""
			if p.Required {
				required = " (required)"
			}
			fmt.Printf("  - %s: %s in %s%s\n", p.ParamName, p.ParamType, p.ParamLocation, required)
		}
	}

	// Get header rules
	rules, err := helper.GetHeaderRules("banxa")
	if err != nil {
		log.Printf("Error getting header rules: %v", err)
	} else {
		fmt.Printf("\nBanxa header rules:\n")
		for _, rule := range rules {
			fmt.Printf("  - %s: %s\n", rule.HeaderName, rule.ValueExpression)
		}
	}

	// Get response mappings
	mappings, err := helper.GetResponseMappings("banxa", "create_widget_url")
	if err != nil {
		log.Printf("Error getting response mappings: %v", err)
	} else {
		fmt.Printf("\nResponse mappings for create_widget_url:\n")
		for _, m := range mappings {
			transform := ""
			if m.Transform != "" {
				transform = fmt.Sprintf(" (transform: %s)", m.Transform)
			}
			fmt.Printf("  - %s: %s%s\n", m.TargetField, m.SourceJSONPath, transform)
		}
	}

	// ============================================
	// Example 4: Now use the orchestrator!
	// ============================================
	fmt.Println("\n=== Example 4: Use the orchestrator ===")

	response, err := orch.Call("banxa", "create_widget_url", map[string]interface{}{
		"account_reference": "user123",
		"fiat_code":         "USD",
		"coin_code":         "BTC",
		"fiat_amount":       100.00,
	})

	if err != nil {
		log.Printf("API call failed: %v", err)
	} else {
		fmt.Printf("API call successful!\n")
		fmt.Printf("  Success: %v\n", response.Success)
		fmt.Printf("  Data: %+v\n", response.Data)
	}

	fmt.Println("\n✅ All examples complete!")
}
