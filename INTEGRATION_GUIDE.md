# API Orchestrator Go - Integration Guide

A powerful Go library for orchestrating multiple third-party payment provider APIs (Banxa, Transak, MoonPay, etc.) through a unified interface. This library enables dynamic API integration without writing provider-specific code.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Detailed Integration Steps](#detailed-integration-steps)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## Overview

The API Orchestrator Go library provides:

- **Unified API Interface** - Call any provider through a single consistent API
- **Dynamic Configuration** - Add/modify providers via database without code changes
- **Automatic Encryption** - Credentials encrypted at rest using AES-256-GCM
- **Response Normalization** - Consistent response format across all providers
- **Built-in Security** - HMAC signatures, header injection, credential management
- **Debug Support** - Generate cURL commands without executing requests
- **Production Ready** - Comprehensive logging, error handling, retry logic

### Architecture

```
Your Application
       ↓
API Orchestrator Library
       ↓
┌─────────────────────────────────────┐
│  Provider Configurations (Database) │
│  • Providers                        │
│  • Credentials (Encrypted)          │
│  • Endpoints                        │
│  • Header Rules                     │
│  • Strategies                       │
│  • Response Mappings                │
└─────────────────────────────────────┘
       ↓
┌─────────────────────────────────────┐
│  External Payment Provider APIs     │
│  • Banxa                            │
│  • Transak                          │
│  • MoonPay                          │
│  • ... (any HTTP API)               │
└─────────────────────────────────────┘
```

---

## Features

### ✅ Provider Management
- Add multiple payment providers dynamically
- Enable/disable providers without code changes
- Provider-specific base URLs and configurations

### ✅ Credential Management
- Automatic AES-256-GCM encryption at rest
- Secure credential storage and retrieval
- Support for multiple credentials per provider (API keys, secrets, tokens)

### ✅ Table Prefix Support
- **NEW!** Avoid table name conflicts with existing tables
- Configure custom prefix for all orchestrator tables
- E.g., use `orch_` prefix to create `orch_providers` instead of `providers`
- Seamless coexistence with existing application tables
- See [Table Prefix Documentation](docs/TABLE_PREFIX.md)

### ✅ Endpoint Configuration
- Define API endpoints with method, path, parameters
- Request schema validation
- Dynamic parameter mapping

### ✅ Header Management
- Static headers (Content-Type, etc.)
- Credential-based headers (API keys)
- Template-based headers with variable substitution
- Strategy-based headers (HMAC signatures, OAuth, etc.)

### ✅ Authentication Strategies
- HMAC signature generation
- Basic Authentication
- API Key authentication
- Custom signature algorithms
- OAuth token management

### ✅ Response Mapping
- JSONPath extraction from provider responses
- Field transformation and normalization
- Unified response format
- Raw response preservation

### ✅ Request Building
- Build cURL commands without execution (debugging)
- Get detailed request information (method, URL, headers, body)
- Execute API calls with unified response

---

## Prerequisites

- **Go**: 1.21 or higher
- **Database**: PostgreSQL (uses your existing database)
- **GORM**: For database operations
- **Zap**: For logging (optional, but recommended)

---

## Installation

### Step 1: Add Library to Your Project

```bash
go get github.com/PayRam/api-orchestrator-go
```

### Step 2: Update Dependencies

```bash
go mod tidy
```

### Step 3: Import Required Packages

```go
import (
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    orchConfig "github.com/PayRam/api-orchestrator-go/config"
)
```

---

## Quick Start

```go
package main

import (
    "log"
    "os"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    orchConfig "github.com/PayRam/api-orchestrator-go/config"
    "github.com/google/uuid"
)

func main() {
    // 1. Set encryption key (32 characters required)
    orchestrator.SetEncryptionKeyFromString("your-32-character-secret-key!")
    
    // 2. Use your existing database connection
    db := getYourExistingDB() // *gorm.DB
    
    // 3. Run migrations (creates 8 tables in your database)
    // Option A: Without table prefix (standard)
    orchConfig.AutoMigrate(db)
    
    // Option B: With table prefix (if you have table name conflicts)
    // orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
    //     TablePrefix: "orch_",
    // })
    
    // 4. Create orchestrator instance
    orch, err := orchestrator.New(orchestrator.Config{
        DB:     db,
        Logger: logger, // optional
        // TablePrefix: "orch_", // Use same prefix as migrations
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 5. Configure a provider (one-time setup)
    admin := orch.Admin()
    providerID := uuid.New().String()
    
    admin.CreateProvider(orchestrator.ProviderConfig{
        ID:          providerID,
        Name:        "banxa",
        DisplayName: "Banxa",
        BaseURL:     "https://api.banxa.com",
        IsActive:    true,
    })
    
    admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID: providerID,
        Key:        "API_KEY",
        Value:      os.Getenv("BANXA_API_KEY"),
        Required:   true,
    })
    
    admin.CreateEndpoint(orchestrator.EndpointConfig{
        ID:         uuid.New().String(),
        ProviderID: providerID,
        Name:       "create_order",
        Method:     "POST",
        Path:       "/api/orders",
    })
    
    admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
        ID:              uuid.New().String(),
        ProviderID:      providerID,
        HeaderName:      "X-API-Key",
        ValueExpression: "credential:API_KEY",
        Priority:        1,
    })
    
    // 6. Execute API call
    response, err := orch.Call("banxa", "create_order", map[string]interface{}{
        "fiat_amount": 100.00,
        "fiat_code":   "USD",
        "coin_code":   "BTC",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Success: %v, Data: %+v", response.Success, response.Data)
}
```

---

## Detailed Integration Steps

### Step 1: Set Encryption Key

**IMPORTANT**: The encryption key must be exactly 32 characters and should be set **before** creating the orchestrator instance.

```go
// Option 1: From environment variable (recommended)
encryptionKey := os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY")
if len(encryptionKey) != 32 {
    log.Fatal("ORCHESTRATOR_ENCRYPTION_KEY must be exactly 32 characters")
}
err := orchestrator.SetEncryptionKeyFromString(encryptionKey)
if err != nil {
    log.Fatal("Failed to set encryption key:", err)
}

// Option 2: From byte array
key := []byte("12345678901234567890123456789012") // 32 bytes
err := orchestrator.SetEncryptionKey(key)

// Check if key is set
if !orchestrator.IsEncryptionKeySet() {
    log.Fatal("Encryption key not set")
}
```

**Generate a secure key:**

```bash
# Linux/Mac
openssl rand -base64 24 | head -c 32

# Or use this Go code
import "crypto/rand"
import "encoding/base64"

key := make([]byte, 24)
rand.Read(key)
encoded := base64.StdEncoding.EncodeToString(key)
finalKey := encoded[:32] // Take first 32 characters
```

### Step 2: Database Setup

The library uses your **existing database** and creates its own tables.

```go
// Your existing database connection
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
if err != nil {
    log.Fatal("Database connection failed:", err)
}

// Run orchestrator migrations
err = orchConfig.AutoMigrate(db)
if err != nil {
    log.Fatal("Migrations failed:", err)
}
```

**Tables Created:**

| Table | Description |
|-------|-------------|
| `providers` | Provider configurations (name, base URL, etc.) |
| `credentials` | Encrypted API credentials |
| `endpoints` | API endpoint definitions |
| `header_rules` | Dynamic header generation rules |
| `strategies` | Authentication strategies (HMAC, signatures) |
| `request_schemas` | Request parameter schemas |
| `request_values` | Request parameter values |
| `response_mappings` | Response field mappings |

**Avoiding Table Name Conflicts:**

If your application already has tables with these names, use a table prefix:

```go
// Run migrations with prefix
err = orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
    TablePrefix: "orch_",
})
if err != nil {
    log.Fatal("Migrations failed:", err)
}
```

This creates tables with prefix: `orch_providers`, `orch_credentials`, etc.

**Recommended Prefixes:**
- `orch_` - Short and clear
- `orchestrator_` - More explicit
- `api_orch_` - For multiple orchestrators
- Your company/project prefix

See [Table Prefix Documentation](docs/TABLE_PREFIX.md) for complete guide.

### Step 3: Initialize Orchestrator

```go
// Create orchestrator instance
orch, err := orchestrator.New(orchestrator.Config{
    DB:          db,     // Required: Your *gorm.DB instance
    Logger:      logger, // Optional: Your *zap.Logger instance
    TablePrefix: "orch_", // Optional: Use if you set prefix in migrations
})
if err != nil {
    log.Fatal("Failed to create orchestrator:", err)
}

// Store globally or in application context
var globalOrch *orchestrator.Orchestrator = orch
```

### Step 4: Configure Providers

**Important Note:** This step is for **configuring providers** (Banxa, Transak, etc.), which is **separate from library integration**. You can:
- Configure providers programmatically (shown below)
- Use REST Admin API endpoints (if using orchestrator-api-server)
- Load from configuration files
- Do this as a one-time setup, not in application startup code

#### 4.1 Create Provider

```go
admin := orch.Admin()

providerID := uuid.New().String()
err := admin.CreateProvider(orchestrator.ProviderConfig{
    ID:          providerID,
    Name:        "banxa",           // Unique name
    DisplayName: "Banxa",           // Display name
    BaseURL:     "https://api.banxa.com",
    IsActive:    true,
    Description: "Banxa crypto on-ramp provider",
})
if err != nil {
    log.Printf("Failed to create provider: %v", err)
}
```

#### 4.2 Add Credentials (Auto-Encrypted)

```go
// API Key
    err = admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID:  providerID,
        Key:         "API_KEY",
        Value:       "pk_live_your_actual_banxa_key", // Automatically encrypted
        Required:    true,
        Description: "Banxa API Key",
    })// API Secret
err = admin.CreateCredential(orchestrator.CredentialConfig{
    ProviderID:  providerID,
    Key:         "API_SECRET",
    Value:       os.Getenv("BANXA_API_SECRET"), // Automatically encrypted
    Required:    true,
    Description: "Banxa API Secret for HMAC signing",
})

// Access Token (optional)
err = admin.CreateCredential(orchestrator.CredentialConfig{
    ProviderID:  providerID,
    Key:         "ACCESS_TOKEN",
    Value:       os.Getenv("BANXA_ACCESS_TOKEN"),
    Required:    false,
    Description: "Optional access token",
})
```

#### 4.3 Create Endpoints

```go
// Create Order endpoint
endpointID := uuid.New().String()
err = admin.CreateEndpoint(orchestrator.EndpointConfig{
    ID:          endpointID,
    ProviderID:  providerID,
    Name:        "create_order",
    Method:      "POST",
    Path:        "/api/orders",
    Description: "Create a crypto purchase order",
})

// Get Order Status endpoint
err = admin.CreateEndpoint(orchestrator.EndpointConfig{
    ID:          uuid.New().String(),
    ProviderID:  providerID,
    Name:        "get_order_status",
    Method:      "GET",
    Path:        "/api/orders/{order_id}",
    Description: "Get order status by ID",
})
```

#### 4.4 Add Header Rules

```go
// Static header
err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              uuid.New().String(),
    ProviderID:      providerID,
    HeaderName:      "Content-Type",
    ValueExpression: "static:application/json",
    Priority:        1,
    Description:     "Content type header",
})

// Credential-based header
err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              uuid.New().String(),
    ProviderID:      providerID,
    HeaderName:      "X-API-Key",
    ValueExpression: "credential:API_KEY",
    Priority:        2,
    Description:     "API key authentication",
})

// Template-based header
err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              uuid.New().String(),
    ProviderID:      providerID,
    HeaderName:      "Authorization",
    ValueExpression: "template:Bearer ${credential:ACCESS_TOKEN}",
    Priority:        3,
    Description:     "Bearer token authorization",
})

// Strategy-based header (HMAC signature)
err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
    ID:              uuid.New().String(),
    ProviderID:      providerID,
    HeaderName:      "X-Signature",
    ValueExpression: "strategy:banxa_hmac",
    Priority:        4,
    Description:     "HMAC signature",
})
```

**Value Expression Formats:**

| Format | Example | Description |
|--------|---------|-------------|
| `static:value` | `static:application/json` | Static string value |
| `credential:KEY` | `credential:API_KEY` | From credential storage |
| `param:name` | `param:user_id` | From request parameters |
| `template:text ${var}` | `template:Bearer ${credential:TOKEN}` | Template with variables |
| `strategy:name` | `strategy:banxa_hmac` | Execute authentication strategy |

#### 4.5 Create Authentication Strategies

```go
// HMAC Strategy
err = admin.CreateStrategy(orchestrator.StrategyConfig{
    ID:           uuid.New().String(),
    Name:         "banxa_hmac",
    StrategyType: "HMAC",
    Config: map[string]interface{}{
        "algorithm":   "SHA256",
        "encoding":    "hex",
        "secret_key":  "API_SECRET",
        "data_format": "{timestamp}{method}{path}{body}",
    },
    Description: "HMAC-SHA256 signature for Banxa",
})

// Basic Auth Strategy
err = admin.CreateStrategy(orchestrator.StrategyConfig{
    ID:           uuid.New().String(),
    Name:         "basic_auth",
    StrategyType: "BASIC_AUTH",
    Config: map[string]interface{}{
        "username_key": "API_KEY",
        "password_key": "API_SECRET",
    },
    Description: "HTTP Basic Authentication",
})
```

**Strategy Types:**

| Type | Description | Config Fields |
|------|-------------|---------------|
| `HMAC` | HMAC signature generation | algorithm, encoding, secret_key, data_format |
| `SIGNATURE` | Custom signature generation | algorithm, secret_key, fields |
| `BASIC_AUTH` | HTTP Basic Authentication | username_key, password_key |
| `API_KEY` | API key in header | key_name, key_value |
| `OAUTH` | OAuth token management | token_url, client_id, client_secret |

#### 4.6 Add Response Mappings

```go
// Map order_id field
err = admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             uuid.New().String(),
    ProviderID:     providerID,
    Action:         "create_order",
    TargetField:    "order_id",
    SourceJSONPath: "$.data.order.id",
    IsRequired:     true,
    Priority:       1,
})

// Map checkout URL
err = admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             uuid.New().String(),
    ProviderID:     providerID,
    Action:         "create_order",
    TargetField:    "checkout_url",
    SourceJSONPath: "$.data.order.checkout_url",
    Transform:      "",
    IsRequired:     true,
    Priority:       2,
})

// Map status with default value
err = admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
    ID:             uuid.New().String(),
    ProviderID:     providerID,
    Action:         "create_order",
    TargetField:    "status",
    SourceJSONPath: "$.data.order.status",
    DefaultValue:   "pending",
    IsRequired:     false,
    Priority:       3,
})
```

### Step 5: Execute API Calls

#### 5.1 Build cURL (Debug Mode)

```go
// Get cURL command without executing
curlCmd, err := orch.BuildCurl("banxa", "create_order", map[string]interface{}{
    "fiat_amount": 100.00,
    "fiat_code":   "USD",
    "coin_code":   "BTC",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println("cURL Command:")
fmt.Println(curlCmd)
// Output: curl -X POST 'https://api.banxa.com/api/orders' -H 'Content-Type: application/json' ...
```

#### 5.2 Build cURL with Details

```go
// Get detailed request information
details, err := orch.BuildCurlWithDetails("banxa", "create_order", map[string]interface{}{
    "fiat_amount": 100.00,
    "fiat_code":   "USD",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Provider: %s\n", details.Provider)
fmt.Printf("Action: %s\n", details.Action)
fmt.Printf("Method: %s\n", details.Method)
fmt.Printf("URL: %s\n", details.URL)
fmt.Printf("Headers:\n")
for key, value := range details.Headers {
    fmt.Printf("  %s: %s\n", key, value)
}
fmt.Printf("Body: %s\n", details.Body)
fmt.Printf("\ncURL Command:\n%s\n", details.CurlCommand)
```

#### 5.3 Execute API Call

```go
// Execute the actual API call
response, err := orch.Call("banxa", "create_order", map[string]interface{}{
    "fiat_amount":    100.00,
    "fiat_code":      "USD",
    "coin_code":      "BTC",
    "wallet_address": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
})

if err != nil {
    log.Fatal(err)
}

// Check response
if response.Success {
    fmt.Printf("Order ID: %v\n", response.Data["order_id"])
    fmt.Printf("Checkout URL: %v\n", response.Data["checkout_url"])
    fmt.Printf("Status: %v\n", response.Data["status"])
} else {
    fmt.Printf("Error: %s\n", response.Error)
}

// Access raw response
fmt.Printf("Raw Response: %+v\n", response.RawResponse)
```

**Response Structure:**

```go
type Response struct {
    Success     bool                   `json:"success"`      // True if API call succeeded
    StatusCode  int                    `json:"status_code"`  // HTTP status code
    Provider    string                 `json:"provider"`     // Provider name
    Action      string                 `json:"action"`       // Action/endpoint name
    RequestID   string                 `json:"request_id"`   // Unique request ID
    Timestamp   string                 `json:"timestamp"`    // ISO 8601 timestamp
    Data        map[string]interface{} `json:"data"`         // Mapped response data
    Message     string                 `json:"message"`      // Success/error message
    Error       string                 `json:"error"`        // Error details (if any)
    RawResponse map[string]interface{} `json:"raw_response"` // Original provider response
}
```

---

## Configuration

### Environment Variables

```bash
# Required - Only encryption key is needed
ORCHESTRATOR_ENCRYPTION_KEY=your-32-character-secret-key!

# Optional - Table prefix to avoid conflicts
ORCHESTRATOR_TABLE_PREFIX=orch_

# Database (if not already set in your app)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=your_existing_database

# Optional: Application-specific settings
LOG_LEVEL=info
PORT=8080

# Note: Provider credentials (API keys, secrets) are stored in the database
# via CreateCredential API, NOT as environment variables
```

### Configuration Files

Create `.env` file in your project root:

```env
ORCHESTRATOR_ENCRYPTION_KEY=12345678901234567890123456789012
ORCHESTRATOR_TABLE_PREFIX=orch_
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=my_app_database
LOG_LEVEL=info
```

**Note:** Provider API keys and secrets are stored in the database (encrypted), not in environment variables.

**Table Prefix Usage:**

Set `ORCHESTRATOR_TABLE_PREFIX` if your application already has tables named `providers`, `credentials`, `endpoints`, etc. This creates prefixed tables like `orch_providers`, `orch_credentials`, etc., allowing both to coexist.

```go
// Load table prefix from environment
tablePrefix := os.Getenv("ORCHESTRATOR_TABLE_PREFIX")

// Use in migrations
orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
    TablePrefix: tablePrefix,
})

// Use in orchestrator
orchestrator.New(orchestrator.Config{
    DB:          db,
    TablePrefix: tablePrefix,
})
```

Load with godotenv:

```go
import "github.com/joho/godotenv"

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}
```

---

## API Reference

### Initialization Functions

```go
// Set encryption key from string (must be 32 characters)
func SetEncryptionKeyFromString(key string) error

// Set encryption key from byte array (must be 32 bytes)
func SetEncryptionKey(key []byte) error

// Check if encryption key is set
func IsEncryptionKeySet() bool

// Clear encryption key (for testing)
func ClearEncryptionKey()

// Create orchestrator instance
func New(cfg Config) (*Orchestrator, error)

// Run database migrations
func Migrate(db *gorm.DB) error
```

### AdminAPI Methods

Get admin API:
```go
admin := orch.Admin()
```

**Provider Management:**

```go
func (a *AdminAPI) CreateProvider(cfg ProviderConfig) error
func (a *AdminAPI) GetProvider(id string) (*ProviderConfig, error)
func (a *AdminAPI) GetProviderByName(name string) (*ProviderConfig, error)
func (a *AdminAPI) ListProviders() ([]ProviderConfig, error)
func (a *AdminAPI) ListActiveProviders() ([]ProviderConfig, error)
func (a *AdminAPI) UpdateProvider(cfg ProviderConfig) error
func (a *AdminAPI) DeleteProvider(id string) error
```

**Credential Management:**

```go
func (a *AdminAPI) CreateCredential(cfg CredentialConfig) error
func (a *AdminAPI) GetCredentialsByProvider(providerID string) (map[string]string, error)
func (a *AdminAPI) UpdateCredential(id string, value string) error
func (a *AdminAPI) DeleteCredential(id string) error
```

**Endpoint Management:**

```go
func (a *AdminAPI) CreateEndpoint(cfg EndpointConfig) error
func (a *AdminAPI) GetEndpoint(id string) (*EndpointConfig, error)
func (a *AdminAPI) GetEndpointByName(providerID, name string) (*EndpointConfig, error)
func (a *AdminAPI) ListEndpointsByProvider(providerID string) ([]EndpointConfig, error)
func (a *AdminAPI) UpdateEndpoint(cfg EndpointConfig) error
func (a *AdminAPI) DeleteEndpoint(id string) error
```

**Header Rule Management:**

```go
func (a *AdminAPI) CreateHeaderRule(cfg HeaderRuleConfig) error
func (a *AdminAPI) GetHeaderRulesByProvider(providerID string) ([]HeaderRuleConfig, error)
func (a *AdminAPI) UpdateHeaderRule(cfg HeaderRuleConfig) error
func (a *AdminAPI) DeleteHeaderRule(id string) error
```

**Strategy Management:**

```go
func (a *AdminAPI) CreateStrategy(cfg StrategyConfig) error
func (a *AdminAPI) GetStrategy(name string) (*StrategyConfig, error)
func (a *AdminAPI) GetStrategyByName(name string) (*StrategyConfig, error)
func (a *AdminAPI) ListStrategies() ([]StrategyConfig, error)
func (a *AdminAPI) ListStrategiesByType(strategyType string) ([]StrategyConfig, error)
func (a *AdminAPI) UpdateStrategy(cfg StrategyConfig) error
func (a *AdminAPI) DeleteStrategy(id string) error
```

**Response Mapping Management:**

```go
func (a *AdminAPI) CreateResponseMapping(cfg ResponseMappingConfig) error
func (a *AdminAPI) GetResponseMappings(providerID, action string) ([]ResponseMappingConfig, error)
func (a *AdminAPI) UpdateResponseMapping(cfg ResponseMappingConfig) error
func (a *AdminAPI) DeleteResponseMapping(id string) error
```

### Orchestrator Methods

```go
// Build cURL command without executing
func (o *Orchestrator) BuildCurl(provider, action string, params map[string]interface{}) (string, error)

// Build cURL with detailed request information
func (o *Orchestrator) BuildCurlWithDetails(provider, action string, params map[string]interface{}) (*CurlDetails, error)

// Execute API call and return unified response
func (o *Orchestrator) Call(provider, action string, params map[string]interface{}) (*Response, error)

// Get AdminAPI for configuration
func (o *Orchestrator) Admin() *AdminAPI
```

---

## Usage Examples

### Example 1: Mobile Backend Integration

```go
package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "github.com/gin-gonic/gin"
)

type PaymentHandler struct {
    orch *orchestrator.Orchestrator
}

func NewPaymentHandler(orch *orchestrator.Orchestrator) *PaymentHandler {
    return &PaymentHandler{orch: orch}
}

// POST /api/v1/payment/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
    var req struct {
        Provider      string  `json:"provider" binding:"required"`
        Amount        float64 `json:"amount" binding:"required"`
        Currency      string  `json:"currency" binding:"required"`
        CryptoCoin    string  `json:"crypto_coin" binding:"required"`
        WalletAddress string  `json:"wallet_address" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Execute API call through orchestrator
    response, err := h.orch.Call(req.Provider, "create_order", map[string]interface{}{
        "fiat_amount":    req.Amount,
        "fiat_code":      req.Currency,
        "coin_code":      req.CryptoCoin,
        "wallet_address": req.WalletAddress,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// GET /api/v1/payment/orders/:provider/:order_id
func (h *PaymentHandler) GetOrderStatus(c *gin.Context) {
    provider := c.Param("provider")
    orderID := c.Param("order_id")
    
    response, err := h.orch.Call(provider, "get_order_status", map[string]interface{}{
        "order_id": orderID,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, response)
}

func main() {
    // Initialize orchestrator
    orchestrator.SetEncryptionKeyFromString(os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY"))
    db := initDatabase()
    orchConfig.AutoMigrate(db)
    
    orch, _ := orchestrator.New(orchestrator.Config{DB: db})
    
    // Setup routes
    r := gin.Default()
    handler := NewPaymentHandler(orch)
    
    v1 := r.Group("/api/v1")
    {
        v1.POST("/payment/orders", handler.CreateOrder)
        v1.GET("/payment/orders/:provider/:order_id", handler.GetOrderStatus)
    }
    
    r.Run(":8080")
}
```

### Example 2: Multi-Provider Failover

```go
func CreateOrderWithFailover(orch *orchestrator.Orchestrator, amount float64, currency string) (*orchestrator.Response, error) {
    providers := []string{"banxa", "transak", "moonpay"}
    
    var lastErr error
    for _, provider := range providers {
        response, err := orch.Call(provider, "create_order", map[string]interface{}{
            "fiat_amount": amount,
            "fiat_code":   currency,
        })
        
        if err == nil && response.Success {
            return response, nil
        }
        
        lastErr = err
        log.Printf("Provider %s failed: %v", provider, err)
    }
    
    return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}
```

### Example 3: Debug Mode

```go
func DebugAPICall(orch *orchestrator.Orchestrator, provider, action string, params map[string]interface{}) {
    // Build cURL without executing
    curlCmd, err := orch.BuildCurl(provider, action, params)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("=== cURL Command ===")
    fmt.Println(curlCmd)
    fmt.Println()
    
    // Get detailed information
    details, err := orch.BuildCurlWithDetails(provider, action, params)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("=== Request Details ===")
    fmt.Printf("Method: %s\n", details.Method)
    fmt.Printf("URL: %s\n", details.URL)
    fmt.Println("Headers:")
    for key, value := range details.Headers {
        fmt.Printf("  %s: %s\n", key, value)
    }
    fmt.Printf("Body: %s\n", details.Body)
    
    // Now execute
    fmt.Println("\n=== Executing API Call ===")
    response, err := orch.Call(provider, action, params)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Success: %v\n", response.Success)
    fmt.Printf("Status Code: %d\n", response.StatusCode)
    fmt.Printf("Data: %+v\n", response.Data)
}
```

---

## Best Practices

### Security

1. **Never commit encryption keys** to version control
2. **Use environment variables** for all sensitive credentials
3. **Rotate encryption keys** periodically (requires re-encrypting credentials)
4. **Use HTTPS** for all provider API calls
5. **Validate input parameters** before passing to orchestrator
6. **Log only non-sensitive data** (orchestrator logs don't include credentials)

### Performance

1. **Reuse orchestrator instance** - don't create new instances per request
2. **Cache provider configurations** in memory when possible
3. **Use connection pooling** for database
4. **Monitor API latency** and set appropriate timeouts
5. **Implement rate limiting** per provider

### Error Handling

1. **Check errors** from all orchestrator methods
2. **Implement retry logic** for transient failures
3. **Log errors** with context (provider, action, request ID)
4. **Return user-friendly errors** to clients
5. **Monitor error rates** per provider

### Configuration Management

1. **Use database for dynamic configuration** (providers, endpoints)
2. **Use code/config files** for static setup (initial provider setup)
3. **Version control** provider configurations
4. **Test configurations** in staging before production
5. **Document provider-specific requirements** (rate limits, auth methods)

---

## Troubleshooting

### Common Issues

#### Issue 1: "encryption key not set"

**Error:**
```
panic: encryption key not set
```

**Solution:**
```go
// Call SetEncryptionKeyFromString BEFORE creating orchestrator
orchestrator.SetEncryptionKeyFromString("your-32-character-secret-key!")
orch, err := orchestrator.New(orchestrator.Config{DB: db})
```

#### Issue 2: "encryption key must be exactly 32 bytes"

**Error:**
```
encryption key must be exactly 32 bytes, got 16
```

**Solution:**
```bash
# Generate 32-character key
openssl rand -base64 24 | head -c 32
```

#### Issue 3: "provider not found"

**Error:**
```
provider not found: banxa
```

**Solution:**
```go
// Create provider first
admin := orch.Admin()
admin.CreateProvider(orchestrator.ProviderConfig{
    ID:   uuid.New().String(),
    Name: "banxa",
    // ... other fields
})
```

#### Issue 4: "endpoint not found"

**Error:**
```
endpoint not found: create_order
```

**Solution:**
```go
// Create endpoint for the provider
admin.CreateEndpoint(orchestrator.EndpointConfig{
    ProviderID: providerID,
    Name:       "create_order",
    Method:     "POST",
    Path:       "/api/orders",
})
```

#### Issue 5: Tables not created

**Error:**
```
table "providers" does not exist
```

**Solution:**
```go
// Run migrations
import orchConfig "github.com/PayRam/api-orchestrator-go/config"
err := orchConfig.AutoMigrate(db)
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

#### Issue 6: Table name conflicts

**Error:**
```
ERROR: relation "providers" already exists
```
or
```
ERROR: duplicate key value violates unique constraint
```

**Cause:** Your application already has tables with the same names.

**Solution:** Use table prefix to avoid conflicts
```go
// Set environment variable
export ORCHESTRATOR_TABLE_PREFIX=orch_

// Or use in code
err := orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
    TablePrefix: "orch_",
})

// Create orchestrator with same prefix
orch, err := orchestrator.New(orchestrator.Config{
    DB:          db,
    TablePrefix: "orch_",
})
```

This creates: `orch_providers`, `orch_credentials`, etc.

See [Table Prefix Documentation](docs/TABLE_PREFIX.md) for complete guide.

#### Issue 7: Credentials not working

**Issue:** API calls fail with authentication errors

**Solution:**
1. Verify encryption key is the same between application restarts
2. Check credential values in database (should be encrypted)
3. Verify credential keys match header rules (e.g., "API_KEY")
4. Test with cURL command first: `orch.BuildCurl()`

---

## FAQ

### Q: Do I need to create a new database?

**A:** No, the library uses your existing database and creates its own tables alongside yours.

### Q: Can I use this with MongoDB/MySQL/SQLite?

**A:** The library requires PostgreSQL with GORM. MySQL support may be added in future versions.

### Q: How are credentials stored?

**A:** Credentials are encrypted using AES-256-GCM before being stored in the database. Only encrypted values are persisted.

### Q: Can I add providers dynamically without restarting?

**A:** Yes! Use the AdminAPI to add/modify providers, credentials, and endpoints at runtime.

### Q: How do I handle provider-specific authentication?

**A:** Use authentication strategies. Create a strategy (HMAC, signature, OAuth) and reference it in header rules.

### Q: Can I use this in a microservices architecture?

**A:** Yes! Each microservice can have its own orchestrator instance pointing to the same database, or use a shared orchestrator service.

### Q: What happens if a provider API changes?

**A:** Update the endpoint configuration, header rules, or response mappings via AdminAPI without code changes.

### Q: How do I test without calling real APIs?

**A:** Use `BuildCurl()` to generate cURL commands for testing, or mock the HTTP executor in tests.

### Q: Is this production-ready?

**A:** Yes! The library includes comprehensive logging, error handling, encryption, and has been tested in production environments.

### Q: Can I contribute?

**A:** Yes! Contributions are welcome. Please see CONTRIBUTING.md for guidelines.

---

## Support

- **Documentation**: [GitHub Wiki](https://github.com/PayRam/api-orchestrator-go/wiki)
- **Issues**: [GitHub Issues](https://github.com/PayRam/api-orchestrator-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/PayRam/api-orchestrator-go/discussions)

---

## License

MIT License - see LICENSE file for details

---

## Credits

Developed by PayRam Team

**Contributors:**
- Core Team
- Community Contributors

**Special Thanks:**
- GORM team for excellent ORM
- Zap team for structured logging
- Go community for feedback and contributions
