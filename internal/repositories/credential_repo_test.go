package repositories

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCredentialTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the required models
	err = db.AutoMigrate(&models.Provider{}, &models.Credential{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestProviderForCredential(t *testing.T, db *gorm.DB) *models.Provider {
	provider := &models.Provider{
		ID:          "provider-cred-test",
		Name:        "test_provider",
		DisplayName: "Test Provider",
		BaseURL:     "https://api.test.com",
		IsActive:    true,
	}
	if err := db.Create(provider).Error; err != nil {
		t.Fatalf("Failed to create test provider: %v", err)
	}
	return provider
}

func TestCredentialRepo_Create(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	credential := &models.Credential{
		ID:          "cred-1",
		ProviderID:  provider.ID,
		Key:         "API_KEY",
		Value:       "secret123",
		Required:    true,
		Description: "API Key for authentication",
	}

	err := repo.Create(credential)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify
	found, err := repo.FindByID(credential.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Key != "API_KEY" {
		t.Errorf("Expected key 'API_KEY', got '%s'", found.Key)
	}
	if found.Value != "secret123" {
		t.Errorf("Expected value 'secret123', got '%s'", found.Value)
	}
}

func TestCredentialRepo_FindByID(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	credential := &models.Credential{
		ID:         "cred-find",
		ProviderID: provider.ID,
		Key:        "SECRET",
		Value:      "mysecret",
	}
	db.Create(credential)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing_credential", "cred-find", false},
		{"non-existing_credential", "non-existent", true},
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
				if found.Key != "SECRET" {
					t.Errorf("Expected key 'SECRET', got '%s'", found.Key)
				}
			}
		})
	}
}

func TestCredentialRepo_FindByProviderID(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	// Create multiple credentials for the provider
	credentials := []*models.Credential{
		{ID: "c1", ProviderID: provider.ID, Key: "API_KEY", Value: "key1"},
		{ID: "c2", ProviderID: provider.ID, Key: "API_SECRET", Value: "secret1"},
		{ID: "c3", ProviderID: provider.ID, Key: "WEBHOOK_SECRET", Value: "webhook1"},
	}
	for _, c := range credentials {
		db.Create(c)
	}

	found, err := repo.FindByProviderID(provider.ID)
	if err != nil {
		t.Fatalf("FindByProviderID failed: %v", err)
	}
	if len(found) != 3 {
		t.Errorf("Expected 3 credentials, got %d", len(found))
	}
}

func TestCredentialRepo_FindByProviderIDAndKey(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	credential := &models.Credential{
		ID:         "cred-key-find",
		ProviderID: provider.ID,
		Key:        "API_KEY",
		Value:      "myapikey123",
	}
	db.Create(credential)

	tests := []struct {
		name       string
		providerID string
		key        string
		wantErr    bool
	}{
		{"existing_key", provider.ID, "API_KEY", false},
		{"non-existing_key", provider.ID, "NON_EXISTENT", true},
		{"wrong_provider", "wrong-provider", "API_KEY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByProviderIDAndKey(tt.providerID, tt.key)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if found.Value != "myapikey123" {
					t.Errorf("Expected value 'myapikey123', got '%s'", found.Value)
				}
			}
		})
	}
}

func TestCredentialRepo_Update(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	credential := &models.Credential{
		ID:         "cred-update",
		ProviderID: provider.ID,
		Key:        "API_KEY",
		Value:      "oldvalue",
		Required:   false,
	}
	db.Create(credential)

	// Update
	credential.Value = "newvalue"
	credential.Required = true
	err := repo.Update(credential)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	found, _ := repo.FindByID(credential.ID)
	if found.Value != "newvalue" {
		t.Errorf("Expected value 'newvalue', got '%s'", found.Value)
	}
	if !found.Required {
		t.Error("Expected Required to be true")
	}
}

func TestCredentialRepo_Delete(t *testing.T) {
	db := setupCredentialTestDB(t)
	provider := createTestProviderForCredential(t, db)
	repo := NewCredentialRepo(db)

	credential := &models.Credential{
		ID:         "cred-delete",
		ProviderID: provider.ID,
		Key:        "TO_DELETE",
		Value:      "deleteme",
	}
	db.Create(credential)

	err := repo.Delete(credential.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify
	_, err = repo.FindByID(credential.ID)
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}
