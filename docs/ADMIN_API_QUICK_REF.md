# AdminAPI Quick Reference

Quick reference for the most common AdminAPI operations.

## Setup

```go
import (
    "github.com/PayRam/api-orchestrator-go/internal/models"
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "github.com/google/uuid"
)

// 1. Set encryption key (REQUIRED, do this once at startup)
models.SetEncryptionKeyFromString("your-32-character-encryption-key!!")

// 2. Create orchestrator
orch, err := orchestrator.New(orchestrator.Config{DB: db})

// 3. Get admin interface
admin := orch.Admin()
```

## Quick Operations

### Provider
```go
// Create
admin.CreateProvider(orchestrator.ProviderConfig{
    ID: uuid.New().String(), Name: "banxa", DisplayName: "Banxa",
    BaseURL: "https://api.banxa.com", IsActive: true,
})

// Get
provider, _ := admin.GetProviderByName("banxa")

// List active
providers, _ := admin.ListActiveProviders()
```

### Credential (Auto-encrypted)
```go
// Create
admin.CreateCredential(orchestrator.CredentialConfig{
    ProviderID: providerID, Key: "API_KEY",
    Value: "plain-text-key", Required: true,
})

// Get (auto-decrypted)
creds, _ := admin.GetCredentials(providerID)
apiKey := creds["API_KEY"]

// Update
admin.UpdateCredentialValue(providerID, "API_KEY", "new-value")
```

### Endpoint
```go
// Create
admin.CreateEndpoint(orchestrator.EndpointConfig{
    ID: uuid.New().String(), ProviderID: providerID,
    Name: "create_order", Method: "POST", Path: "/api/orders",
})

// Get
endpoint, _ := admin.GetEndpointByProviderAndName("banxa", "create_order")

// List
endpoints, _ := admin.ListEndpointsByProvider(providerID)
```

### Header Rule
```go
// Create
admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID: uuid.New().String(), ProviderID: providerID,
    HeaderName: "Authorization",
    ValueExpression: "Bearer ${CREDENTIAL.API_KEY}",
    Priority: 1,
})

// List
rules, _ := admin.ListHeaderRulesByProvider(providerID)
```

### Strategy
```go
// Create
admin.CreateStrategy(orchestrator.StrategyConfig{
    ID: uuid.New().String(), Name: "hmac_sig", StrategyType: "HMAC",
    Config: map[string]interface{}{
        "algorithm": "SHA256",
        "secret_key_ref": "CREDENTIAL.API_SECRET",
    },
})

// Get
strategy, _ := admin.GetStrategyByName("hmac_sig")

// List by type
hmacStrategies, _ := admin.ListStrategiesByType("HMAC")
```

### Response Mapping
```go
// Create
admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID: uuid.New().String(), ProviderID: providerID,
    Action: "create_order", SourceJSONPath: "$.data.id",
    TargetField: "order_id", Transform: "", IsRequired: true, Priority: 1,
})

// List
mappings, _ := admin.ListResponseMappingsByProviderAndAction(providerID, "create_order")
```

## Value Expression Syntax

```go
// Credential reference
"${CREDENTIAL.API_KEY}"
"Bearer ${CREDENTIAL.TOKEN}"

// Strategy reference
"${STRATEGY.signature}"

// Combined
"${CREDENTIAL.API_KEY}:${CREDENTIAL.API_SECRET}"

// Static value
"application/json"
```

## Transform Functions

```go
// String: "uppercase", "lowercase", "trim"
// Type: "to_string", "to_int", "to_float", "to_bool"
// Numeric: "to_cents", "from_cents"
// Date: "parse_date", "parse_timestamp"
// Encoding: "base64_decode", "json_parse"
```

## Complete Example

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
    // Setup
    db, _ := config.InitDatabase("postgres", "...")
    config.RunMigrations(db)
    models.SetEncryptionKeyFromString("your-32-character-key-here!!!!!")
    
    orch, _ := orchestrator.New(orchestrator.Config{DB: db})
    admin := orch.Admin()
    
    // Configure Banxa
    providerID := uuid.New().String()
    
    admin.CreateProvider(orchestrator.ProviderConfig{
        ID: providerID, Name: "banxa", DisplayName: "Banxa",
        BaseURL: "https://api.banxa.com", IsActive: true,
    })
    
    admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID: providerID, Key: "API_KEY",
        Value: "your-api-key", Required: true,
    })
    
    admin.CreateEndpoint(orchestrator.EndpointConfig{
        ID: uuid.New().String(), ProviderID: providerID,
        Name: "create_widget_url", Method: "POST", Path: "/api/orders",
    })
    
    admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
        ID: uuid.New().String(), ProviderID: providerID,
        HeaderName: "Authorization",
        ValueExpression: "Bearer ${CREDENTIAL.API_KEY}",
        Priority: 1,
    })
    
    admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
        ID: uuid.New().String(), ProviderID: providerID,
        Action: "create_widget_url",
        SourceJSONPath: "$.data.order.checkout_url",
        TargetField: "widget_url", IsRequired: true, Priority: 1,
    })
    
    // Make API call
    response, err := orch.Call("banxa", "create_widget_url", map[string]interface{}{
        "account_reference": "user123",
        "source": "USD",
        "target": "BTC",
        "source_amount": "100",
    })
    
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    fmt.Printf("Widget URL: %s\n", response.Data["widget_url"])
}
```

## Error Handling

```go
if err := admin.CreateProvider(cfg); err != nil {
    // Handle error
    log.Printf("Failed: %v", err)
}

// Check existence before creating
if _, err := admin.GetProviderByName("banxa"); err != nil {
    // Provider doesn't exist, create it
    admin.CreateProvider(cfg)
}
```

## See Also

- **Full Guide**: `docs/ADMIN_API_GUIDE.md` (650+ lines with examples)
- **Summary**: `docs/ADMIN_API_SUMMARY.md` (architecture and design)
- **Integration**: `docs/INTEGRATION_GUIDE.md` (how to integrate into your project)
