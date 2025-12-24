package repositories

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&models.Provider{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestProviderRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	provider := &models.Provider{
		ID:           "test-provider-1",
		Name:         "banxa",
		DisplayName:  "Banxa",
		PipelineType: "standard",
		IsActive:     true,
	}

	err := repo.Create(provider)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify provider was created
	found, err := repo.FindByID(provider.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Name != provider.Name {
		t.Errorf("expected name %s, got %s", provider.Name, found.Name)
	}
	if found.DisplayName != provider.DisplayName {
		t.Errorf("expected display_name %s, got %s", provider.DisplayName, found.DisplayName)
	}
}

func TestProviderRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create a provider first
	provider := &models.Provider{
		ID:          "test-provider-2",
		Name:        "transak",
		DisplayName: "Transak",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	tests := []struct {
		name    string
		id      string
		want    *models.Provider
		wantErr bool
	}{
		{
			name:    "existing provider",
			id:      "test-provider-2",
			want:    provider,
			wantErr: false,
		},
		{
			name:    "non-existing provider",
			id:      "non-existent-id",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.want.ID {
				t.Errorf("FindByID() got ID = %v, want %v", got.ID, tt.want.ID)
			}
		})
	}
}

func TestProviderRepo_FindByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create a provider first
	provider := &models.Provider{
		ID:          "test-provider-3",
		Name:        "moonpay",
		DisplayName: "MoonPay",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	tests := []struct {
		name     string
		findName string
		want     *models.Provider
		wantErr  bool
	}{
		{
			name:     "existing provider by name",
			findName: "moonpay",
			want:     provider,
			wantErr:  false,
		},
		{
			name:     "non-existing provider by name",
			findName: "non-existent-name",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByName(tt.findName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want.Name {
				t.Errorf("FindByName() got Name = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}

func TestProviderRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create multiple providers
	providers := []*models.Provider{
		{ID: "provider-1", Name: "provider1", DisplayName: "Provider 1", IsActive: true},
		{ID: "provider-2", Name: "provider2", DisplayName: "Provider 2", IsActive: false},
		{ID: "provider-3", Name: "provider3", DisplayName: "Provider 3", IsActive: true},
	}

	for _, p := range providers {
		_ = repo.Create(p)
	}

	// List all providers
	list, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 3 {
		t.Errorf("List() got %d providers, want 3", len(list))
	}
}

func TestProviderRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create a provider first
	provider := &models.Provider{
		ID:          "test-provider-4",
		Name:        "wyre",
		DisplayName: "Wyre",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	// Update the provider
	provider.DisplayName = "Wyre Updated"
	provider.IsActive = false

	err := repo.Update(provider)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	found, _ := repo.FindByID(provider.ID)
	if found.DisplayName != "Wyre Updated" {
		t.Errorf("Update() DisplayName = %v, want 'Wyre Updated'", found.DisplayName)
	}
	if found.IsActive != false {
		t.Errorf("Update() IsActive = %v, want false", found.IsActive)
	}
}

func TestProviderRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create a provider first
	provider := &models.Provider{
		ID:          "test-provider-5",
		Name:        "ramp",
		DisplayName: "Ramp",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	// Delete the provider
	err := repo.Delete(provider.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify the provider is deleted (soft delete)
	_, err = repo.FindByID(provider.ID)
	if err == nil {
		t.Error("Delete() expected error when finding deleted provider, got nil")
	}
}

func TestProviderRepo_CreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db)

	// Create a provider
	provider := &models.Provider{
		ID:          "test-provider-dup",
		Name:        "duplicate-test",
		DisplayName: "Duplicate Test",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	// Try to create another with the same ID
	duplicate := &models.Provider{
		ID:          "test-provider-dup", // same ID
		Name:        "different-name",
		DisplayName: "Different Name",
		IsActive:    true,
	}

	err := repo.Create(duplicate)
	if err == nil {
		t.Error("Create() expected error when creating duplicate ID, got nil")
	}
}
