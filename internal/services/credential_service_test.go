package services

import (
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Mock credential repository for testing
type mockCredentialRepo struct {
	credentials map[string]*models.Credential
	byProvider  map[string][]*models.Credential
}

func newMockCredentialRepo() *mockCredentialRepo {
	return &mockCredentialRepo{
		credentials: make(map[string]*models.Credential),
		byProvider:  make(map[string][]*models.Credential),
	}
}

func (m *mockCredentialRepo) Create(credential *models.Credential) error {
	m.credentials[credential.ID] = credential
	m.byProvider[credential.ProviderID] = append(m.byProvider[credential.ProviderID], credential)
	return nil
}

func (m *mockCredentialRepo) FindByID(id string) (*models.Credential, error) {
	if cred, ok := m.credentials[id]; ok {
		return cred, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockCredentialRepo) FindByProviderID(providerID string) ([]*models.Credential, error) {
	return m.byProvider[providerID], nil
}

func (m *mockCredentialRepo) FindByProviderIDAndKey(providerID, key string) (*models.Credential, error) {
	for _, cred := range m.byProvider[providerID] {
		if cred.Key == key {
			return cred, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockCredentialRepo) Update(credential *models.Credential) error {
	m.credentials[credential.ID] = credential
	return nil
}

func (m *mockCredentialRepo) Delete(id string) error {
	delete(m.credentials, id)
	return nil
}

// Mock provider service for credential tests
type mockProviderServiceForCred struct {
	providers map[string]*models.Provider
}

func newMockProviderServiceForCred() *mockProviderServiceForCred {
	return &mockProviderServiceForCred{
		providers: map[string]*models.Provider{
			"provider-1": {ID: "provider-1", Name: "test_provider", IsActive: true},
		},
	}
}

func (m *mockProviderServiceForCred) CreateProvider(provider *models.Provider) error { return nil }
func (m *mockProviderServiceForCred) GetProviderByID(id string) (*models.Provider, error) {
	if p, ok := m.providers[id]; ok {
		return p, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockProviderServiceForCred) GetProviderByName(name string) (*models.Provider, error) {
	return nil, nil
}
func (m *mockProviderServiceForCred) GetAllActive() ([]*models.Provider, error)      { return nil, nil }
func (m *mockProviderServiceForCred) ListProviders() ([]*models.Provider, error)     { return nil, nil }
func (m *mockProviderServiceForCred) UpdateProvider(provider *models.Provider) error { return nil }
func (m *mockProviderServiceForCred) DeleteProvider(id string) error                 { return nil }

func newTestCredentialService() (*credentialServiceImpl, *mockCredentialRepo) {
	logger, _ := zap.NewDevelopment()
	repo := newMockCredentialRepo()
	providerService := newMockProviderServiceForCred()
	service := &credentialServiceImpl{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
	return service, repo
}

func TestCredentialService_CreateCredential(t *testing.T) {
	service, _ := newTestCredentialService()

	tests := []struct {
		name       string
		credential *models.Credential
		wantErr    bool
	}{
		{
			name: "valid_credential",
			credential: &models.Credential{
				ID:         "c1",
				ProviderID: "provider-1",
				Key:        "API_KEY",
				Value:      "secret123",
			},
			wantErr: false,
		},
		{
			name: "missing_key",
			credential: &models.Credential{
				ID:         "c2",
				ProviderID: "provider-1",
				Key:        "",
				Value:      "secret123",
			},
			wantErr: true,
		},
		{
			name: "missing_value",
			credential: &models.Credential{
				ID:         "c3",
				ProviderID: "provider-1",
				Key:        "API_KEY",
				Value:      "",
			},
			wantErr: true,
		},
		{
			name: "invalid_provider",
			credential: &models.Credential{
				ID:         "c4",
				ProviderID: "non-existent",
				Key:        "API_KEY",
				Value:      "secret123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateCredential(tt.credential)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCredentialService_CreateCredentialWithPlainValue(t *testing.T) {
	service, _ := newTestCredentialService()

	cred, err := service.CreateCredentialWithPlainValue(
		"provider-1",
		"API_KEY",
		"my-secret-api-key",
		"Test API Key",
		true,
	)
	if err != nil {
		t.Fatalf("CreateCredentialWithPlainValue failed: %v", err)
	}

	if cred.Key != "API_KEY" {
		t.Errorf("Expected key 'API_KEY', got '%s'", cred.Key)
	}

	// Value should be encrypted (not plain text)
	if cred.Value == "my-secret-api-key" {
		t.Error("Value should be encrypted, but it's stored as plain text")
	}

	// Should be able to decrypt it back
	decrypted, err := cred.GetDecryptedValue()
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	if decrypted != "my-secret-api-key" {
		t.Errorf("Expected decrypted value 'my-secret-api-key', got '%s'", decrypted)
	}

	t.Logf("Encrypted value: %s", cred.Value)
	t.Logf("Decrypted value: %s", decrypted)
}

func TestCredentialService_GetDecryptedCredentialsByProviderID(t *testing.T) {
	service, repo := newTestCredentialService()

	// Create credentials with encrypted values
	cred1 := &models.Credential{
		ID:         "c1",
		ProviderID: "provider-1",
		Key:        "API_KEY",
	}
	cred1.SetEncryptedValue("api-key-value")

	cred2 := &models.Credential{
		ID:         "c2",
		ProviderID: "provider-1",
		Key:        "API_SECRET",
	}
	cred2.SetEncryptedValue("api-secret-value")

	repo.Create(cred1)
	repo.Create(cred2)

	// Get decrypted credentials
	decrypted, err := service.GetDecryptedCredentialsByProviderID("provider-1")
	if err != nil {
		t.Fatalf("GetDecryptedCredentialsByProviderID failed: %v", err)
	}

	if len(decrypted) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(decrypted))
	}

	if decrypted["API_KEY"] != "api-key-value" {
		t.Errorf("Expected API_KEY='api-key-value', got '%s'", decrypted["API_KEY"])
	}

	if decrypted["API_SECRET"] != "api-secret-value" {
		t.Errorf("Expected API_SECRET='api-secret-value', got '%s'", decrypted["API_SECRET"])
	}

	t.Logf("Decrypted credentials: %v", decrypted)
}

func TestCredentialService_GetCredentialValue(t *testing.T) {
	service, repo := newTestCredentialService()

	// Create a credential with encrypted value
	cred := &models.Credential{
		ID:         "c1",
		ProviderID: "provider-1",
		Key:        "API_KEY",
	}
	cred.SetEncryptedValue("my-secret-value")
	repo.Create(cred)

	// Get the value (should be decrypted)
	value, err := service.GetCredentialValue("provider-1", "API_KEY")
	if err != nil {
		t.Fatalf("GetCredentialValue failed: %v", err)
	}

	if value != "my-secret-value" {
		t.Errorf("Expected 'my-secret-value', got '%s'", value)
	}
}

func TestCredentialService_UpdateCredentialValue(t *testing.T) {
	service, repo := newTestCredentialService()

	// Create initial credential
	cred := &models.Credential{
		ID:         "c1",
		ProviderID: "provider-1",
		Key:        "API_KEY",
	}
	cred.SetEncryptedValue("old-value")
	repo.Create(cred)

	// Update the value
	err := service.UpdateCredentialValue("c1", "new-secret-value")
	if err != nil {
		t.Fatalf("UpdateCredentialValue failed: %v", err)
	}

	// Verify the new value
	updated, _ := repo.FindByID("c1")
	decrypted, _ := updated.GetDecryptedValue()
	if decrypted != "new-secret-value" {
		t.Errorf("Expected 'new-secret-value', got '%s'", decrypted)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plainText string
	}{
		{"simple_text", "hello"},
		{"api_key", "sk_live_abc123xyz"},
		{"long_secret", "this-is-a-very-long-secret-key-that-should-still-work-correctly"},
		{"special_chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "日本語テスト🔐"},
		{"empty_string", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := models.EncryptValue(tc.plainText)
			if err != nil {
				t.Fatalf("EncryptValue failed: %v", err)
			}

			if tc.plainText != "" && encrypted == tc.plainText {
				t.Error("Encrypted value should not equal plain text")
			}

			decrypted, err := models.DecryptValue(encrypted)
			if err != nil {
				t.Fatalf("DecryptValue failed: %v", err)
			}

			if decrypted != tc.plainText {
				t.Errorf("Decrypted value mismatch. Expected '%s', got '%s'", tc.plainText, decrypted)
			}

			t.Logf("Plain: '%s' -> Encrypted: '%s' -> Decrypted: '%s'", tc.plainText, encrypted, decrypted)
		})
	}
}

func TestEncryptionUniqueness(t *testing.T) {
	// Same plaintext should produce different ciphertexts (due to random nonce)
	plainText := "same-secret"

	encrypted1, _ := models.EncryptValue(plainText)
	encrypted2, _ := models.EncryptValue(plainText)

	if encrypted1 == encrypted2 {
		t.Error("Same plaintext should produce different ciphertexts due to random nonce")
	}

	// But both should decrypt to the same value
	decrypted1, _ := models.DecryptValue(encrypted1)
	decrypted2, _ := models.DecryptValue(encrypted2)

	if decrypted1 != plainText || decrypted2 != plainText {
		t.Error("Both encrypted values should decrypt to the same plaintext")
	}

	t.Logf("Same plaintext, different ciphertexts:")
	t.Logf("  Encrypted 1: %s", encrypted1)
	t.Logf("  Encrypted 2: %s", encrypted2)
}

func TestBackwardCompatibility(t *testing.T) {
	// Plain text that was stored before encryption was implemented
	plainText := "old-unencrypted-value"

	// DecryptValue should return the original value if it can't decrypt
	decrypted, err := models.DecryptValue(plainText)
	if err != nil {
		t.Fatalf("DecryptValue failed on plain text: %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Expected '%s', got '%s'", plainText, decrypted)
	}
}
