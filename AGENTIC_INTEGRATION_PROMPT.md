# AI Agent Integration Prompt - API Orchestrator Go

This document provides comprehensive instructions for AI agents to integrate the API Orchestrator Go library into existing Go applications.

---

## Mission Statement

You are tasked with integrating the **API Orchestrator Go** library into an existing Go application. This library enables dynamic integration with multiple payment provider APIs (Banxa, Transak, MoonPay, etc.) through a unified interface, eliminating the need for provider-specific code.

---

## Prerequisites Check

Before starting, verify:
- ✅ Application is written in Go 1.21+
- ✅ Application uses PostgreSQL with GORM
- ✅ Application has an existing database connection
- ✅ You have access to provider API credentials (Banxa, Transak, etc.)

---

## Integration Objectives

1. Add the library to the existing Go project
2. Initialize the orchestrator using the existing database
3. Create HTTP endpoints that use the orchestrator
4. Test the integration with cURL commands

**Note:** Provider configuration (Banxa, Transak, etc.) is done separately via Admin API, not during library integration.

---

## Step-by-Step Integration Instructions

### Step 1: Add Library Dependency

**Action:** Add the library to go.mod and download dependencies.

**Commands:**
```bash
cd /path/to/project
go get github.com/PayRam/api-orchestrator-go
go mod tidy
```

**Verification:**
- Check that `go.mod` contains `github.com/PayRam/api-orchestrator-go`
- Run `go mod verify` to ensure all dependencies are downloaded

---

### Step 2: Locate Database Initialization

**Action:** Find where the application initializes the GORM database connection.

**Search Strategy:**
1. Look for files containing:
   - `gorm.Open(`
   - `*gorm.DB`
   - Database initialization patterns
   
2. Common locations:
   - `main.go`
   - `internal/database/database.go`
   - `pkg/database/db.go`
   - `config/database.go`
   - `db/connection.go`

**Code Pattern to Find:**
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

**What to Note:**
- Variable name for database instance (usually `db`)
- Where it's stored (global, struct field, context)
- Whether it's already being passed to other components

---

### Step 3: Set Environment Variables

**Action:** Add the required encryption key environment variable.

**Required Variable:**
```bash
# ONLY THIS IS REQUIRED - Provider credentials are stored in database
ORCHESTRATOR_ENCRYPTION_KEY=your-32-character-secret-key!

# OPTIONAL - Table prefix to avoid conflicts with existing tables
# If your app already has "providers", "credentials" tables, set a prefix
ORCHESTRATOR_TABLE_PREFIX=orch_
```

**Implementation:**

**Option A: Add to existing .env file**
```bash
# Find .env file
find . -name ".env" -o -name ".env.example"

# Add only the encryption key (required)
echo "ORCHESTRATOR_ENCRYPTION_KEY=12345678901234567890123456789012" >> .env

# Add table prefix (optional, only if you have table name conflicts)
echo "ORCHESTRATOR_TABLE_PREFIX=orch_" >> .env
```

**Option B: Create new .env file**
```bash
cat > .env << EOF
ORCHESTRATOR_ENCRYPTION_KEY=12345678901234567890123456789012
ORCHESTRATOR_TABLE_PREFIX=orch_
EOF
```

**Generate Secure Encryption Key:**
```bash
# Linux/Mac
openssl rand -base64 24 | head -c 32

# Or use a strong password (exactly 32 characters)
echo "MySecureKey12345678901234567890" | wc -c  # Should output 33 (32 + newline)
```

**Important Notes:**
- ⚠️ **Provider credentials (API keys, secrets) are NOT stored as environment variables**
- ✅ **Provider credentials are stored in the database via CreateCredential API**
- ✅ **Credentials are automatically encrypted using the encryption key**
- ✅ **Credentials are retrieved at runtime when making API calls**

**Verification:**
- Encryption key is exactly 32 characters
- Application can load environment variables (check for godotenv or similar)

---

### Step 4: Add Orchestrator Initialization

**Action:** Create a function to initialize the orchestrator using the existing database.

**Location:** Add to the file where database is initialized, or create new file `internal/orchestrator/setup.go`

**Code to Add:**

