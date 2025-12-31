package main

import (
	"log"
	"os"

	"github.com/PayRam/api-orchestrator-go/config"
	"github.com/PayRam/api-orchestrator-go/pkg/orchestrator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Example 1: Using table prefix to avoid conflicts with existing tables
func ExampleWithTablePrefix() {
	// Connect to existing database
	dsn := "host=localhost user=postgres password=password dbname=myapp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Set encryption key
	err = orchestrator.SetEncryptionKeyFromString("your-32-character-secret-key!")
	if err != nil {
		log.Fatal("Failed to set encryption key:", err)
	}

	// Run migrations with prefix to avoid conflicts
	// This will create: orch_providers, orch_credentials, orch_endpoints, etc.
	err = config.AutoMigrateWithOptions(db, config.AutoMigrateOptions{
		TablePrefix: "orch_",
	})
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create orchestrator with same prefix
	orch, err := orchestrator.New(orchestrator.Config{
		DB:          db,
		TablePrefix: "orch_",
	})
	if err != nil {
		log.Fatal("Failed to create orchestrator:", err)
	}

	log.Println("✅ Orchestrator initialized with table prefix 'orch_'")
	log.Println("Tables created: orch_providers, orch_credentials, orch_endpoints, etc.")

	// Use orchestrator...
	_ = orch
}

// Example 2: Without prefix (default behavior)
func ExampleWithoutTablePrefix() {
	dsn := "host=localhost user=postgres password=password dbname=orchestrator port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = orchestrator.SetEncryptionKeyFromString("your-32-character-secret-key!")
	if err != nil {
		log.Fatal("Failed to set encryption key:", err)
	}

	// Run migrations without prefix
	// This will create: providers, credentials, endpoints, etc.
	err = config.AutoMigrate(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create orchestrator without prefix
	orch, err := orchestrator.New(orchestrator.Config{
		DB: db,
	})
	if err != nil {
		log.Fatal("Failed to create orchestrator:", err)
	}

	log.Println("✅ Orchestrator initialized with default table names")
	log.Println("Tables created: providers, credentials, endpoints, etc.")

	_ = orch
}

// Example 3: Custom prefix
func ExampleWithCustomPrefix() {
	dsn := "host=localhost user=postgres password=password dbname=myapp port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = orchestrator.SetEncryptionKeyFromString("your-32-character-secret-key!")
	if err != nil {
		log.Fatal("Failed to set encryption key:", err)
	}

	// Use custom prefix
	customPrefix := "api_orchestrator_"
	err = config.AutoMigrateWithOptions(db, config.AutoMigrateOptions{
		TablePrefix: customPrefix,
	})
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	orch, err := orchestrator.New(orchestrator.Config{
		DB:          db,
		TablePrefix: customPrefix,
	})
	if err != nil {
		log.Fatal("Failed to create orchestrator:", err)
	}

	log.Printf("✅ Orchestrator initialized with table prefix '%s'", customPrefix)
	log.Println("Tables created: api_orchestrator_providers, api_orchestrator_credentials, etc.")

	_ = orch
}

// Example 4: Using environment variable for prefix
func ExampleWithEnvVarPrefix() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = orchestrator.SetEncryptionKeyFromString(os.Getenv("ORCHESTRATOR_ENCRYPTION_KEY"))
	if err != nil {
		log.Fatal("Failed to set encryption key:", err)
	}

	// Get prefix from environment variable (default to empty)
	tablePrefix := os.Getenv("ORCHESTRATOR_TABLE_PREFIX") // e.g., "orch_"

	err = config.AutoMigrateWithOptions(db, config.AutoMigrateOptions{
		TablePrefix: tablePrefix,
	})
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	orch, err := orchestrator.New(orchestrator.Config{
		DB:          db,
		TablePrefix: tablePrefix,
	})
	if err != nil {
		log.Fatal("Failed to create orchestrator:", err)
	}

	if tablePrefix == "" {
		log.Println("✅ Orchestrator initialized with default table names")
	} else {
		log.Printf("✅ Orchestrator initialized with table prefix '%s'", tablePrefix)
	}

	_ = orch
}

func main() {
	log.Println("Example 1: With Table Prefix")
	// ExampleWithTablePrefix()

	log.Println("\nExample 2: Without Table Prefix")
	// ExampleWithoutTablePrefix()

	log.Println("\nExample 3: With Custom Prefix")
	// ExampleWithCustomPrefix()

	log.Println("\nExample 4: With Environment Variable Prefix")
	// ExampleWithEnvVarPrefix()

	log.Println("\nUncomment the examples above to run them")
}
