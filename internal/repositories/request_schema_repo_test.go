package repositories

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRequestSchemaTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the required models
	err = db.AutoMigrate(&models.Provider{}, &models.Endpoint{}, &models.RequestSchema{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestProviderAndEndpoint(t *testing.T, db *gorm.DB) (*models.Provider, *models.Endpoint) {
	provider := &models.Provider{
		ID:          "provider-1",
		Name:        "test_provider",
		DisplayName: "Test Provider",
		BaseURL:     "https://api.test.com",
		IsActive:    true,
	}
	if err := db.Create(provider).Error; err != nil {
		t.Fatalf("Failed to create test provider: %v", err)
	}

	endpoint := &models.Endpoint{
		ID:          "endpoint-1",
		ProviderID:  provider.ID,
		Name:        "create_order",
		Method:      "POST",
		Path:        "/api/v1/orders/{order_id}",
		Description: "Create a new order",
	}
	if err := db.Create(endpoint).Error; err != nil {
		t.Fatalf("Failed to create test endpoint: %v", err)
	}

	return provider, endpoint
}

func TestRequestSchemaRepo_Create(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	schema := &models.RequestSchema{
		ID:            "schema-1",
		EndpointID:    endpoint.ID,
		ParamName:     "amount",
		ParamLocation: models.ParamLocationBody,
		ParamType:     models.ParamTypeNumber,
		Required:      true,
		Description:   "Transaction amount",
	}

	err := repo.Create(schema)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify
	found, err := repo.FindByID(schema.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.ParamName != "amount" {
		t.Errorf("Expected param name 'amount', got '%s'", found.ParamName)
	}
	if found.ParamLocation != models.ParamLocationBody {
		t.Errorf("Expected location 'body', got '%s'", found.ParamLocation)
	}
}

func TestRequestSchemaRepo_FindByID(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	schema := &models.RequestSchema{
		ID:            "schema-find",
		EndpointID:    endpoint.ID,
		ParamName:     "currency",
		ParamLocation: models.ParamLocationQuery,
		ParamType:     models.ParamTypeString,
		Required:      false,
		DefaultValue:  "USD",
	}
	db.Create(schema)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing_schema", "schema-find", false},
		{"non-existing_schema", "non-existent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if found.ParamName != "currency" {
					t.Errorf("Expected param name 'currency', got '%s'", found.ParamName)
				}
			}
		})
	}
}

func TestRequestSchemaRepo_FindByEndpointID(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	// Create multiple schemas for the endpoint
	schemas := []*models.RequestSchema{
		{ID: "s1", EndpointID: endpoint.ID, ParamName: "amount", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeNumber},
		{ID: "s2", EndpointID: endpoint.ID, ParamName: "currency", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString},
		{ID: "s3", EndpointID: endpoint.ID, ParamName: "order_id", ParamLocation: models.ParamLocationPath, ParamType: models.ParamTypeString},
	}
	for _, s := range schemas {
		db.Create(s)
	}

	found, err := repo.FindByEndpointID(endpoint.ID)
	if err != nil {
		t.Fatalf("FindByEndpointID failed: %v", err)
	}
	if len(found) != 3 {
		t.Errorf("Expected 3 schemas, got %d", len(found))
	}
}

func TestRequestSchemaRepo_FindByEndpointIDAndLocation(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	// Create schemas with different locations
	schemas := []*models.RequestSchema{
		{ID: "s1", EndpointID: endpoint.ID, ParamName: "amount", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeNumber},
		{ID: "s2", EndpointID: endpoint.ID, ParamName: "currency", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString},
		{ID: "s3", EndpointID: endpoint.ID, ParamName: "order_id", ParamLocation: models.ParamLocationPath, ParamType: models.ParamTypeString},
		{ID: "s4", EndpointID: endpoint.ID, ParamName: "api_version", ParamLocation: models.ParamLocationQuery, ParamType: models.ParamTypeString},
	}
	for _, s := range schemas {
		db.Create(s)
	}

	tests := []struct {
		name     string
		location string
		expected int
	}{
		{"body_params", models.ParamLocationBody, 2},
		{"path_params", models.ParamLocationPath, 1},
		{"query_params", models.ParamLocationQuery, 1},
		{"header_params", models.ParamLocationHeader, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByEndpointIDAndLocation(endpoint.ID, tt.location)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(found) != tt.expected {
				t.Errorf("Expected %d schemas, got %d", tt.expected, len(found))
			}
		})
	}
}

func TestRequestSchemaRepo_FindRequiredByEndpointID(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	// Create schemas with different required values
	schemas := []*models.RequestSchema{
		{ID: "s1", EndpointID: endpoint.ID, ParamName: "amount", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeNumber, Required: true},
		{ID: "s2", EndpointID: endpoint.ID, ParamName: "currency", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString, Required: true},
		{ID: "s3", EndpointID: endpoint.ID, ParamName: "note", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString, Required: false},
	}
	for _, s := range schemas {
		db.Create(s)
	}

	found, err := repo.FindRequiredByEndpointID(endpoint.ID)
	if err != nil {
		t.Fatalf("FindRequiredByEndpointID failed: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("Expected 2 required schemas, got %d", len(found))
	}
}

func TestRequestSchemaRepo_Update(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	schema := &models.RequestSchema{
		ID:            "schema-update",
		EndpointID:    endpoint.ID,
		ParamName:     "old_name",
		ParamLocation: models.ParamLocationBody,
		ParamType:     models.ParamTypeString,
		Required:      false,
	}
	db.Create(schema)

	// Update
	schema.ParamName = "new_name"
	schema.Required = true
	err := repo.Update(schema)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	found, _ := repo.FindByID(schema.ID)
	if found.ParamName != "new_name" {
		t.Errorf("Expected param name 'new_name', got '%s'", found.ParamName)
	}
	if !found.Required {
		t.Error("Expected Required to be true")
	}
}

func TestRequestSchemaRepo_Delete(t *testing.T) {
	db := setupRequestSchemaTestDB(t)
	_, endpoint := createTestProviderAndEndpoint(t, db)
	repo := NewRequestSchemaRepo(db)

	schema := &models.RequestSchema{
		ID:            "schema-delete",
		EndpointID:    endpoint.ID,
		ParamName:     "to_delete",
		ParamLocation: models.ParamLocationBody,
		ParamType:     models.ParamTypeString,
	}
	db.Create(schema)

	err := repo.Delete(schema.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	_, err = repo.FindByID(schema.ID)
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}