```go
package main // or appropriate package

import (
    "fmt"
    "log"
    "os"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    orchConfig "github.com/PayRam/api-orchestrator-go/config"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// Global orchestrator instance (or add to your app context)
var globalOrchestrator *orchestrator.Orchestrator

// InitializeOrchestrator sets up the API orchestrator with the existing database
func InitializeOrchestrator(db *gorm.DB, logger *zap.Logger) error {
    // 1. Set encryption key
    encryptionKey := os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY")
    if len(encryptionKey) != 32 {
        return fmt.Errorf("ORCHESTRATOR_ENCRYPTION_KEY must be exactly 32 characters, got %d", len(encryptionKey))
    }
    
    err := orchestrator.SetEncryptionKeyFromString(encryptionKey)
    if err != nil {
        return fmt.Errorf("failed to set encryption key: %w", err)
    }
    
    log.Println("✅ Encryption key set successfully")
    
    // 2. (Optional) Set table prefix to avoid conflicts with existing tables
    // If your app already has tables named "providers", "credentials", etc.,
    // set a prefix to create "orch_providers", "orch_credentials", etc.
    tablePrefix := os.Getenv("ORCHESTRATOR_TABLE_PREFIX") // e.g., "orch_"
    if tablePrefix != "" {
        // Option A: Using AutoMigrateWithOptions (recommended)
        err = orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
            TablePrefix: tablePrefix,
        })
        if err != nil {
            return fmt.Errorf("failed to run orchestrator migrations: %w", err)
        }
        log.Printf("✅ Orchestrator tables created with prefix: %s", tablePrefix)
    } else {
        // Option B: Standard migration without prefix
        err = orchConfig.AutoMigrate(db)
        if err != nil {
            return fmt.Errorf("failed to run orchestrator migrations: %w", err)
        }
        log.Println("✅ Orchestrator tables created in existing database")
    }
    
    // 3. Create orchestrator instance (use same prefix if set)
    globalOrchestrator, err = orchestrator.New(orchestrator.Config{
        DB:          db,          // Use existing database connection
        Logger:      logger,      // Use existing logger (optional)
        TablePrefix: tablePrefix, // Use same prefix as migrations
    })
    if err != nil {
        return fmt.Errorf("failed to create orchestrator: %w", err)
    }
    
    log.Println("✅ Orchestrator initialized successfully")
    
    return nil
}

// GetOrchestrator returns the global orchestrator instance
func GetOrchestrator() *orchestrator.Orchestrator {
    return globalOrchestrator
}
```

**Integration Point:**

Find the application's initialization sequence (usually in `main()` or `Run()` function):

```go
func main() {
    // Existing code...
    db := initializeDatabase() // Existing database initialization
    logger := initializeLogger() // Existing logger initialization
    
    // ADD THIS: Initialize orchestrator
    err := InitializeOrchestrator(db, logger)
    if err != nil {
        log.Fatal("Failed to initialize orchestrator:", err)
    }
    
    // Rest of initialization...
    startServer()
}
```

**Verification:**
- Application compiles without errors
- On startup, you see logs: "Encryption key set successfully", "Orchestrator tables created"
- Database contains 8 new tables: providers, credentials, endpoints, header_rules, strategies, request_schemas, request_values, response_mappings
  - OR with prefix: orch_providers, orch_credentials, orch_endpoints, orch_header_rules, orch_strategies, orch_request_schemas, orch_request_values, orch_response_mappings

**Table Name Conflicts:**

If your application already has tables named `providers`, `credentials`, `endpoints`, etc., you'll get migration errors. Use the table prefix feature:

```bash
# Set environment variable
export ORCHESTRATOR_TABLE_PREFIX=orch_

# Or in .env file
ORCHESTRATOR_TABLE_PREFIX=orch_
```

This creates tables with prefix:
- `orch_providers` instead of `providers`
- `orch_credentials` instead of `credentials`
- etc.

**Example with Table Prefix:**

