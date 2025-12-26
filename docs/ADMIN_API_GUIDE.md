# AdminAPI Usage Guide

The AdminAPI provides a clean, type-safe interface for configuring the API Orchestrator from external Go projects. This allows you to programmatically manage providers, credentials, endpoints, headers, strategies, and response mappings without accessing internal packages.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Provider Management](#provider-management)
3. [Credential Management](#credential-management)
4. [Endpoint Management](#endpoint-management)
5. [Header Rule Management](#header-rule-management)
6. [Strategy Management](#strategy-management)
7. [Response Mapping Management](#response-mapping-management)
8. [Complete Example](#complete-example)

## Getting Started

### 1. Initialize the Orchestrator

```go
package main

import (
    "log"
    
    "github.com/PayRam/api-orchestrator-go/config"
    "github.com/PayRam/api-orchestrator-go/internal/models"
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
)

func main() {
    // Initialize database
    db, err := config.InitDatabase("postgres", 
        "host=localhost user=postgres password=postgres dbname=orchestrator port=5432 sslmode=disable")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // Run migrations
    if err := config.RunMigrations(db); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // REQUIRED: Set encryption key for credentials
    if err := models.SetEncryptionKeyFromString("your-32-character-encryption-key!!"); err != nil {
        log.Fatalf("Failed to set encryption key: %v", err)
    }

    // Create orchestrator instance
    orch, err := orchestrator.New(orchestrator.Config{
        DB:     db,
        Logger: nil, // nil = default production logger
    })
    if err != nil {
        log.Fatalf("Failed to create orchestrator: %v", err)
    }

    // Get AdminAPI
    admin := orch.Admin()

    // Now you can use admin to configure providers, credentials, etc.
}
```

## Provider Management

### Create Provider

```go
import "github.com/google/uuid"

err := admin.CreateProvider(orchestrator.ProviderConfig{
    ID:          uuid.New().String(),
    Name:        "banxa",
    DisplayName: "Banxa",
    BaseURL:     "https://api.banxa.com",
    IsActive:    true,
})
```

### Get Provider by ID

```go
provider, err := admin.GetProvider(providerID)
if err != nil {
    log.Printf("Failed to get provider: %v", err)
}
fmt.Printf("Provider: %s (%s)\n", provider.DisplayName, provider.Name)
```

### Get Provider by Name

```go
provider, err := admin.GetProviderByName("banxa")
if err != nil {
    log.Printf("Failed to get provider: %v", err)
}
```

### List All Providers

```go
providers, err := admin.ListProviders()
if err != nil {
    log.Printf("Failed to list providers: %v", err)
}
for _, p := range providers {
    fmt.Printf("- %s (%s) - Active: %v\n", p.DisplayName, p.Name, p.IsActive)
}
```

### List Active Providers Only

```go
activeProviders, err := admin.ListActiveProviders()
```

### Update Provider

```go
err := admin.UpdateProvider(orchestrator.ProviderConfig{
    ID:          providerID,
    Name:        "banxa",
    DisplayName: "Banxa Updated",
    BaseURL:     "https://api.banxa.com/v2",
    IsActive:    true,
})
```

### Delete Provider (Soft Delete)

```go
err := admin.DeleteProvider(providerID)
```

## Credential Management

**Important**: Credentials are automatically encrypted using AES-256-GCM before storage.

### Create Credential

```go
err := admin.CreateCredential(orchestrator.CredentialConfig{
    ProviderID:  providerID,
    Key:         "API_KEY",
    Value:       "your-api-key-here",          // Plain text (will be encrypted)
    Description: "API Key for authentication",
    Required:    true,
})
```

### Get All Credentials for a Provider (Decrypted)

```go
credentials, err := admin.GetCredentials(providerID)
if err != nil {
    log.Printf("Failed to get credentials: %v", err)
}

// credentials is a map[string]string
for key, value := range credentials {
    fmt.Printf("%s: %s\n", key, value)
}
```

### Get Single Credential Value (Decrypted)

```go
apiKey, err := admin.GetCredentialValue(providerID, "API_KEY")
if err != nil {
    log.Printf("Failed to get credential: %v", err)
}
fmt.Printf("API Key: %s\n", apiKey)
```

### Update Credential Value

```go
err := admin.UpdateCredentialValue(providerID, "API_KEY", "new-api-key-value")
```

### Delete Credential

```go
err := admin.DeleteCredentialByKey(providerID, "API_KEY")
```

## Endpoint Management

### Create Endpoint

```go
err := admin.CreateEndpoint(orchestrator.EndpointConfig{
    ID:          uuid.New().String(),
    ProviderID:  providerID,
    Name:        "create_widget_url",
    Method:      "POST",                    // HTTP method
    Path:        "/api/orders",             // URL path
    BaseURL:     "",                        // Optional: override provider base URL
    Description: "Create order and get widget URL",
})
```

### Get Endpoint by ID

```go
endpoint, err := admin.GetEndpoint(endpointID)
```

### Get Endpoint by Provider and Name

```go
endpoint, err := admin.GetEndpointByProviderAndName("banxa", "create_widget_url")
```

### List Endpoints for a Provider

```go
endpoints, err := admin.ListEndpointsByProvider(providerID)
for _, e := range endpoints {
    fmt.Printf("- %s: %s %s\n", e.Name, e.Method, e.Path)
}
```

### Update Endpoint

```go
err := admin.UpdateEndpoint(orchestrator.EndpointConfig{
    ID:          endpointID,
    ProviderID:  providerID,
    Name:        "create_widget_url",
    Method:      "POST",
    Path:        "/api/v2/orders",
    Description: "Updated endpoint",
})
```

### Delete Endpoint (Soft Delete)

```go
err := admin.DeleteEndpoint(endpointID)
```

## Header Rule Management

Header rules define how HTTP headers are dynamically constructed for each request.

### Create Header Rule

```go
err := admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              uuid.New().String(),
    ProviderID:      providerID,
    HeaderName:      "Authorization",
    ValueExpression: "Bearer ${CREDENTIAL.API_KEY}",  // Expression syntax
    Priority:        1,                               // Lower = evaluated first
})
```

#### Value Expression Examples

```go
// Static value
ValueExpression: "application/json"

// Credential reference
ValueExpression: "Bearer ${CREDENTIAL.API_KEY}"

// Strategy reference (for computed values like HMAC signatures)
ValueExpression: "${STRATEGY.banxa_signature}"

// Combined
ValueExpression: "Bearer ${CREDENTIAL.API_KEY}:${CREDENTIAL.API_SECRET}"
```

### Get Header Rule

```go
rule, err := admin.GetHeaderRule(headerRuleID)
```

### List Header Rules for a Provider

```go
rules, err := admin.ListHeaderRulesByProvider(providerID)
for _, r := range rules {
    fmt.Printf("- %s: %s (Priority: %d)\n", r.HeaderName, r.ValueExpression, r.Priority)
}
```

### Update Header Rule

```go
err := admin.UpdateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              headerRuleID,
    ProviderID:      providerID,
    HeaderName:      "Authorization",
    ValueExpression: "Bearer ${CREDENTIAL.NEW_API_KEY}",
    Priority:        1,
})
```

### Delete Header Rule

```go
err := admin.DeleteHeaderRule(headerRuleID)
```

## Strategy Management

Strategies are reusable components for dynamic request building (signatures, tokens, etc.).

### Create Strategy

```go
err := admin.CreateStrategy(orchestrator.StrategyConfig{
    ID:           uuid.New().String(),
    Name:         "banxa_hmac_signature",
    StrategyType: "HMAC",
    Config: map[string]interface{}{
        "algorithm":      "SHA256",
        "encoding":       "hex",
        "secret_key_ref": "CREDENTIAL.API_SECRET",
        "include_fields": []string{"timestamp", "nonce", "body"},
    },
})
```

### Strategy Types

- **HMAC**: HMAC-based signature generation
- **JWT**: JSON Web Token generation
- **SIGNATURE**: Custom signature algorithms
- **TOKEN_REFRESH**: OAuth token refresh
- **PAYLOAD**: Dynamic payload generation

### Get Strategy by ID

```go
strategy, err := admin.GetStrategy(strategyID)
```

### Get Strategy by Name

```go
strategy, err := admin.GetStrategyByName("banxa_hmac_signature")
```

### List All Strategies

```go
strategies, err := admin.ListStrategies()
for _, s := range strategies {
    fmt.Printf("- %s (%s)\n", s.Name, s.StrategyType)
}
```

### List Strategies by Type

```go
hmacStrategies, err := admin.ListStrategiesByType("HMAC")
```

### Update Strategy

```go
err := admin.UpdateStrategy(orchestrator.StrategyConfig{
    ID:           strategyID,
    Name:         "banxa_hmac_signature",
    StrategyType: "HMAC",
    Config: map[string]interface{}{
        "algorithm": "SHA512", // Updated algorithm
        "encoding":  "base64",
    },
})
```

### Delete Strategy (Soft Delete)

```go
err := admin.DeleteStrategy(strategyID)
```

## Response Mapping Management

Response mappings normalize provider-specific responses into a unified format using JSONPath.

### Create Response Mapping

```go
err := admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             uuid.New().String(),
    ProviderID:     providerID,
    Action:         "create_widget_url",
    SourceJSONPath: "$.data.order.id",           // JSONPath to extract value
    TargetField:    "order_id",                  // Unified field name
    Transform:      "",                          // Optional transformation
    DefaultValue:   "",                          // Fallback if path not found
    IsRequired:     true,                        // Error if not found
    Priority:       1,                           // Evaluation order
    Description:    "Extract order ID from response",
})
```

### Transform Functions

Available transform functions:

```go
// String transformations
"uppercase", "lowercase", "trim"

// Type conversions
"to_string", "to_int", "to_float", "to_bool"

// Numeric transformations
"to_cents", "from_cents"

// Date/Time
"parse_date", "parse_timestamp"

// Encoding
"base64_decode", "json_parse"
```

### Example with Transform

```go
err := admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             uuid.New().String(),
    ProviderID:     providerID,
    Action:         "create_widget_url",
    SourceJSONPath: "$.data.order.amount",
    TargetField:    "amount_cents",
    Transform:      "to_cents",         // Converts 99.99 → 9999
    IsRequired:     true,
    Priority:       2,
})
```

### Get Response Mapping

```go
mapping, err := admin.GetResponseMapping(mappingID)
```

### List Response Mappings by Provider and Action

```go
mappings, err := admin.ListResponseMappingsByProviderAndAction(providerID, "create_widget_url")
for _, m := range mappings {
    fmt.Printf("- %s ← %s (Transform: %s, Priority: %d)\n",
        m.TargetField, m.SourceJSONPath, m.Transform, m.Priority)
}
```

### Update Response Mapping

```go
err := admin.UpdateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             mappingID,
    ProviderID:     providerID,
    Action:         "create_widget_url",
    SourceJSONPath: "$.data.order.checkout_url",
    TargetField:    "widget_url",
    Transform:      "",
    IsRequired:     true,
    Priority:       1,
})
```

### Delete Response Mapping

```go
err := admin.DeleteResponseMapping(mappingID)
```

## Complete Example

Here's a complete example showing how to configure a provider from scratch:

```go
package main

import (
    "fmt"
    "log"

    "github.com/PayRam/api-orchestrator-go/config"
    "github.com/PayRam/api-orchestrator-go/internal/models"
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "github.com/google/uuid"
)

func main() {
    // 1. Initialize database
    db, err := config.InitDatabase("postgres", 
        "host=localhost user=postgres password=postgres dbname=orchestrator port=5432 sslmode=disable")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // 2. Run migrations
    if err := config.RunMigrations(db); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // 3. Set encryption key (REQUIRED)
    if err := models.SetEncryptionKeyFromString("your-32-character-encryption-key!!"); err != nil {
        log.Fatalf("Failed to set encryption key: %v", err)
    }

    // 4. Create orchestrator
    orch, err := orchestrator.New(orchestrator.Config{
        DB:     db,
        Logger: nil,
    })
    if err != nil {
        log.Fatalf("Failed to create orchestrator: %v", err)
    }

    // 5. Get AdminAPI
    admin := orch.Admin()

    // 6. Configure Banxa Provider
    providerID := uuid.New().String()
    
    // Create provider
    err = admin.CreateProvider(orchestrator.ProviderConfig{
        ID:          providerID,
        Name:        "banxa",
        DisplayName: "Banxa",
        BaseURL:     "https://api.banxa.com",
        IsActive:    true,
    })
    if err != nil {
        log.Fatalf("Failed to create provider: %v", err)
    }

    // Add credentials
    admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID:  providerID,
        Key:         "API_KEY",
        Value:       "your-api-key",
        Description: "Banxa API Key",
        Required:    true,
    })

    admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID:  providerID,
        Key:         "API_SECRET",
        Value:       "your-api-secret",
        Description: "Banxa API Secret",
        Required:    true,
    })

    // Create endpoint
    endpointID := uuid.New().String()
    admin.CreateEndpoint(orchestrator.EndpointConfig{
        ID:          endpointID,
        ProviderID:  providerID,
        Name:        "create_widget_url",
        Method:      "POST",
        Path:        "/api/orders",
        Description: "Create widget URL",
    })

    // Add header rules
    admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
        ID:              uuid.New().String(),
        ProviderID:      providerID,
        HeaderName:      "Authorization",
        ValueExpression: "Bearer ${CREDENTIAL.API_KEY}",
        Priority:        1,
    })

    admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
        ID:              uuid.New().String(),
        ProviderID:      providerID,
        HeaderName:      "Content-Type",
        ValueExpression: "application/json",
        Priority:        2,
    })

    // Create strategy
    admin.CreateStrategy(orchestrator.StrategyConfig{
        ID:           uuid.New().String(),
        Name:         "banxa_hmac",
        StrategyType: "HMAC",
        Config: map[string]interface{}{
            "algorithm":      "SHA256",
            "secret_key_ref": "CREDENTIAL.API_SECRET",
        },
    })

    // Add response mappings
    mappings := []orchestrator.ResponseMappingConfig{
        {
            ID:             uuid.New().String(),
            ProviderID:     providerID,
            Action:         "create_widget_url",
            SourceJSONPath: "$.data.order.id",
            TargetField:    "order_id",
            IsRequired:     true,
            Priority:       1,
        },
        {
            ID:             uuid.New().String(),
            ProviderID:     providerID,
            Action:         "create_widget_url",
            SourceJSONPath: "$.data.order.checkout_url",
            TargetField:    "widget_url",
            IsRequired:     true,
            Priority:       2,
        },
    }

    for _, mapping := range mappings {
        admin.CreateResponseMapping(mapping)
    }

    fmt.Println("✓ Provider configured successfully!")

    // 7. Now make an API call
    response, err := orch.Call("banxa", "create_widget_url", map[string]interface{}{
        "account_reference": "user123",
        "source":            "USD",
        "target":            "BTC",
        "source_amount":     "100",
        "return_url_on_success": "https://example.com/success",
    })

    if err != nil {
        log.Fatalf("API call failed: %v", err)
    }

    fmt.Printf("Order ID: %s\n", response.Data["order_id"])
    fmt.Printf("Widget URL: %s\n", response.Data["widget_url"])
}
```

## Best Practices

### 1. Encryption Key Management

```go
// Option 1: Load from environment variable
key := os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY")
if err := models.SetEncryptionKeyFromString(key); err != nil {
    log.Fatal(err)
}

// Option 2: Load from secure vault/secrets manager
// key := getKeyFromVault()
// models.SetEncryptionKey([]byte(key))
```

### 2. Error Handling

```go
if err := admin.CreateProvider(cfg); err != nil {
    // Check if it's a duplicate error
    if strings.Contains(err.Error(), "duplicate") {
        log.Printf("Provider already exists")
    } else {
        log.Fatalf("Failed to create provider: %v", err)
    }
}
```

### 3. Batch Operations

```go
// Create multiple providers
providers := []orchestrator.ProviderConfig{
    {ID: uuid.New().String(), Name: "banxa", DisplayName: "Banxa", BaseURL: "https://api.banxa.com", IsActive: true},
    {ID: uuid.New().String(), Name: "transak", DisplayName: "Transak", BaseURL: "https://api.transak.com", IsActive: true},
}

for _, cfg := range providers {
    if err := admin.CreateProvider(cfg); err != nil {
        log.Printf("Failed to create %s: %v", cfg.Name, err)
    }
}
```

### 4. Configuration from JSON

```go
type ProviderSetup struct {
    Provider         orchestrator.ProviderConfig            `json:"provider"`
    Credentials      []orchestrator.CredentialConfig        `json:"credentials"`
    Endpoints        []orchestrator.EndpointConfig          `json:"endpoints"`
    HeaderRules      []orchestrator.HeaderRuleConfig        `json:"header_rules"`
    ResponseMappings []orchestrator.ResponseMappingConfig   `json:"response_mappings"`
}

func loadProviderFromJSON(admin *orchestrator.AdminAPI, jsonData []byte) error {
    var setup ProviderSetup
    if err := json.Unmarshal(jsonData, &setup); err != nil {
        return err
    }

    // Create provider
    if err := admin.CreateProvider(setup.Provider); err != nil {
        return err
    }

    // Create credentials
    for _, cred := range setup.Credentials {
        if err := admin.CreateCredential(cred); err != nil {
            return err
        }
    }

    // Create endpoints
    for _, endpoint := range setup.Endpoints {
        if err := admin.CreateEndpoint(endpoint); err != nil {
            return err
        }
    }

    // ... continue for other resources
    return nil
}
```

## Troubleshooting

### "Encryption key not set"

```go
// Make sure to call this BEFORE creating the orchestrator
models.SetEncryptionKeyFromString("your-32-character-key-here!!")
```

### "Provider not found"

```go
// Check if provider exists before referencing
provider, err := admin.GetProviderByName("banxa")
if err != nil {
    log.Printf("Provider doesn't exist, creating...")
    admin.CreateProvider(/* ... */)
}
```

### "Method not allowed" in HTTP Executor

```go
// Ensure endpoint Method is one of: GET, POST, PUT, PATCH, DELETE
admin.CreateEndpoint(orchestrator.EndpointConfig{
    Method: "POST", // Must be uppercase
    // ...
})
```
