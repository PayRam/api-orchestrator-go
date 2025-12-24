package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
)

// Mock implementations for testing

type mockEndpointService struct {
	endpoints map[string]*models.Endpoint
}

func (m *mockEndpointService) CreateEndpoint(endpoint *models.Endpoint) error { return nil }
func (m *mockEndpointService) GetEndpointByID(id string) (*models.Endpoint, error) {
	if ep, ok := m.endpoints[id]; ok {
		return ep, nil
	}
	return nil, nil
}
func (m *mockEndpointService) GetEndpointByProviderAndName(providerName, name string) (*models.Endpoint, error) {
	key := providerName + ":" + name
	if ep, ok := m.endpoints[key]; ok {
		return ep, nil
	}
	return nil, nil
}
func (m *mockEndpointService) GetEndpointsByProviderID(providerID string) ([]*models.Endpoint, error) {
	return nil, nil
}
func (m *mockEndpointService) UpdateEndpoint(endpoint *models.Endpoint) error { return nil }
func (m *mockEndpointService) DeleteEndpoint(id string) error                 { return nil }

type mockProviderServiceForBuilder struct {
	providers map[string]*models.Provider
}

func (m *mockProviderServiceForBuilder) CreateProvider(provider *models.Provider) error { return nil }
func (m *mockProviderServiceForBuilder) GetProviderByID(id string) (*models.Provider, error) {
	return nil, nil
}
func (m *mockProviderServiceForBuilder) GetProviderByName(name string) (*models.Provider, error) {
	if p, ok := m.providers[name]; ok {
		return p, nil
	}
	return nil, nil
}
func (m *mockProviderServiceForBuilder) GetAllActive() ([]*models.Provider, error)      { return nil, nil }
func (m *mockProviderServiceForBuilder) ListProviders() ([]*models.Provider, error)     { return nil, nil }
func (m *mockProviderServiceForBuilder) UpdateProvider(provider *models.Provider) error { return nil }
func (m *mockProviderServiceForBuilder) DeleteProvider(id string) error                 { return nil }

type mockRequestSchemaRepo struct {
	schemas map[string][]*models.RequestSchema
}

func (m *mockRequestSchemaRepo) Create(schema *models.RequestSchema) error { return nil }
func (m *mockRequestSchemaRepo) FindByID(id string) (*models.RequestSchema, error) {
	return nil, nil
}
func (m *mockRequestSchemaRepo) FindByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	if schemas, ok := m.schemas[endpointID]; ok {
		return schemas, nil
	}
	return []*models.RequestSchema{}, nil
}
func (m *mockRequestSchemaRepo) FindByEndpointIDAndLocation(endpointID, location string) ([]*models.RequestSchema, error) {
	var result []*models.RequestSchema
	if schemas, ok := m.schemas[endpointID]; ok {
		for _, s := range schemas {
			if s.ParamLocation == location {
				result = append(result, s)
			}
		}
	}
	return result, nil
}
func (m *mockRequestSchemaRepo) FindRequiredByEndpointID(endpointID string) ([]*models.RequestSchema, error) {
	var result []*models.RequestSchema
	if schemas, ok := m.schemas[endpointID]; ok {
		for _, s := range schemas {
			if s.Required {
				result = append(result, s)
			}
		}
	}
	return result, nil
}
func (m *mockRequestSchemaRepo) Update(schema *models.RequestSchema) error { return nil }
func (m *mockRequestSchemaRepo) Delete(id string) error                    { return nil }

type mockRequestValueRepo struct {
	values map[string][]*models.RequestValue
}

func (m *mockRequestValueRepo) Create(value *models.RequestValue) error { return nil }
func (m *mockRequestValueRepo) FindByID(id string) (*models.RequestValue, error) {
	return nil, nil
}
func (m *mockRequestValueRepo) FindBySchemaID(schemaID string) ([]*models.RequestValue, error) {
	if values, ok := m.values[schemaID]; ok {
		return values, nil
	}
	return []*models.RequestValue{}, nil
}
func (m *mockRequestValueRepo) Update(value *models.RequestValue) error { return nil }
func (m *mockRequestValueRepo) Delete(id string) error                  { return nil }

func newTestRequestBuilderService() *requestBuilderServiceImpl {
	logger, _ := zap.NewDevelopment()
	return &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{endpoints: make(map[string]*models.Endpoint)},
		providerService:  &mockProviderServiceForBuilder{providers: make(map[string]*models.Provider)},
		schemaRepo:       &mockRequestSchemaRepo{schemas: make(map[string][]*models.RequestSchema)},
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}
}

