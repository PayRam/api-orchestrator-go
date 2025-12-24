package services

import (
	"errors"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
)

// mockProviderRepo is a mock implementation of ProviderRepo for testing
type mockProviderRepo struct {
	providers map[string]*models.Provider
	nameIndex map[string]string // maps name to ID
}

func newMockProviderRepo() *mockProviderRepo {
	return &mockProviderRepo{
		providers: make(map[string]*models.Provider),
		nameIndex: make(map[string]string),
	}
}

func (m *mockProviderRepo) Create(provider *models.Provider) error {
	if _, exists := m.providers[provider.ID]; exists {
		return errors.New("provider already exists")
	}
	m.providers[provider.ID] = provider
	m.nameIndex[provider.Name] = provider.ID
	return nil
}

func (m *mockProviderRepo) FindByID(id string) (*models.Provider, error) {
	if provider, exists := m.providers[id]; exists {
		return provider, nil
	}
	return nil, errors.New("provider not found")
}

func (m *mockProviderRepo) FindByName(name string) (*models.Provider, error) {
	if id, exists := m.nameIndex[name]; exists {
		return m.providers[id], nil
	}
	return nil, errors.New("provider not found")
}

func (m *mockProviderRepo) FindAllActive() ([]*models.Provider, error) {
	result := make([]*models.Provider, 0)
	for _, p := range m.providers {
		if p.IsActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProviderRepo) List() ([]*models.Provider, error) {
	result := make([]*models.Provider, 0, len(m.providers))
	for _, p := range m.providers {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProviderRepo) Update(provider *models.Provider) error {
	if _, exists := m.providers[provider.ID]; !exists {
		return errors.New("provider not found")
	}
	// Update name index if name changed
	for name, id := range m.nameIndex {
		if id == provider.ID && name != provider.Name {
			delete(m.nameIndex, name)
			break
		}
	}
	m.providers[provider.ID] = provider
	m.nameIndex[provider.Name] = provider.ID
	return nil
}

func (m *mockProviderRepo) Delete(id string) error {
	if provider, exists := m.providers[id]; exists {
		delete(m.nameIndex, provider.Name)
		delete(m.providers, id)
		return nil
	}
	return errors.New("provider not found")
}

func TestProviderService_CreateProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	tests := []struct {
		name     string
		provider *models.Provider
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid provider",
			provider: &models.Provider{
				ID:          "provider-1",
				Name:        "banxa",
				DisplayName: "Banxa",
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			provider: &models.Provider{
				ID:          "provider-2",
				Name:        "",
				DisplayName: "No Name",
				IsActive:    true,
			},
			wantErr: true,
			errMsg:  "provider name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateProvider(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("CreateProvider() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestProviderService_GetProviderByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create a test provider
	testProvider := &models.Provider{
		ID:          "test-id-1",
		Name:        "test-provider",
		DisplayName: "Test Provider",
		IsActive:    true,
	}
	_ = repo.Create(testProvider)

	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{
			name:    "existing provider",
			id:      "test-id-1",
			want:    "test-provider",
			wantErr: false,
		},
		{
			name:    "non-existing provider",
			id:      "non-existent",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetProviderByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProviderByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.want {
				t.Errorf("GetProviderByID() got Name = %v, want %v", got.Name, tt.want)
			}
		})
	}
}

func TestProviderService_GetProviderByName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create a test provider
	testProvider := &models.Provider{
		ID:          "test-id-2",
		Name:        "moonpay",
		DisplayName: "MoonPay",
		IsActive:    true,
	}
	_ = repo.Create(testProvider)

	tests := []struct {
		name     string
		findName string
		wantID   string
		wantErr  bool
	}{
		{
			name:     "existing provider by name",
			findName: "moonpay",
			wantID:   "test-id-2",
			wantErr:  false,
		},
		{
			name:     "non-existing provider by name",
			findName: "nonexistent",
			wantID:   "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetProviderByName(tt.findName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProviderByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("GetProviderByName() got ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func TestProviderService_ListProviders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create multiple providers
	providers := []*models.Provider{
		{ID: "id-1", Name: "provider1", DisplayName: "Provider 1", IsActive: true},
		{ID: "id-2", Name: "provider2", DisplayName: "Provider 2", IsActive: false},
	}

	for _, p := range providers {
		_ = repo.Create(p)
	}

	list, err := service.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}

	if len(list) != 2 {
		t.Errorf("ListProviders() got %d providers, want 2", len(list))
	}
}

func TestProviderService_UpdateProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create a provider first
	provider := &models.Provider{
		ID:          "update-test-id",
		Name:        "original-name",
		DisplayName: "Original Name",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	// Update the provider
	provider.DisplayName = "Updated Name"
	provider.IsActive = false

	err := service.UpdateProvider(provider)
	if err != nil {
		t.Fatalf("UpdateProvider() error = %v", err)
	}

	// Verify the update
	updated, _ := service.GetProviderByID(provider.ID)
	if updated.DisplayName != "Updated Name" {
		t.Errorf("UpdateProvider() DisplayName = %v, want 'Updated Name'", updated.DisplayName)
	}
	if updated.IsActive != false {
		t.Errorf("UpdateProvider() IsActive = %v, want false", updated.IsActive)
	}
}

func TestProviderService_DeleteProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create a provider first
	provider := &models.Provider{
		ID:          "delete-test-id",
		Name:        "delete-me",
		DisplayName: "Delete Me",
		IsActive:    true,
	}
	_ = repo.Create(provider)

	// Delete the provider
	err := service.DeleteProvider(provider.ID)
	if err != nil {
		t.Fatalf("DeleteProvider() error = %v", err)
	}

	// Verify the provider is deleted
	_, err = service.GetProviderByID(provider.ID)
	if err == nil {
		t.Error("DeleteProvider() expected error when getting deleted provider, got nil")
	}
}

func TestProviderService_CreateProviderValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Test empty name validation
	provider := &models.Provider{
		ID:          "validation-test",
		Name:        "",
		DisplayName: "Test",
		IsActive:    true,
	}

	err := service.CreateProvider(provider)
	if err == nil {
		t.Error("CreateProvider() expected validation error for empty name, got nil")
	}

	if err.Error() != "provider name cannot be empty" {
		t.Errorf("CreateProvider() error = %v, want 'provider name cannot be empty'", err)
	}
}

func TestProviderService_GetActiveProviders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := newMockProviderRepo()
	service := NewProviderService(repo, logger)

	// Create providers with different active states
	providers := []*models.Provider{
		{ID: "active-1", Name: "active1", DisplayName: "Active 1", IsActive: true},
		{ID: "inactive-1", Name: "inactive1", DisplayName: "Inactive 1", IsActive: false},
		{ID: "active-2", Name: "active2", DisplayName: "Active 2", IsActive: true},
	}

	for _, p := range providers {
		_ = repo.Create(p)
	}

	// List all and manually filter active (since service doesn't have GetActiveProviders yet)
	list, err := service.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}

	activeCount := 0
	for _, p := range list {
		if p.IsActive {
			activeCount++
		}
	}

	if activeCount != 2 {
		t.Errorf("Expected 2 active providers, got %d", activeCount)
	}
}
