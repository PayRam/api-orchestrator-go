package repositories

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupStrategyTestDB creates an in-memory SQLite database for testing
func setupStrategyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&models.Strategy{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestStrategyRepo_Create(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	strategy := &models.Strategy{
		ID:           "strategy-1",
		Name:         "banxa_hmac",
		StrategyType: "HMAC",
		Config:       datatypes.JSON([]byte(`{"algorithm": "SHA256"}`)),
	}

	err := repo.Create(strategy)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify strategy was created
	found, err := repo.FindByID(strategy.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if found.Name != strategy.Name {
		t.Errorf("expected name %s, got %s", strategy.Name, found.Name)
	}
	if found.StrategyType != strategy.StrategyType {
		t.Errorf("expected type %s, got %s", strategy.StrategyType, found.StrategyType)
	}
}

func TestStrategyRepo_FindByID(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create a strategy first
	strategy := &models.Strategy{
		ID:           "strategy-2",
		Name:         "transak_signature",
		StrategyType: "SIGNATURE",
	}
	_ = repo.Create(strategy)

	tests := []struct {
		name    string
		id      string
		want    *models.Strategy
		wantErr bool
	}{
		{
			name:    "existing strategy",
			id:      "strategy-2",
			want:    strategy,
			wantErr: false,
		},
		{
			name:    "non-existing strategy",
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

func TestStrategyRepo_FindByName(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create a strategy first
	strategy := &models.Strategy{
		ID:           "strategy-3",
		Name:         "moonpay_basic_auth",
		StrategyType: "BASIC_AUTH",
	}
	_ = repo.Create(strategy)

	tests := []struct {
		name     string
		findName string
		want     *models.Strategy
		wantErr  bool
	}{
		{
			name:     "existing strategy by name",
			findName: "moonpay_basic_auth",
			want:     strategy,
			wantErr:  false,
		},
		{
			name:     "non-existing strategy by name",
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

func TestStrategyRepo_FindByType(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create multiple strategies with different types
	strategies := []*models.Strategy{
		{ID: "s1", Name: "hmac_1", StrategyType: "HMAC"},
		{ID: "s2", Name: "hmac_2", StrategyType: "HMAC"},
		{ID: "s3", Name: "signature_1", StrategyType: "SIGNATURE"},
		{ID: "s4", Name: "basic_auth_1", StrategyType: "BASIC_AUTH"},
	}

	for _, s := range strategies {
		_ = repo.Create(s)
	}

	// Fetch HMAC strategies
	result, err := repo.FindByType("HMAC")
	if err != nil {
		t.Fatalf("FindByType() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("FindByType() got %d strategies, want 2", len(result))
	}

	// Verify all returned strategies are HMAC type
	for _, s := range result {
		if s.StrategyType != "HMAC" {
			t.Errorf("FindByType() returned strategy with type %s, want HMAC", s.StrategyType)
		}
	}
}

func TestStrategyRepo_List(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create multiple strategies
	strategies := []*models.Strategy{
		{ID: "list-1", Name: "strategy_1", StrategyType: "HMAC"},
		{ID: "list-2", Name: "strategy_2", StrategyType: "SIGNATURE"},
		{ID: "list-3", Name: "strategy_3", StrategyType: "API_KEY"},
	}

	for _, s := range strategies {
		_ = repo.Create(s)
	}

	result, err := repo.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("List() got %d strategies, want 3", len(result))
	}
}

func TestStrategyRepo_Update(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create a strategy first
	strategy := &models.Strategy{
		ID:           "strategy-update",
		Name:         "original_name",
		StrategyType: "HMAC",
		Config:       datatypes.JSON([]byte(`{"algorithm": "SHA256"}`)),
	}
	_ = repo.Create(strategy)

	// Update the strategy
	strategy.StrategyType = "SIGNATURE"
	strategy.Config = datatypes.JSON([]byte(`{"algorithm": "SHA512"}`))

	err := repo.Update(strategy)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify the update
	found, _ := repo.FindByID(strategy.ID)
	if found.StrategyType != "SIGNATURE" {
		t.Errorf("Update() StrategyType = %v, want 'SIGNATURE'", found.StrategyType)
	}
}

func TestStrategyRepo_Delete(t *testing.T) {
	db := setupStrategyTestDB(t)
	repo := NewStrategyRepo(db)

	// Create a strategy first
	strategy := &models.Strategy{
		ID:           "strategy-delete",
		Name:         "delete_me",
		StrategyType: "HMAC",
	}
	_ = repo.Create(strategy)

	// Delete the strategy
	err := repo.Delete(strategy.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify the strategy is deleted
	_, err = repo.FindByID(strategy.ID)
	if err == nil {
		t.Error("Delete() expected error when finding deleted strategy, got nil")
	}
}