func TestRequestBuilderService_BuildRequest_Simple(t *testing.T) {
	service := newTestRequestBuilderService()

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "get_orders",
		Method:     "GET",
		Path:       "/api/orders",
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{"api_key": "test123"},
		Input:        map[string]interface{}{},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{"Authorization": "Bearer test"},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if request.Method != "GET" {
		t.Errorf("Expected method GET, got %s", request.Method)
	}
	if request.URL != "https://api.banxa.com/api/orders" {
		t.Errorf("Expected URL 'https://api.banxa.com/api/orders', got '%s'", request.URL)
	}
	if request.Headers["Authorization"] != "Bearer test" {
		t.Errorf("Expected Authorization header, got %v", request.Headers)
	}
}

func TestRequestBuilderService_BuildRequest_WithPathParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "order_id", ParamLocation: models.ParamLocationPath, ParamType: models.ParamTypeString},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "get_order",
		Method:     "GET",
		Path:       "/api/orders/{order_id}",
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{},
		Input:        map[string]interface{}{"order_id": "12345"},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	expectedURL := "https://api.banxa.com/api/orders/12345"
	if request.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL)
	}
	if request.Metadata.PathParams["order_id"] != "12345" {
		t.Errorf("Expected path param order_id='12345', got %v", request.Metadata.PathParams)
	}
}

func TestRequestBuilderService_BuildRequest_WithQueryParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "page", ParamLocation: models.ParamLocationQuery, ParamType: models.ParamTypeInteger},
				{ID: "s2", EndpointID: "endpoint-1", ParamName: "limit", ParamLocation: models.ParamLocationQuery, ParamType: models.ParamTypeInteger, DefaultValue: "10"},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "list_orders",
		Method:     "GET",
		Path:       "/api/orders",
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{},
		Input:        map[string]interface{}{"page": 2},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if request.QueryParams["page"] != "2" {
		t.Errorf("Expected page=2, got %s", request.QueryParams["page"])
	}
	if request.QueryParams["limit"] != "10" {
		t.Errorf("Expected default limit=10, got %s", request.QueryParams["limit"])
	}
}

func TestRequestBuilderService_BuildRequest_WithBody(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "amount", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeNumber, Required: true},
				{ID: "s2", EndpointID: "endpoint-1", ParamName: "currency", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString, Required: true},
				{ID: "s3", EndpointID: "endpoint-1", ParamName: "wallet_address", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "create_order",
		Method:     "POST",
		Path:       "/api/orders",
	}

	ctx := &RequestBuildContext{
		Provider:    provider,
		Endpoint:    endpoint,
		Credentials: map[string]string{},
		Input: map[string]interface{}{
			"amount":         100.50,
			"currency":       "USD",
			"wallet_address": "0x123abc",
		},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	if request.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", request.Headers)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(request.Body, &body); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if body["amount"] != 100.50 {
		t.Errorf("Expected amount=100.50, got %v", body["amount"])
	}
	if body["currency"] != "USD" {
		t.Errorf("Expected currency='USD', got %v", body["currency"])
	}
	if body["wallet_address"] != "0x123abc" {
		t.Errorf("Expected wallet_address='0x123abc', got %v", body["wallet_address"])
	}
}

func TestRequestBuilderService_BuildRequest_EndpointBaseURLOverride(t *testing.T) {
	service := newTestRequestBuilderService()

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "sandbox_orders",
		Method:     "GET",
		Path:       "/api/orders",
		BaseURL:    "https://sandbox.banxa.com", // Override
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{},
		Input:        map[string]interface{}{},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	expectedURL := "https://sandbox.banxa.com/api/orders"
	if request.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL)
	}
}

func TestRequestBuilderService_GenerateCurl_GET(t *testing.T) {
	service := newTestRequestBuilderService()

	request := &BuiltRequest{
		Method:      "GET",
		URL:         "https://api.test.com/orders",
		Headers:     map[string]string{"Authorization": "Bearer token123"},
		QueryParams: map[string]string{"page": "1", "limit": "10"},
	}

	curl := service.GenerateCurl(request)

	if !strings.Contains(curl, "curl") {
		t.Error("Expected curl command to start with 'curl'")
	}
	if !strings.Contains(curl, "-H 'Authorization: Bearer token123'") {
		t.Error("Expected Authorization header in curl")
	}
	if !strings.Contains(curl, "https://api.test.com/orders") {
		t.Error("Expected URL in curl")
	}
	// GET should not have -X GET (it's the default)
	if strings.Contains(curl, "-X GET") {
		t.Error("GET method should not include -X flag")
	}

	t.Logf("Generated cURL: %s", curl)
}

func TestRequestBuilderService_GenerateCurl_POST(t *testing.T) {
	service := newTestRequestBuilderService()

	body, _ := json.Marshal(map[string]interface{}{
		"amount":   100,
		"currency": "USD",
	})

	request := &BuiltRequest{
		Method:      "POST",
		URL:         "https://api.test.com/orders",
		Headers:     map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
		QueryParams: map[string]string{},
		Body:        body,
	}

	curl := service.GenerateCurl(request)

	if !strings.Contains(curl, "-X POST") {
		t.Error("Expected -X POST in curl")
	}
	if !strings.Contains(curl, "-d '") {
		t.Error("Expected -d flag with body in curl")
	}
	if !strings.Contains(curl, "amount") {
		t.Error("Expected body content in curl")
	}

	t.Logf("Generated cURL: %s", curl)
}