```go
func InitializeOrchestrator(db *gorm.DB, logger *zap.Logger) error {
    // Set encryption key
    orchestrator.SetEncryptionKeyFromString(os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY"))
    
    // Get table prefix (optional)
    tablePrefix := os.Getenv("ORCHESTRATOR_TABLE_PREFIX")
    
    // Run migrations with prefix
    if tablePrefix != "" {
        err := orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
            TablePrefix: tablePrefix,
        })
        if err != nil {
            return err
        }
    } else {
        err := orchConfig.AutoMigrate(db)
        if err != nil {
            return err
        }
    }
    
    // Create orchestrator with same prefix
    globalOrchestrator, err = orchestrator.New(orchestrator.Config{
        DB:          db,
        Logger:      logger,
        TablePrefix: tablePrefix, // Important: use same prefix
    })
    
    return err
}
```

**Recommended Prefixes:**
- `orch_` - Short and clear
- `orchestrator_` - More explicit
- `api_orch_` - If you have multiple orchestrators
- `payram_` - Your company/project name

See [Table Prefix Documentation](docs/TABLE_PREFIX.md) for detailed usage.

---

### Step 5: Create HTTP Handlers

**Action:** Create handlers that use the orchestrator to call provider APIs.

**Location:** Create new file `internal/handlers/orchestrator_handler.go` (or add to existing handlers)

**Important:** This is a **generic handler pattern**. Adapt the request/response structure to your application's domain (payments, shipping, SMS, etc.).

**Code to Add:**

