# API Orchestrator Go

A modular Go library following clean architecture principles for dynamically orchestrating HTTP API calls using metadata-driven configuration stored in PostgreSQL.

## 🎯 Overview

`api-orchestrator-go` enables you to:

- **Dynamically construct and execute external HTTP API calls** based on database-driven configuration
- **Support multiple providers** (e.g., Banxa, Transak) with different authentication and request formats
- **Normalize responses** using configurable JSONPath mappings
- **Plug-and-play architecture** that can be embedded in internal backend applications
- **Clean architecture** with clear separation of concerns (model, repository, service, orchestrator)

## 🏗️ Architecture

```
api-orchestrator-go/
├── model/                      # All DB models (one file per entity)
│   ├── provider.go
│   ├── provider_credential.go
│   ├── header_rule.go
│   ├── endpoint.go
│   ├── request_schema.go
│   ├── request_value.go
│   ├── strategy.go
│   └── response_mapping.go
├── repository/                 # Repository interfaces and implementations
│   ├── provider_repo.go        # Interface
│   ├── provider_repo_impl.go   # GORM implementation
│   ├── ...
├── service/                    # Business logic services
│   ├── provider_service.go     # Interface
│   ├── provider_service_impl.go # Implementation (calls other services)
│   ├── ...
├── orchestrator/               # Core orchestration logic
│   ├── context/                # Runtime context
│   ├── builder/                # Request and header building
│   ├── executor/               # HTTP execution
│   ├── response/               # Response mapping
│   ├── pipeline/               # Strategy pipeline
│   └── orchestrator.go         # Main orchestrator
├── config/                     # Configuration management
│   ├── config.go               # Env config
│   └── database.go             # DB connection
├── utils/                      # Utilities (JSONPath, transformations, logger)
├── testdata/                   # Sample fixtures
└── cmd/                        # Executables
```

## ✅ Implementation Complete

All core components have been implemented following clean architecture principles:

### 📦 Models (model/)

All DB models with GORM tags for PostgreSQL:

- **Provider**: Main API provider entity (e.g., Banxa, Transak)
- **ProviderCredential**: Authentication credentials storage
- **HeaderRule**: Dynamic HTTP header construction rules
- **Endpoint**: API endpoint definitions
- **RequestSchema**: Request payload structure definitions
- **RequestValue**: Field mappings and transformation rules
- **Strategy**: Pluggable strategies (signature generation, token refresh)
- **ResponseMapping**: Response normalization mappings

### 🗄️ Repository Layer (repository/)

Clean separation of interfaces and implementations:

- Each repository has `*_repo.go` (interface) and `*_repo_impl.go` (GORM implementation)
- `ProviderRepo`, `CredentialRepo`, `HeaderRuleRepo`, `EndpointRepo`, `StrategyRepo`, etc.
- All implementations use PostgreSQL via GORM

### 🔧 Service Layer (service/)

Business logic with dependency injection:

- Each service has `*_service.go` (interface) and `*_service_impl.go` (implementation)
- Services only call other services (never repositories directly, except their own)
- `ProviderService`, `CredentialService`, `EndpointService`, `StrategyService`, etc.
- Includes validation, error handling, and logging

### 🎭 Orchestrator (orchestrator/)

Complete orchestration pipeline:

- **Context**: Runtime context management with credentials, headers, computed values
- **Builder**: Dynamic request and header building from DB metadata
  - `RequestBuilder`: Constructs URL, query params, body
  - `HeaderBuilder`: Applies header rules with template support
- **Executor**: HTTP client for API execution
- **Response**: JSONPath-based response mapping and normalization
- **Pipeline**: Pluggable strategy execution pipeline
- **Orchestrator**: Main entry point tying everything together

### 🛠️ Infrastructure

- **Config**: Environment-based configuration management
- **Database**: GORM connection with auto-migration
- **Utils**: JSONPath extraction (gjson), transformations, logging (Zap)
- **Testdata**: Sample provider configs, requests, and responses

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 12+

### Installation

```bash
go get github.com/PayRam/api-orchestrator-go
```

### Table Name Conflicts

If your existing application already has tables named `providers`, `credentials`, `endpoints`, etc., you can configure a table prefix to avoid conflicts:

```go
// Run migrations with prefix
err := config.AutoMigrateWithOptions(db, config.AutoMigrateOptions{
    TablePrefix: "orch_", // Creates: orch_providers, orch_credentials, etc.
})

// Create orchestrator with same prefix
orch, err := orchestrator.New(orchestrator.Config{
    DB:          db,
    TablePrefix: "orch_",
})
```

See [Table Prefix Documentation](docs/TABLE_PREFIX.md) for detailed usage.

### Environment Variables

Create a `.env` file or set the following environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=api_orchestrator
DB_SSL_MODE=disable

# Logger Configuration
LOG_LEVEL=info
ENV=development
```

### Basic Usage

```go
package main

