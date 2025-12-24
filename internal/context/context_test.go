package context

import (
	"testing"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
)

func TestNewContext(t *testing.T) {
	provider := models.Provider{
		ID:          "test-provider-id",
		Name:        "test-provider",
		DisplayName: "Test Provider",
	}
	action := "create_widget"
	input := map[string]interface{}{
		"wallet_address": "0x123",
		"crypto_type":    "BTC",
	}

	ctx := NewContext(provider, action, input)

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	if ctx.Provider.ID != provider.ID {
		t.Errorf("expected provider ID %s, got %s", provider.ID, ctx.Provider.ID)
	}

	if ctx.Action != action {
		t.Errorf("expected action %s, got %s", action, ctx.Action)
	}

	if ctx.Input["wallet_address"] != "0x123" {
		t.Errorf("expected input wallet_address 0x123, got %v", ctx.Input["wallet_address"])
	}

	if ctx.RequestID.String() == "" {
		t.Error("expected RequestID to be set")
	}

	if ctx.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}

	if ctx.Credentials == nil {
		t.Error("expected Credentials map to be initialized")
	}

	if ctx.Intermediate == nil {
		t.Error("expected Intermediate map to be initialized")
	}

	if ctx.Headers == nil {
		t.Error("expected Headers map to be initialized")
	}
}

func TestContext_SetAndGet(t *testing.T) {
	ctx := NewContext(models.Provider{}, "test", nil)

	// Test Set and Get
	ctx.Set("signature", "abc123")
	val := ctx.Get("signature")
	if val != "abc123" {
		t.Errorf("expected 'abc123', got %v", val)
	}

	// Test Get non-existent key
	val = ctx.Get("non_existent")
	if val != nil {
		t.Errorf("expected nil for non-existent key, got %v", val)
	}

	// Test GetOk
	val, ok := ctx.GetOk("signature")
	if !ok {
		t.Error("expected ok to be true for existing key")
	}
	if val != "abc123" {
		t.Errorf("expected 'abc123', got %v", val)
	}

	val, ok = ctx.GetOk("non_existent")
	if ok {
		t.Error("expected ok to be false for non-existent key")
	}
}

func TestContext_Credentials(t *testing.T) {
	ctx := NewContext(models.Provider{}, "test", nil)

	// Test SetCredential and GetCredential
	ctx.SetCredential("API_KEY", "secret123")
	ctx.SetCredential("API_SECRET", "supersecret")

	val, ok := ctx.GetCredential("API_KEY")
	if !ok {
		t.Error("expected credential to exist")
	}
	if val != "secret123" {
		t.Errorf("expected 'secret123', got %s", val)
	}

	// Test non-existent credential
	val, ok = ctx.GetCredential("NON_EXISTENT")
	if ok {
		t.Error("expected credential to not exist")
	}
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}

func TestContext_Headers(t *testing.T) {
	ctx := NewContext(models.Provider{}, "test", nil)

	// Test SetHeader and GetHeader
	ctx.SetHeader("Authorization", "Bearer token123")
	ctx.SetHeader("Content-Type", "application/json")

	val, ok := ctx.GetHeader("Authorization")
	if !ok {
		t.Error("expected header to exist")
	}
	if val != "Bearer token123" {
		t.Errorf("expected 'Bearer token123', got %s", val)
	}

	// Test MergeHeaders
	newHeaders := map[string]string{
		"X-Custom-Header": "custom-value",
		"Content-Type":    "text/plain", // Should overwrite
	}
	ctx.MergeHeaders(newHeaders)

	val, ok = ctx.GetHeader("X-Custom-Header")
	if !ok || val != "custom-value" {
		t.Errorf("expected 'custom-value', got %s", val)
	}

	val, ok = ctx.GetHeader("Content-Type")
	if !ok || val != "text/plain" {
		t.Errorf("expected 'text/plain' (overwritten), got %s", val)
	}

	// Authorization should still exist
	val, ok = ctx.GetHeader("Authorization")
	if !ok || val != "Bearer token123" {
		t.Errorf("expected 'Bearer token123' to still exist, got %s", val)
	}
}