func TestRequestBuilderService_ValidateRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "amount", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeNumber, Required: true},
				{ID: "s2", EndpointID: "endpoint-1", ParamName: "currency", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString, Required: true},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name:    "all_required_provided",
			input:   map[string]interface{}{"amount": 100, "currency": "USD"},
			wantErr: false,
		},
		{
			name:    "missing_required_amount",
			input:   map[string]interface{}{"currency": "USD"},
			wantErr: true,
		},
		{
			name:    "missing_all_required",
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &RequestBuildContext{
				Provider:     &models.Provider{ID: "p1"},
				Endpoint:     &models.Endpoint{ID: "endpoint-1"},
				Input:        tt.input,
				Credentials:  map[string]string{},
				Intermediate: map[string]interface{}{},
			}

			err := service.ValidateRequest(ctx)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRequestBuilderService_ConvertValue(t *testing.T) {
	service := newTestRequestBuilderService()

	tests := []struct {
		name      string
		value     interface{}
		paramType string
		expected  interface{}
		wantErr   bool
	}{
		{"string_to_string", "hello", models.ParamTypeString, "hello", false},
		{"int_to_string", 123, models.ParamTypeString, "123", false},
		{"string_to_int", "456", models.ParamTypeInteger, 456, false},
		{"float_to_int", 78.9, models.ParamTypeInteger, 78, false},
		{"int_to_number", 100, models.ParamTypeNumber, 100.0, false},
		{"string_to_number", "3.14", models.ParamTypeNumber, 3.14, false},
		{"string_to_bool_true", "true", models.ParamTypeBoolean, true, false},
		{"string_to_bool_false", "false", models.ParamTypeBoolean, false, false},
		{"bool_to_bool", true, models.ParamTypeBoolean, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.convertValue(tt.value, tt.paramType)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestRequestBuilderService_BuildRequest_WithCredentials(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "api_key", ParamLocation: models.ParamLocationBody, ParamType: models.ParamTypeString},
			},
		},
	}
	valueRepo := &mockRequestValueRepo{
		values: map[string][]*models.RequestValue{
			"s1": {
				{ID: "v1", SchemaID: "s1", SourceType: models.SourceTypeCredential, SourceKey: "api_key"},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: valueRepo,
		logger:           logger,
	}

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "auth_test",
		Method:     "POST",
		Path:       "/api/auth",
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{"api_key": "secret123"},
		Input:        map[string]interface{}{},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(request.Body, &body); err != nil {
		t.Fatalf("Failed to unmarshal body: %v", err)
	}

	if body["api_key"] != "secret123" {
		t.Errorf("Expected api_key='secret123' from credentials, got %v", body["api_key"])
	}
}

func TestRequestBuilderService_BuildRequest_MultiplePathParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	schemaRepo := &mockRequestSchemaRepo{
		schemas: map[string][]*models.RequestSchema{
			"endpoint-1": {
				{ID: "s1", EndpointID: "endpoint-1", ParamName: "provider_id", ParamLocation: models.ParamLocationPath, ParamType: models.ParamTypeString},
				{ID: "s2", EndpointID: "endpoint-1", ParamName: "order_id", ParamLocation: models.ParamLocationPath, ParamType: models.ParamTypeString},
			},
		},
	}
	service := &requestBuilderServiceImpl{
		endpointService:  &mockEndpointService{},
		providerService:  &mockProviderServiceForBuilder{},
		schemaRepo:       schemaRepo,
		requestValueRepo: &mockRequestValueRepo{values: make(map[string][]*models.RequestValue)},
		logger:           logger,
	}

	provider := &models.Provider{
		ID:      "provider-1",
		Name:    "banxa",
		BaseURL: "https://api.banxa.com",
	}
	endpoint := &models.Endpoint{
		ID:         "endpoint-1",
		ProviderID: "provider-1",
		Name:       "get_provider_order",
		Method:     "GET",
		Path:       "/api/providers/{provider_id}/orders/{order_id}",
	}

	ctx := &RequestBuildContext{
		Provider:     provider,
		Endpoint:     endpoint,
		Credentials:  map[string]string{},
		Input:        map[string]interface{}{"provider_id": "banxa", "order_id": "order-123"},
		Intermediate: map[string]interface{}{},
		Headers:      map[string]string{},
	}

	request, err := service.BuildRequest(ctx)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}

	expectedURL := "https://api.banxa.com/api/providers/banxa/orders/order-123"
	if request.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, request.URL)
	}
}