```go
package handlers

import (
    "net/http"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "github.com/gin-gonic/gin" // or your router framework
    "go.uber.org/zap"
)

// OrchestratorHandler handles generic provider API calls
type OrchestratorHandler struct {
    orch   *orchestrator.Orchestrator
    logger *zap.Logger
}

// NewOrchestratorHandler creates a new orchestrator handler
func NewOrchestratorHandler(orch *orchestrator.Orchestrator, logger *zap.Logger) *OrchestratorHandler {
    return &OrchestratorHandler{
        orch:   orch,
        logger: logger,
    }
}

// CallProvider executes an API call through the orchestrator
// POST /api/v1/orchestrator/call
func (h *OrchestratorHandler) CallProvider(c *gin.Context) {
    var req struct {
        Provider string                 `json:"provider" binding:"required"`
        Action   string                 `json:"action" binding:"required"`
        Params   map[string]interface{} `json:"params"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid request", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if req.Params == nil {
        req.Params = make(map[string]interface{})
    }
    
    h.logger.Info("Calling provider",
        zap.String("provider", req.Provider),
        zap.String("action", req.Action))
    
    // Execute API call through orchestrator
    response, err := h.orch.Call(req.Provider, req.Action, req.Params)
    if err != nil {
        h.logger.Error("Failed to call provider", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // Return unified response
    c.JSON(response.StatusCode, response)
}

// BuildCurl generates a cURL command without executing (for debugging)
// POST /api/v1/orchestrator/build-curl
func (h *OrchestratorHandler) BuildCurl(c *gin.Context) {
    var req struct {
        Provider string                 `json:"provider" binding:"required"`
        Action   string                 `json:"action" binding:"required"`
        Params   map[string]interface{} `json:"params"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if req.Params == nil {
        req.Params = make(map[string]interface{})
    }
    
    // Build cURL without executing
    curlCmd, err := h.orch.BuildCurl(req.Provider, req.Action, req.Params)
    if err != nil {
        h.logger.Error("Failed to build cURL", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "curl_command": curlCmd,
        "provider":     req.Provider,
        "action":       req.Action,
    })
}

// BuildCurlDetailed generates detailed cURL information
// POST /api/v1/orchestrator/build-curl-detailed
func (h *OrchestratorHandler) BuildCurlDetailed(c *gin.Context) {
    var req struct {
        Provider string                 `json:"provider" binding:"required"`
        Action   string                 `json:"action" binding:"required"`
        Params   map[string]interface{} `json:"params"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if req.Params == nil {
        req.Params = make(map[string]interface{})
    }
    
    // Build detailed cURL
    details, err := h.orch.BuildCurlWithDetails(req.Provider, req.Action, req.Params)
    if err != nil {
        h.logger.Error("Failed to build detailed cURL", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    details,
    })
}
```

**Domain-Specific Wrapper Example (Optional):**

If your application needs domain-specific validation or business logic, create a wrapper:

```go
// Example: Payment-specific handler (wraps orchestrator)
type PaymentHandler struct {
    orchestratorHandler *OrchestratorHandler
}

func (h *PaymentHandler) CreateOrder(c *gin.Context) {
    var req struct {
        Provider   string  `json:"provider" binding:"required"`
        Amount     float64 `json:"amount" binding:"required,gt=0"`
        Currency   string  `json:"currency" binding:"required,len=3"`
        // ... other payment-specific fields
    }
    
    // Validate business rules
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Transform to generic params
    params := map[string]interface{}{
        "amount":   req.Amount,
        "currency": req.Currency,
    }
    
    // Delegate to orchestrator
    response, _ := h.orchestratorHandler.orch.Call(req.Provider, "create_order", params)
    c.JSON(response.StatusCode, response)
}
```

**Verification:**
- Handler file compiles without errors
- All required imports are present
- Handler methods match your routing framework (Gin/Echo/Gorilla)
- Generic pattern can be adapted to any domain

---

### Step 6: Register Routes

**Action:** Add routes for the orchestrator handler.

**Location:** Find where routes are registered (usually `routes.go`, `router.go`, or `main.go`)

**Code to Add:**

```go
func SetupRoutes(r *gin.Engine, orch *orchestrator.Orchestrator, logger *zap.Logger) {
    // Existing routes...
    
    // ADD THIS: Orchestrator routes
    orchHandler := handlers.NewOrchestratorHandler(orch, logger)
    
    v1 := r.Group("/api/v1")
    {
        // Generic orchestrator endpoints
        orchestratorGroup := v1.Group("/orchestrator")
        {
            // Execute API call
            orchestratorGroup.POST("/call", orchHandler.CallProvider)
            
            // Debug/testing endpoints
            orchestratorGroup.POST("/build-curl", orchHandler.BuildCurl)
            orchestratorGroup.POST("/build-curl-detailed", orchHandler.BuildCurlDetailed)
        }
        
        // Optional: Add domain-specific routes here
        // payment := v1.Group("/payment")
        // shipping := v1.Group("/shipping")
        // etc.
    }
}
```

**Call Route Setup:**

```go
func main() {
    // ... initialization ...
    
    router := gin.Default()
    
    // ADD THIS: Setup routes with orchestrator
    SetupRoutes(router, GetOrchestrator(), logger)
    
    router.Run(":8080")
}
```

**Verification:**
- Application compiles
- Routes are registered (check startup logs)
- Can access generic orchestrator endpoints

---

### Step 7: Test Integration

**Important Note:** Before testing, you need to configure at least one provider using the Admin API or REST endpoints. See the "Provider Configuration" section below for details.

**Action:** Test the integration with cURL commands.

**Test 1: Health Check (if exists)**
```bash
curl http://localhost:8080/health
```

**Test 2: Preview cURL without executing**
```bash
curl -X POST http://localhost:8080/api/v1/orchestrator/build-curl \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "your-provider-name",
    "action": "your-action-name",
    "params": {
      "param1": "value1",
      "param2": "value2"
    }
  }'
```

**Expected Response:**
```json
{
  "curl_command": "curl -X POST 'https://api.provider.com/endpoint' -H 'Content-Type: application/json' -H 'Authorization: Bearer ...' -d '{...}'",
  "provider": "your-provider-name",
  "action": "your-action-name"
}
```

**Test 3: Get detailed cURL information**
```bash
curl -X POST http://localhost:8080/api/v1/orchestrator/build-curl-detailed \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "your-provider-name",
    "action": "your-action-name",
    "params": {
      "param1": "value1"
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "curl_command": "curl -X POST...",
    "method": "POST",
    "url": "https://api.provider.com/endpoint",
    "headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer ..."
    },
    "body": "{\"param1\":\"value1\"}",
    "provider": "your-provider-name",
    "action": "your-action-name"
  }
}
```

**Test 4: Execute actual API call**
```bash
curl -X POST http://localhost:8080/api/v1/orchestrator/call \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "your-provider-name",
    "action": "your-action-name",
    "params": {
      "param1": "value1",
      "param2": "value2"
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "status_code": 200,
  "provider": "your-provider-name",
  "action": "your-action-name",
  "request_id": "uuid-here",
  "timestamp": "2025-12-30T10:30:00Z",
  "data": {
    "field1": "value1",
    "field2": "value2"
  },
  "message": "Request successful",
  "raw_response": {...}
}
```

**Verification Checklist:**
- ✅ Preview cURL returns valid cURL command
- ✅ Detailed preview shows method, URL, headers, body
- ✅ Actual API call succeeds (or returns expected provider error)
- ✅ Response has consistent format (success, status_code, data, etc.)
- ✅ Logs show request flow through orchestrator

---

## Provider Configuration (Separate from Integration)

**Important:** Provider configuration is **NOT part of library integration**. It's done separately after integration is complete.

### Option 1: Using AdminAPI Directly (Recommended)

Create a separate admin script or one-time setup function:

```go
// cmd/setup-providers/main.go
package main