func TestContext_Body(t *testing.T) {
	ctx := NewContext(models.Provider{}, "test", nil)

	// Test SetBody and GetBody with map
	body := map[string]interface{}{
		"amount":   100,
		"currency": "USD",
	}
	ctx.SetBody(body)

	result := ctx.GetBody()
	if result == nil {
		t.Fatal("expected body to be set")
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected body to be map[string]interface{}")
	}
	if resultMap["amount"] != 100 {
		t.Errorf("expected amount 100, got %v", resultMap["amount"])
	}

	// Test SetBody with struct
	type RequestBody struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	structBody := RequestBody{Name: "test", Value: 42}
	ctx.SetBody(structBody)

	result = ctx.GetBody()
	resultStruct, ok := result.(RequestBody)
	if !ok {
		t.Fatal("expected body to be RequestBody struct")
	}
	if resultStruct.Name != "test" {
		t.Errorf("expected name 'test', got %s", resultStruct.Name)
	}
}

func TestContext_Input(t *testing.T) {
	input := map[string]interface{}{
		"wallet_address": "0xabc",
		"amount":         100.5,
		"count":          42,
	}
	ctx := NewContext(models.Provider{}, "test", input)

	// Test GetInput
	val, ok := ctx.GetInput("wallet_address")
	if !ok {
		t.Error("expected input to exist")
	}
	if val != "0xabc" {
		t.Errorf("expected '0xabc', got %v", val)
	}

	// Test GetInput non-existent
	val, ok = ctx.GetInput("non_existent")
	if ok {
		t.Error("expected input to not exist")
	}

	// Test GetInputString
	str := ctx.GetInputString("wallet_address")
	if str != "0xabc" {
		t.Errorf("expected '0xabc', got %s", str)
	}

	// Test GetInputString with non-string value
	str = ctx.GetInputString("amount")
	if str != "" {
		t.Errorf("expected empty string for non-string value, got %s", str)
	}

	// Test GetInputString with non-existent key
	str = ctx.GetInputString("non_existent")
	if str != "" {
		t.Errorf("expected empty string, got %s", str)
	}
}

func TestContext_Duration(t *testing.T) {
	ctx := NewContext(models.Provider{}, "test", nil)

	// Sleep a tiny bit to ensure duration is measurable
	time.Sleep(10 * time.Millisecond)

	duration := ctx.Duration()
	if duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", duration)
	}
}

func TestContext_NilMaps(t *testing.T) {
	// Test that methods handle nil maps gracefully
	ctx := &Context{}

	// Get from nil Intermediate
	val := ctx.Get("key")
	if val != nil {
		t.Error("expected nil from nil Intermediate map")
	}

	// GetOk from nil Intermediate
	_, ok := ctx.GetOk("key")
	if ok {
		t.Error("expected false from nil Intermediate map")
	}

	// GetCredential from nil Credentials
	_, ok = ctx.GetCredential("key")
	if ok {
		t.Error("expected false from nil Credentials map")
	}

	// GetHeader from nil Headers
	_, ok = ctx.GetHeader("key")
	if ok {
		t.Error("expected false from nil Headers map")
	}

	// GetInput from nil Input
	_, ok = ctx.GetInput("key")
	if ok {
		t.Error("expected false from nil Input map")
	}

	// GetInputString from nil Input
	str := ctx.GetInputString("key")
	if str != "" {
		t.Error("expected empty string from nil Input map")
	}

	// Set should initialize Intermediate
	ctx.Set("key", "value")
	if ctx.Intermediate == nil {
		t.Error("expected Intermediate to be initialized after Set")
	}
	if ctx.Get("key") != "value" {
		t.Error("expected value after Set")
	}

	// SetCredential should initialize Credentials
	ctx2 := &Context{}
	ctx2.SetCredential("key", "value")
	if ctx2.Credentials == nil {
		t.Error("expected Credentials to be initialized after SetCredential")
	}

	// SetHeader should initialize Headers
	ctx3 := &Context{}
	ctx3.SetHeader("key", "value")
	if ctx3.Headers == nil {
		t.Error("expected Headers to be initialized after SetHeader")
	}

	// MergeHeaders should initialize Headers
	ctx4 := &Context{}
	ctx4.MergeHeaders(map[string]string{"key": "value"})
	if ctx4.Headers == nil {
		t.Error("expected Headers to be initialized after MergeHeaders")
	}
}

func TestOrchestratorContextAlias(t *testing.T) {
	// Test that OrchestratorContext alias works
	var ctx *OrchestratorContext = NewContext(models.Provider{Name: "test"}, "action", nil)

	if ctx.Provider.Name != "test" {
		t.Errorf("expected provider name 'test', got %s", ctx.Provider.Name)
	}
}
