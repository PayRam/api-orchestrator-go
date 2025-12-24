package repositories

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupHeaderRuleTestDB creates an in-memory SQLite database for testing
func setupHeaderRuleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&models.HeaderRule{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestHeaderRuleRepo_Create(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	rule := &models.HeaderRule{
		ID:              "header-rule-1",
		ProviderID:      "provider-1",
		HeaderName:      "Authorization",
		ValueExpression: "credential:api_key",
		Priority:        1,
	}

	err := repo.Create(rule)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify rule was created
	found, err := repo.FindByID(rule.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.HeaderName != rule.HeaderName {
		t.Errorf("expected header name %s, got %s", rule.HeaderName, found.HeaderName)
	}
	if found.ValueExpression != rule.ValueExpression {
		t.Errorf("expected value expression %s, got %s", rule.ValueExpression, found.ValueExpression)
	}
}

func TestHeaderRuleRepo_FindByID(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	// Create a rule first
	rule := &models.HeaderRule{
		ID:              "header-rule-2",
		ProviderID:      "provider-1",
		HeaderName:      "X-API-Key",
		ValueExpression: "static:test-key",
		Priority:        0,
	}
	_ = repo.Create(rule)

	tests := []struct {
		name    string
		id      string
		want    *models.HeaderRule
		wantErr bool
	}{
		{
			name:    "existing rule",
			id:      "header-rule-2",
			want:    rule,
			wantErr: false,
		},
		{
			name:    "non-existing rule",
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

func TestHeaderRuleRepo_FindByProviderID(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	// Create multiple rules with different priorities
	rules := []*models.HeaderRule{
		{ID: "rule-1", ProviderID: "provider-1", HeaderName: "Header-A", ValueExpression: "static:a", Priority: 3},
		{ID: "rule-2", ProviderID: "provider-1", HeaderName: "Header-B", ValueExpression: "static:b", Priority: 1},
		{ID: "rule-3", ProviderID: "provider-1", HeaderName: "Header-C", ValueExpression: "static:c", Priority: 2},
		{ID: "rule-4", ProviderID: "provider-2", HeaderName: "Header-D", ValueExpression: "static:d", Priority: 0},
	}

	for _, r := range rules {
		_ = repo.Create(r)
	}

	// Fetch rules for provider-1
	result, err := repo.FindByProviderID("provider-1")
	if err != nil {
		t.Fatalf("FindByProviderID() error = %v", err)
	}

	// Should have 3 rules for provider-1
	if len(result) != 3 {
		t.Errorf("FindByProviderID() got %d rules, want 3", len(result))
	}

	// Verify priority ordering (ASC)
	expectedOrder := []string{"rule-2", "rule-3", "rule-1"} // priority 1, 2, 3
	for i, r := range result {
		if r.ID != expectedOrder[i] {
			t.Errorf("FindByProviderID() priority order wrong at index %d: got %s, want %s",
				i, r.ID, expectedOrder[i])
		}
	}
}

func TestHeaderRuleRepo_Update(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	// Create a rule first
	rule := &models.HeaderRule{
		ID:              "header-rule-update",
		ProviderID:      "provider-1",
		HeaderName:      "Content-Type",
		ValueExpression: "static:application/json",
		Priority:        0,
	}
	_ = repo.Create(rule)

	// Update the rule
	rule.ValueExpression = "static:application/xml"
	rule.Priority = 5

	err := repo.Update(rule)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	found, _ := repo.FindByID(rule.ID)
	if found.ValueExpression != "static:application/xml" {
		t.Errorf("Update() ValueExpression = %v, want 'static:application/xml'", found.ValueExpression)
	}
	if found.Priority != 5 {
		t.Errorf("Update() Priority = %v, want 5", found.Priority)
	}
}

func TestHeaderRuleRepo_Delete(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	// Create a rule first
	rule := &models.HeaderRule{
		ID:              "header-rule-delete",
		ProviderID:      "provider-1",
		HeaderName:      "X-Delete-Me",
		ValueExpression: "static:delete",
		Priority:        0,
	}
	_ = repo.Create(rule)

	// Delete the rule
	err := repo.Delete(rule.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify the rule is deleted
	_, err = repo.FindByID(rule.ID)
	if err == nil {
		t.Error("Delete() expected error when finding deleted rule, got nil")
	}
}

func TestHeaderRuleRepo_PriorityOrdering(t *testing.T) {
	db := setupHeaderRuleTestDB(t)
	repo := NewHeaderRuleRepo(db)

	// Create rules with specific priorities to test ordering
	rules := []*models.HeaderRule{
		{ID: "priority-10", ProviderID: "test-provider", HeaderName: "H10", ValueExpression: "s:10", Priority: 10},
		{ID: "priority-0", ProviderID: "test-provider", HeaderName: "H0", ValueExpression: "s:0", Priority: 0},
		{ID: "priority-5", ProviderID: "test-provider", HeaderName: "H5", ValueExpression: "s:5", Priority: 5},
		{ID: "priority--1", ProviderID: "test-provider", HeaderName: "H-1", ValueExpression: "s:-1", Priority: -1},
	}

	for _, r := range rules {
		_ = repo.Create(r)
	}

	result, err := repo.FindByProviderID("test-provider")
	if err != nil {
		t.Fatalf("FindByProviderID() error = %v", err)
	}

	// Verify ascending priority order: -1, 0, 5, 10
	expectedPriorities := []int{-1, 0, 5, 10}
	for i, r := range result {
		if r.Priority != expectedPriorities[i] {
			t.Errorf("Priority at index %d: got %d, want %d", i, r.Priority, expectedPriorities[i])
		}
	}
}