import (
    "log"
    "os"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    orchConfig "github.com/PayRam/api-orchestrator-go/config"
    "github.com/google/uuid"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // 1. Connect to database
    dsn := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // 2. Set encryption key
    err = orchestrator.SetEncryptionKeyFromString(os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY"))
    if err != nil {
        log.Fatal("Failed to set encryption key:", err)
    }
    
    // 3. Run migrations
    err = orchConfig.AutoMigrate(db)
    if err != nil {
        log.Fatal("Failed to run migrations:", err)
    }
    
    // 4. Initialize orchestrator
    orch, err := orchestrator.New(orchestrator.Config{DB: db})
    if err != nil {
        log.Fatal("Failed to create orchestrator:", err)
    }
    
    admin := orch.Admin()
    
    // 5. Configure provider
    providerID := uuid.New().String()
    err = admin.CreateProvider(orchestrator.ProviderConfig{
        ID:          providerID,
        Name:        "your-provider-name",
        DisplayName: "Your Provider",
        BaseURL:     "https://api.provider.com",
        IsActive:    true,
        Description: "Provider description",
    })
    if err != nil {
        log.Fatal("Failed to create provider:", err)
    }
    
    // 6. Add credentials (will be encrypted)
    err = admin.CreateCredential(orchestrator.CredentialConfig{
        ProviderID:  providerID,
        Key:         "API_KEY",
        Value:       "your-actual-api-key",
        Required:    true,
        Description: "API Key",
    })
    if err != nil {
        log.Fatal("Failed to create credential:", err)
    }
    
    // 7. Create endpoint
    err = admin.CreateEndpoint(orchestrator.EndpointConfig{
        ID:          uuid.New().String(),
        ProviderID:  providerID,
        Name:        "your-action-name",
        Method:      "POST",
        Path:        "/api/endpoint",
        Description: "Action description",
    })
    if err != nil {
        log.Fatal("Failed to create endpoint:", err)
    }
    
    // 8. Add header rules
    err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
        ID:              uuid.New().String(),
        ProviderID:      providerID,
        HeaderName:      "Authorization",
        ValueExpression: "credential:API_KEY",
        Priority:        1,
        Description:     "API authentication",
    })
    if err != nil {
        log.Fatal("Failed to create header rule:", err)
    }
    
    // 9. Add response mappings (optional)
    err = admin.CreateResponseMapping(orchestrator.ResponseMappingConfig{
        ID:             uuid.New().String(),
        ProviderID:     providerID,
        Action:         "your-action-name",
        TargetField:    "result_id",
        SourceJSONPath: "$.data.id",
        IsRequired:     true,
        Priority:       1,
    })
    if err != nil {
        log.Fatal("Failed to create response mapping:", err)
    }
    
    log.Println("✅ Provider configured successfully")
}
```

Run once:
```bash
go run cmd/setup-providers/main.go
```

### Option 2: Configuration on First Run (Check if Provider Exists)

Add provider setup in your application startup that only runs if provider doesn't exist:

```go
func ensureProvidersConfigured(orch *orchestrator.Orchestrator) error {
    admin := orch.Admin()
    
    // Check if provider already exists
    existingProvider, _ := admin.GetProvider("your-provider-name")
    if existingProvider != nil {
        log.Println("Provider already configured, skipping setup")
        return nil
    }
    
    // Configure provider (same as Option 1)
    providerID := uuid.New().String()
    err := admin.CreateProvider(orchestrator.ProviderConfig{
        ID:          providerID,
        Name:        "your-provider-name",
        DisplayName: "Your Provider",
        BaseURL:     "https://api.provider.com",
        IsActive:    true,
    })
    // ... rest of configuration
    
    return err
}

