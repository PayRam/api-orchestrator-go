# Helper API Documentation

The Helper API provides a simple, straightforward way to configure your API orchestrator from external Go projects. This API allows you to add providers, credentials, endpoints, parameters, header rules, strategies, and response mappings without needing to work with internal packages.

## Table of Contents

- [Getting Started](#getting-started)
- [Provider Management](#provider-management)
- [Credential Management](#credential-management)
- [Endpoint Management](#endpoint-management)
- [Parameter Management](#parameter-management)
- [Header Rule Management](#header-rule-management)
- [Strategy Management](#strategy-management)
- [Response Mapping Management](#response-mapping-management)
- [Bulk Setup](#bulk-setup)
- [Complete Examples](#complete-examples)

## Getting Started

### Prerequisites

1. **Set Encryption Key** (Required for credentials)
```go
import "github.com/PayRam/api-orchestrator-go/internal/models"

// Must be 32 bytes for AES-256
encryptionKey := "my-32-byte-encryption-key-here!!"
if err := models.SetEncryptionKeyFromString(encryptionKey); err != nil {
    log.Fatalf("Failed to set encryption key: %v", err)
}
```

2. **Initialize Database**
```go
import (
    "github.com/PayRam/api-orchestrator-go/config"
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
)

dbConfig := &config.DatabaseConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "postgres",
    Database: "orchestrator_db",
    SSLMode:  "disable",
}

conn, err := config.NewConnection(dbConfig)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}

// Run migrations
if err := conn.AutoMigrate(); err != nil {
    log.Fatalf("Failed to migrate: %v", err)
}
```

3. **Create Orchestrator and Get Helper**
```go
orch, err := orchestrator.New(orchestrator.Config{
    DB:     conn.DB,
    Logger: logger,
})
if err != nil {
    log.Fatalf("Failed to create orchestrator: %v", err)
}

helper := orch.Helper()
```

## Provider Management

### Add Provider

```go
err := helper.AddProvider(
    "banxa",                        // id
    "Banxa",                        // name
    "Banxa Payment Provider",       // displayName
    "https://api.banxa.com",        // baseURL
    true,                           // isActive
)
```

### Get Provider

```go
provider, err := helper.GetProvider("banxa")
if err != nil {
    log.Printf("Provider not found: %v", err)
}
fmt.Printf("Provider: %s - %s\n", provider.Name, provider.BaseURL)
```

### List Providers

```go
providers, err := helper.ListProviders()
if err != nil {
    log.Printf("Failed to list providers: %v", err)
}
for _, p := range providers {
    fmt.Printf("  - %s (%s)\n", p.DisplayName, p.BaseURL)
}
```

## Credential Management

### Add Credential

Credentials are automatically encrypted using AES-256-GCM.

```go
err := helper.AddCredential(
    "banxa",                // providerID
    "api_key",              // key
    "your-secret-key",      // value (will be encrypted)
    true,                   // required
)
```

### Get Credentials

Returns decrypted credentials as a map.

```go
creds, err := helper.GetCredentials("banxa")
if err != nil {
    log.Printf("Failed to get credentials: %v", err)
}

apiKey := creds["api_key"]
apiSecret := creds["api_secret"]
```

## Endpoint Management

### Add Endpoint

```go
err := helper.AddEndpoint(
    "endpoint-banxa-create_widget",  // id
    "banxa",                         // providerID
    "create_widget_url",             // name
    "POST",                          // method
    "/api/orders",                   // path
)
```

### Get Endpoint

```go
endpoint, err := helper.GetEndpoint("endpoint-banxa-create_widget")
if err != nil {
    log.Printf("Endpoint not found: %v", err)
}
fmt.Printf("Endpoint: %s %s\n", endpoint.Method, endpoint.Path)
```

### List Endpoints

```go
endpoints, err := helper.ListEndpoints("banxa")
if err != nil {
    log.Printf("Failed to list endpoints: %v", err)
}
for _, ep := range endpoints {
    fmt.Printf("  - %s: %s %s\n", ep.Name, ep.Method, ep.Path)
}
```

## Parameter Management

### Add Parameter

```go
err := helper.AddParam(
    "endpoint-banxa-create_widget",  // endpointID
    "fiat_code",                     // paramName
    "body",                          // paramLocation: "body", "query", "path", "header"
    "string",                        // paramType: "string", "number", "integer", "boolean", "object", "array"
    true,                            // required
    "USD",                           // defaultValue
)
```

### Get Parameters

```go
params, err := helper.GetParams("endpoint-banxa-create_widget")
if err != nil {
    log.Printf("Failed to get params: %v", err)
}
for _, p := range params {
    fmt.Printf("  - %s: %s in %s\n", p.ParamName, p.ParamType, p.ParamLocation)
}
```

## Header Rule Management

### Add Header Rule

Header rules support various value expressions:
- `static:value` - Static value
- `credential:KEY` - Value from credential
- `template:...` - Template with placeholders
- `strategy:key` - Value from strategy
- `param:name` - Value from input parameter

```go
// Static header
err := helper.AddHeaderRule(
    "banxa",                           // providerID
    "Content-Type",                    // headerName
    "static:application/json",         // valueExpression
    1,                                 // priority
)

// Credential-based header
err = helper.AddHeaderRule(
    "banxa",
    "X-API-Key",
    "credential:api_key",
    2,
)

// Strategy-based header (e.g., HMAC signature)
err = helper.AddHeaderRule(
    "banxa",
    "Authorization",
    "strategy:banxa_hmac",
    3,
)
```

### Get Header Rules

```go
rules, err := helper.GetHeaderRules("banxa")
if err != nil {
    log.Printf("Failed to get header rules: %v", err)
}
for _, rule := range rules {
    fmt.Printf("  - %s: %s\n", rule.HeaderName, rule.ValueExpression)
}
```

## Strategy Management

### Add Strategy

Strategies handle complex authentication like HMAC, signatures, JWT, etc.

```go
// HMAC Strategy
hmacConfig := map[string]interface{}{
    "algorithm":     "sha256",
    "key_from":      "credential:api_key",
    "secret_from":   "credential:api_secret",
    "include_body":  true,
    "include_query": true,
}

err := helper.AddStrategy(
    "strategy-banxa-hmac",   // id
    "banxa_hmac",            // name
    "HMAC",                  // strategyType
    hmacConfig,              // config
)

// Basic Auth Strategy
basicConfig := map[string]interface{}{
    "username_from": "credential:api_key",
    "password_from": "credential:api_secret",
}

err = helper.AddStrategy(
    "strategy-transak-basic",
    "transak_basic",
    "BASIC_AUTH",
    basicConfig,
)
```

### Get Strategy

```go
strategy, err := helper.GetStrategy("banxa_hmac")
if err != nil {
    log.Printf("Strategy not found: %v", err)
}
fmt.Printf("Strategy: %s (Type: %s)\n", strategy.Name, strategy.StrategyType)
```

### List Strategies

```go
strategies, err := helper.ListStrategies()
if err != nil {
    log.Printf("Failed to list strategies: %v", err)
}
for _, s := range strategies {
    fmt.Printf("  - %s: %s\n", s.Name, s.StrategyType)
}
```

## Response Mapping Management

### Add Response Mapping

Response mappings extract and transform data from API responses.

```go
err := helper.AddResponseMapping(
    "banxa",                  // providerID
    "create_widget_url",      // action
    "order_id",               // targetField (field name in unified response)
    "order.id",               // sourceJSONPath (JSONPath to extract from response)
    true,                     // required
    "",                       // transform (optional: "to_upper", "to_lower", "trim", etc.)
    "",                       // defaultValue
    1,                        // priority
)

// With transformation
err = helper.AddResponseMapping(
    "banxa",
    "create_widget_url",
    "status",
    "order.status",
    true,
    "to_upper",               // Transform status to uppercase
    "",
    2,
)

// With default value
err = helper.AddResponseMapping(
    "banxa",
    "create_widget_url",
    "currency",
    "order.currency",
    false,
    "",
    "USD",                    // Default to USD if not present
    3,
)
```

### Available Transforms

- `to_upper` - Convert to uppercase
- `to_lower` - Convert to lowercase
- `trim` - Trim whitespace
- `trim_space` - Trim leading/trailing spaces
- `bool` - Convert to boolean
- `int` - Convert to integer
- `float` - Convert to float
- `string` - Convert to string
- `json_encode` - Encode as JSON
- `json_decode` - Decode JSON
- `url_encode` - URL encode
- `url_decode` - URL decode
- `base64_encode` - Base64 encode
- `base64_decode` - Base64 decode
- `timestamp` - Convert to Unix timestamp
- `iso8601` - Convert to ISO 8601 format

### Get Response Mappings

```go
mappings, err := helper.GetResponseMappings("banxa", "create_widget_url")
if err != nil {
    log.Printf("Failed to get mappings: %v", err)
}
for _, m := range mappings {
    fmt.Printf("  - %s: %s\n", m.TargetField, m.SourceJSONPath)
}
```

## Bulk Setup

### SetupProvider

For convenience, you can set up a complete provider in one call:

```go
setup := orchestrator.ProviderSetup{
    ID:          "banxa",
    Name:        "Banxa",
    DisplayName: "Banxa Payment Provider",
    BaseURL:     "https://api.banxa.com",
    
    Credentials: []orchestrator.CredentialSetup{
        {Key: "api_key", Value: "your-api-key", Required: true},
        {Key: "api_secret", Value: "your-api-secret", Required: true},
    },
    
    Endpoints: []orchestrator.EndpointSetup{
        {
            Name:   "create_widget_url",
            Method: "POST",
            Path:   "/api/orders",
            Params: []orchestrator.ParamSetup{
                {Name: "account_reference", Location: "body", Type: "string", Required: true},
                {Name: "fiat_code", Location: "body", Type: "string", Required: true},
                {Name: "coin_code", Location: "body", Type: "string", Required: true},
                {Name: "fiat_amount", Location: "body", Type: "number", Required: false},
            },
        },
    },
    
    HeaderRules: []orchestrator.HeaderRuleSetup{
        {HeaderName: "Content-Type", ValueExpression: "static:application/json"},
        {HeaderName: "Authorization", ValueExpression: "strategy:banxa_hmac"},
    },
    
    Strategies: []orchestrator.StrategySetup{
        {
            ID:   "strategy-banxa-hmac",
            Name: "banxa_hmac",
            Type: "HMAC",
            Config: map[string]interface{}{
                "algorithm":     "sha256",
                "key_from":      "credential:api_key",
                "secret_from":   "credential:api_secret",
                "include_body":  true,
                "include_query": true,
            },
        },
    },
    
    ResponseMappings: []orchestrator.ResponseMappingSetup{
        {Action: "create_widget_url", TargetField: "order_id", SourceJSONPath: "order.id", Required: true, Priority: 1},
        {Action: "create_widget_url", TargetField: "widget_url", SourceJSONPath: "order.checkout_url", Required: true, Priority: 2},
        {Action: "create_widget_url", TargetField: "status", SourceJSONPath: "order.status", Required: true, Transform: "to_upper", Priority: 3},
    },
}

if err := helper.SetupProvider(setup); err != nil {
    log.Fatalf("Failed to setup provider: %v", err)
}
```

## Complete Examples

### Example 1: Simple Provider Setup

```go
package main

import (
    "log"
    "github.com/PayRam/api-orchestrator-go/config"
    "github.com/PayRam/api-orchestrator-go/internal/models"
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "go.uber.org/zap"
)

func main() {
    // Set encryption key
    models.SetEncryptionKeyFromString("my-32-byte-encryption-key-here!!")
    
    // Connect to database
    dbConfig := &config.DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "postgres",
        Database: "orchestrator_db",
        SSLMode:  "disable",
    }
    conn, _ := config.NewConnection(dbConfig)
    conn.AutoMigrate()
    
    // Create orchestrator
    logger, _ := zap.NewDevelopment()
    orch, _ := orchestrator.New(orchestrator.Config{
        DB:     conn.DB,
        Logger: logger,
    })
    
    helper := orch.Helper()
    
    // Add provider
    helper.AddProvider("banxa", "Banxa", "Banxa Payment", "https://api.banxa.com", true)
    
    // Add credentials
    helper.AddCredential("banxa", "api_key", "your-key", true)
    helper.AddCredential("banxa", "api_secret", "your-secret", true)
    
    // Add endpoint
    helper.AddEndpoint("ep-banxa-widget", "banxa", "create_widget_url", "POST", "/api/orders")
    
    // Add parameters
    helper.AddParam("ep-banxa-widget", "fiat_code", "body", "string", true, "USD")
    helper.AddParam("ep-banxa-widget", "coin_code", "body", "string", true, "BTC")
    
    // Add headers
    helper.AddHeaderRule("banxa", "Content-Type", "static:application/json", 1)
    
    // Use it!
    response, err := orch.Call("banxa", "create_widget_url", map[string]interface{}{
        "fiat_code": "USD",
        "coin_code": "BTC",
    })
    
    if err != nil {
        log.Printf("Error: %v", err)
    } else {
        log.Printf("Success: %+v", response.Data)
    }
}
```

### Example 2: Complete Setup with Bulk API

```go
setup := orchestrator.ProviderSetup{
    ID:          "transak",
    Name:        "Transak",
    DisplayName: "Transak Payment Provider",
    BaseURL:     "https://api.transak.com",
    
    Credentials: []orchestrator.CredentialSetup{
        {Key: "api_key", Value: "your-transak-api-key", Required: true},
    },
    
    Endpoints: []orchestrator.EndpointSetup{
        {
            Name:   "create_order",
            Method: "POST",
            Path:   "/api/v2/orders",
            Params: []orchestrator.ParamSetup{
                {Name: "walletAddress", Location: "body", Type: "string", Required: true},
                {Name: "cryptoCurrency", Location: "body", Type: "string", Required: true},
                {Name: "fiatCurrency", Location: "body", Type: "string", Required: true, DefaultValue: "USD"},
                {Name: "network", Location: "body", Type: "string", Required: false, DefaultValue: "ethereum"},
            },
        },
    },
    
    HeaderRules: []orchestrator.HeaderRuleSetup{
        {HeaderName: "Content-Type", ValueExpression: "static:application/json"},
        {HeaderName: "api-key", ValueExpression: "credential:api_key"},
    },
    
    ResponseMappings: []orchestrator.ResponseMappingSetup{
        {Action: "create_order", TargetField: "order_id", SourceJSONPath: "response.id", Required: true, Priority: 1},
        {Action: "create_order", TargetField: "status", SourceJSONPath: "response.status", Required: true, Transform: "to_lower", Priority: 2},
    },
}

helper.SetupProvider(setup)
```

## Best Practices

1. **Always set encryption key first** before working with credentials
2. **Use meaningful IDs** like `"endpoint-provider-action"` for endpoints
3. **Set priorities** for header rules and response mappings to control order
4. **Use bulk setup** for complete provider configurations
5. **Test with simple calls** before adding complex strategies
6. **Check errors** - all methods return errors that should be handled
7. **Use transactions** if you need atomic multi-provider setup

## Error Handling

All Helper methods return errors. Always check them:

```go
if err := helper.AddProvider(...); err != nil {
    log.Printf("Failed to add provider: %v", err)
    // Handle error appropriately
}
```

## Thread Safety

The Helper API is thread-safe. You can safely use it from multiple goroutines.

## Performance Tips

- Use `SetupProvider` for bulk operations instead of many individual calls
- Batch related operations in a single database transaction if needed
- Cache provider/endpoint lookups if calling frequently

## See Also

- [Full Example](../examples/setup_provider.go)
- [Main README](../README.md)
- [Response Normalization Guide](./RESPONSE_NORMALIZATION.md)