import (
    "log"

    "github.com/PayRam/api-orchestrator-go/config"
    "github.com/PayRam/api-orchestrator-go/orchestrator"
    "github.com/PayRam/api-orchestrator-go/repository"
    "github.com/PayRam/api-orchestrator-go/service"
    "github.com/PayRam/api-orchestrator-go/utils"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize logger
    if err := utils.InitLogger(cfg.Logger.Environment, cfg.Logger.Level); err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }
    defer utils.Sync()

    logger := utils.GetLogger()

    // Connect to database
    dbConn, err := config.NewConnection(&cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database")
    }
    defer dbConn.Close()

    // Run migrations
    if err := dbConn.AutoMigrate(); err != nil {
        logger.Fatal("Failed to run migrations")
    }

    // Initialize repositories
    db := dbConn.GetDB()
    providerRepo := repository.NewProviderRepo(db)
    credentialRepo := repository.NewCredentialRepo(db)
    endpointRepo := repository.NewEndpointRepo(db)
    headerRuleRepo := repository.NewHeaderRuleRepo(db)
    requestValueRepo := repository.NewRequestValueRepo(db)
    responseMappingRepo := repository.NewResponseMappingRepo(db)

    // Initialize services (services depend on other services, not repos)
    providerService := service.NewProviderService(providerRepo, logger)
    credentialService := service.NewCredentialService(credentialRepo, providerService, logger)
    endpointService := service.NewEndpointService(endpointRepo, providerService, logger)

    // Initialize orchestrator
    orch := orchestrator.New(
        providerService,
        credentialService,
        endpointService,
        headerRuleRepo,
        requestValueRepo,
        responseMappingRepo,
        logger,
    )

    // Make an API call
    params := map[string]interface{}{
        "account_reference": "user123",
        "source_currency":   "USD",
        "target_currency":   "BTC",
        "wallet_address":    "0x123...",
    }

    response, err := orch.Call("banxa", "create_widget_url", params)
    if err != nil {
        logger.Error("API call failed", zap.Error(err))
        return
    }

    logger.Info("API call successful",
        zap.Bool("success", response.Success),
        zap.Int("status_code", response.StatusCode),
        zap.Any("data", response.Data))
}
```

## 📊 Database Schema

The library automatically creates the following tables:

- `providers` - API provider definitions (Banxa, Transak, etc.)
- `provider_credentials` - Authentication credentials (API keys, secrets)
- `header_rules` - Dynamic header construction rules
- `endpoints` - API endpoint configurations (method, path, action)
- `request_schemas` - Request structure schemas (JSON schemas)
- `request_values` - Field mappings and transformations
- `strategies` - Pluggable strategy definitions (signature, auth, transform)
- `response_mappings` - Response normalization rules (JSONPath mappings)

## 🎯 Key Features

### Dynamic Request Building
- Path parameter substitution: `/api/{id}` → `/api/123`
- Query parameter building from metadata
- Body construction from field mappings
- Support for static, parameter, credential, and strategy-based values

### Header Construction
- Priority-based header rules
- Template support: `Bearer {token}`, `{api_key}`
- Multiple value types: static, credential, strategy, dynamic
- Automatic credential injection from context

### Response Normalization
- JSONPath-based field extraction using gjson
- Configurable field mappings: `data.checkout_url` → `widget_url`
- Default values for missing fields
- Required field validation
- Transformation functions (to_upper, to_lower, format_date, etc.)

### Strategy Pipeline
- Pluggable strategy execution
- Support for signature generation, token refresh, payload transformation
- Strategy results stored in context for use in headers/body
- Extensible architecture for custom strategies

### Clean Architecture
- **Repository** layer: Data access interfaces and implementations
- **Service** layer: Business logic (services only call other services)
- **Orchestrator** layer: Request orchestration and coordination
- Clear dependency injection and separation of concerns

## 🔄 Roadmap

### Phase 2: Request Builder
- Dynamic URL construction with path/query/body support
- Parameter validation and transformation
- Request body building from schemas

### Phase 3: Header Construction Pipeline
- Header rule evaluation engine
- Strategy execution for dynamic values
- Template interpolation

### Phase 4: Strategy Engine
- Signature generation strategies
- Token refresh strategies
- Custom transformation strategies

### Phase 5: Response Normalization
- JSONPath-based field extraction
- Response mapping engine
- Error handling and standardization

### Phase 6: Orchestrator Service
- High-level orchestration API
- Public entrypoint: `orchestrator.Call(provider, action, params)`
- HTTP client integration

### Phase 7: Testing & Tooling
- Database seeding utilities
- CLI tools for testing
- Comprehensive test suite

## 🎯 Target Usage (Future)

```go
// Initialize orchestrator
orchestrator := orchestrator.New(db)

// Make a call to any configured provider
response, err := orchestrator.Call("banxa", "create_widget_url", map[string]interface{}{
    "amount": "100.00",
    "currency": "USD",
    "wallet_address": "0x123...",
})

if err != nil {
    log.Fatal(err)
}

// Access normalized response
fmt.Printf("Widget URL: %s\n", response.Data["widget_url"])
fmt.Printf("Status: %s\n", response.Data["status"])
```

## 🤝 Contributing

Contributions are welcome! This is currently in active development.

## 📝 License

MIT License

## 📧 Contact

For questions or support, please open an issue on GitHub.