// In main.go
func main() {
    // ... after InitializeOrchestrator ...
    err := ensureProvidersConfigured(GetOrchestrator())
    if err != nil {
        log.Printf("Warning: Provider setup failed: %v", err)
    }
    // ... continue with server startup
}
```

### Option 3: Load from Configuration File

Create a JSON/YAML configuration file and load it:

```json
// config/providers.json
{
  "providers": [
    {
      "name": "banxa",
      "display_name": "Banxa",
      "base_url": "https://api.banxa.com",
      "credentials": [
        {"key": "API_KEY", "value": "pk_live_..."},
        {"key": "API_SECRET", "value": "sk_live_..."}
      ],
      "endpoints": [
        {"name": "create_order", "method": "POST", "path": "/api/orders"}
      ]
    }
  ]
}
```

### Option 3: Load from Configuration File

Create a JSON/YAML configuration file and load it:

```json
// config/providers.json
{
  "providers": [
    {
      "name": "your-provider-name",
      "display_name": "Your Provider",
      "base_url": "https://api.provider.com",
      "credentials": [
        {"key": "API_KEY", "value": "your-actual-key"},
        {"key": "API_SECRET", "value": "your-actual-secret"}
      ],
      "endpoints": [
        {"name": "your-action", "method": "POST", "path": "/api/endpoint"}
      ],
      "header_rules": [
        {"name": "Authorization", "expression": "credential:API_KEY", "priority": 1}
      ]
    }
  ]
}
```

Load and configure:
```go
package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    
    "github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
    "github.com/google/uuid"
)

type ProviderConfig struct {
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    BaseURL     string `json:"base_url"`
    Credentials []struct {
        Key   string `json:"key"`
        Value string `json:"value"`
    } `json:"credentials"`
    Endpoints []struct {
        Name   string `json:"name"`
        Method string `json:"method"`
        Path   string `json:"path"`
    } `json:"endpoints"`
    HeaderRules []struct {
        Name       string `json:"name"`
        Expression string `json:"expression"`
        Priority   int    `json:"priority"`
    } `json:"header_rules"`
}

type Config struct {
    Providers []ProviderConfig `json:"providers"`
}

func loadProvidersFromConfig(orch *orchestrator.Orchestrator, configFile string) error {
    data, err := ioutil.ReadFile(configFile)
    if err != nil {
        return err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return err
    }
    
    admin := orch.Admin()
    
    for _, p := range config.Providers {
        providerID := uuid.New().String()
        
        // Create provider
        err = admin.CreateProvider(orchestrator.ProviderConfig{
            ID:          providerID,
            Name:        p.Name,
            DisplayName: p.DisplayName,
            BaseURL:     p.BaseURL,
            IsActive:    true,
        })
        if err != nil {
            return err
        }
        
        // Add credentials
        for _, c := range p.Credentials {
            err = admin.CreateCredential(orchestrator.CredentialConfig{
                ProviderID: providerID,
                Key:        c.Key,
                Value:      c.Value,
                Required:   true,
            })
            if err != nil {
                return err
            }
        }
        
        // Add endpoints
        for _, e := range p.Endpoints {
            err = admin.CreateEndpoint(orchestrator.EndpointConfig{
                ID:         uuid.New().String(),
                ProviderID: providerID,
                Name:       e.Name,
                Method:     e.Method,
                Path:       e.Path,
            })
            if err != nil {
                return err
            }
        }
        
        // Add header rules
        for _, h := range p.HeaderRules {
            err = admin.CreateHeaderRule(orchestrator.HeaderRuleConfig{
                ID:              uuid.New().String(),
                ProviderID:      providerID,
                HeaderName:      h.Name,
                ValueExpression: h.Expression,
                Priority:        h.Priority,
            })
            if err != nil {
                return err
            }
        }
        
        log.Printf("✅ Configured provider: %s", p.Name)
    }
    
    return nil
}
```

Run in your application:
```go
func main() {
    // ... after InitializeOrchestrator ...
    err := loadProvidersFromConfig(GetOrchestrator(), "config/providers.json")
    if err != nil {
        log.Fatal("Failed to load provider config:", err)
    }
    // ...
}
```

---

## Troubleshooting Guide for Agent

### Problem: "encryption key not set"

**Cause:** SetEncryptionKeyFromString not called before orchestrator.New()

**Solution:**
```go
// Move this BEFORE orchestrator.New()
orchestrator.SetEncryptionKeyFromString(os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY"))
orch, err := orchestrator.New(orchestrator.Config{DB: db})
```

### Problem: "table providers does not exist"

**Cause:** Migrations not run

**Solution:**
```go
import orchConfig "github.com/PayRam/api-orchestrator-go/config"
err := orchConfig.AutoMigrate(db)
```

### Problem: "table providers already exists" or migration conflicts

**Cause:** Your application already has tables with the same names (providers, credentials, etc.)

**Solution:** Use table prefix to avoid conflicts
```go
// Set table prefix in environment
export ORCHESTRATOR_TABLE_PREFIX=orch_

// Or in code
err := orchConfig.AutoMigrateWithOptions(db, orchConfig.AutoMigrateOptions{
    TablePrefix: "orch_",
})

// Create orchestrator with same prefix
orch, err := orchestrator.New(orchestrator.Config{
    DB:          db,
    TablePrefix: "orch_",
})
```

This creates: `orch_providers`, `orch_credentials`, etc. instead of conflicting table names.

See [Table Prefix Documentation](docs/TABLE_PREFIX.md) for more details.

### Problem: "provider not found: your-provider-name"

**Cause:** Provider not created in database

**Solution:**
```go
admin := orch.Admin()
admin.CreateProvider(orchestrator.ProviderConfig{
    Name:    "your-provider-name",
    BaseURL: "https://api.provider.com",
    // ...
})
```

### Problem: "endpoint not found: your-action-name"

**Cause:** Endpoint not created for provider

**Solution:**
```go
admin.CreateEndpoint(orchestrator.EndpointConfig{
    ProviderID: providerID,
    Name:       "your-action-name",
    ...
})
```

### Problem: Compilation errors on imports

**Cause:** Missing dependencies

**Solution:**
```bash
go mod tidy
go get github.com/PayRam/api-orchestrator-go
go get github.com/google/uuid
```

### Problem: Database connection errors

**Cause:** Using wrong database or database not accessible

**Solution:**
- Verify database connection in existing code works
- Ensure same `*gorm.DB` instance is passed to orchestrator
- Check PostgreSQL is running and accessible

---

## Success Criteria

Integration is successful when:

1. ✅ Application compiles without errors
2. ✅ Application starts without panics
3. ✅ Database contains 8 orchestrator tables
4. ✅ Can preview cURL commands via API
5. ✅ Can execute API calls through orchestrator
6. ✅ Responses have unified format
7. ✅ Logs show orchestrator activity
8. ✅ Credentials are encrypted in database

---

## Post-Integration Tasks

After successful integration:

1. **Configure Providers**
   - Use Admin API or configuration files
   - Add Banxa, Transak, MoonPay, etc. as needed
   - Store credentials securely

2. **Add More Endpoints**
   - get_order_status
   - cancel_order
   - get_payment_methods
   - etc.

3. **Implement Error Handling**
   - Retry logic for transient failures
   - Fallback to alternative providers
   - User-friendly error messages

4. **Add Monitoring**
   - Track API call success rates
   - Monitor latency per provider
   - Alert on high error rates

5. **Security Hardening**
   - Rotate encryption keys
   - Implement rate limiting
   - Add request validation
   - Log only non-sensitive data

6. **Documentation**
   - Document provider-specific configurations
   - Create runbook for common issues
   - Add API documentation for new endpoints

---

## Quick Reference Commands

```bash
# Check if tables exist
psql -U postgres -d your_db -c "\dt providers"

# View provider data
psql -U postgres -d your_db -c "SELECT name, is_active FROM providers;"

# View encrypted credentials (should be gibberish)
psql -U postgres -d your_db -c "SELECT key, encrypted_value FROM credentials LIMIT 1;"

# Test cURL generation
curl -X POST http://localhost:8080/api/v1/orchestrator/build-curl \
  -H "Content-Type: application/json" \
  -d '{"provider":"your-provider","action":"your-action","params":{}}'

# Check logs
tail -f /path/to/logs/app.log | grep orchestrator
```

---

## Integration Checklist

Use this checklist to track progress:

**Library Integration:**
- [ ] Added library dependency (`go get github.com/PayRam/api-orchestrator-go`)
- [ ] Located existing database initialization
- [ ] Added ORCHESTRATOR_ENCRYPTION_KEY environment variable
- [ ] Created InitializeOrchestrator function
- [ ] Called InitializeOrchestrator in main()
- [ ] Verified orchestrator tables created in database
- [ ] Created OrchestratorHandler with generic methods
- [ ] Registered orchestrator routes
- [ ] Application compiles and starts successfully

**Provider Configuration (Separate Step):**
- [ ] Configured at least one provider via Admin API or script
- [ ] Verified provider exists in database
- [ ] Verified credentials encrypted in database
- [ ] Verified endpoints configured

**Testing:**
- [ ] Tested build-curl endpoint
- [ ] Tested build-curl-detailed endpoint
- [ ] Tested actual API call
- [ ] Verified response format
- [ ] Checked logs for orchestrator activity

**Documentation:**
- [ ] Documented integration in project README
- [ ] Documented provider configuration process
- [ ] Committed changes to version control

---

## Final Notes for AI Agent

**Important Principles:**

1. **Use Existing Database** - Never create a new database, always use the existing one
2. **Preserve Existing Code** - Don't modify existing functionality, only add new code
3. **Follow Patterns** - Match the existing project's code style and patterns
4. **Verify Each Step** - Test after each major change
5. **Log Everything** - Use the existing logger to track orchestrator activity
6. **Handle Errors** - Check all error returns, don't ignore them
7. **Security First** - Never log sensitive data (credentials, API keys)
8. **Document Changes** - Add comments explaining orchestrator integration

**Integration Philosophy:**

The orchestrator library is designed to **augment** your application, not replace existing functionality. It should:
- Use your existing database connection
- Use your existing logger
- Follow your existing routing patterns
- Integrate seamlessly with your existing code

Think of it as adding a new capability to your application, like adding a new feature module.

**When in Doubt:**

1. Check the INTEGRATION_GUIDE.md for detailed examples
2. Use BuildCurl() to test without executing real API calls
3. Start with one provider (Banxa) and add more after it works
4. Test in development environment before production
5. Ask for human review if integration pattern is unclear

---

## Example Integration Summary

```
Project: mobile-backend-api
Language: Go 1.21
Framework: Gin
Database: PostgreSQL (existing)

Integration Changes (Code):
1. Added github.com/PayRam/api-orchestrator-go to go.mod
2. Created internal/orchestrator/setup.go (InitializeOrchestrator)
3. Created internal/handlers/orchestrator_handler.go (OrchestratorHandler)
4. Modified main.go to call InitializeOrchestrator
5. Modified routes/router.go to register orchestrator routes
6. Added .env variable: ORCHESTRATOR_ENCRYPTION_KEY

Provider Configuration (Separate):
- Configured providers via Admin API (domain-specific)
- Added credentials - encrypted in DB
- Created endpoints (actions specific to providers)
- Added header rules and response mappings

Testing:
✅ Application compiles
✅ Database tables created (8 new tables)
✅ Providers configured in database
✅ Build cURL works (generic endpoint)
✅ API calls execute successfully
✅ Responses have unified format

Next Steps:
- Configure providers for your domain (payment/shipping/SMS/etc.)
- Create domain-specific wrapper handlers if needed
- Implement error handling and retry logic
- Add monitoring and alerting
```

---

**End of Agentic Integration Prompt**

This document should provide complete guidance for AI agents to successfully integrate the API Orchestrator Go library into existing applications. Follow the steps sequentially, verify each step, and refer to the troubleshooting guide when issues arise.
