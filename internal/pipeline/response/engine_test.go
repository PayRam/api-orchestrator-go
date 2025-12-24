package response

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// ==================================================
// Test ResponseContext
// ==================================================

func TestNewResponseContext(t *testing.T) {
	provider := models.Provider{
		ID:   "provider-1",
		Name: "test-provider",
	}
	rawResponse := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"id": "order-123",
		},
	}

	ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 200)

	assert.Equal(t, provider.ID, ctx.Provider.ID)
	assert.Equal(t, models.ActionCreateOrder, ctx.Action)
	assert.Equal(t, 200, ctx.StatusCode)
	assert.NotNil(t, ctx.RawResponse)
	assert.NotNil(t, ctx.RawResponseBytes)
	assert.NotNil(t, ctx.ExtractedValues)
	assert.NotNil(t, ctx.TransformedValues)
}

func TestNewResponseContextFromBytes(t *testing.T) {
	provider := models.Provider{
		ID:   "provider-1",
		Name: "test-provider",
	}
	rawBytes := json.RawMessage(`{"status":"success","data":{"id":"order-123"}}`)

	ctx, err := NewResponseContextFromBytes(provider, models.ActionCreateOrder, rawBytes, 200)

	require.NoError(t, err)
	assert.Equal(t, provider.ID, ctx.Provider.ID)
	assert.Equal(t, "success", ctx.RawResponse["status"])
}

func TestResponseContext_SetGetValues(t *testing.T) {
	ctx := NewResponseContext(models.Provider{}, models.ActionCreateOrder, nil, 200)

	// Test extracted values
	ctx.SetExtractedValue("order_id", "123")
	value, ok := ctx.GetExtractedValue("order_id")
	assert.True(t, ok)
	assert.Equal(t, "123", value)

	// Test transformed values
	ctx.SetTransformedValue("order_id", "ORDER-123")
	value, ok = ctx.GetTransformedValue("order_id")
	assert.True(t, ok)
	assert.Equal(t, "ORDER-123", value)

	// Test GetFinalValue (prefers transformed)
	value, ok = ctx.GetFinalValue("order_id")
	assert.True(t, ok)
	assert.Equal(t, "ORDER-123", value)

	// Test non-existent key
	_, ok = ctx.GetExtractedValue("non_existent")
	assert.False(t, ok)
}

func TestResponseContext_IsSuccess(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}

	for _, tt := range tests {
		ctx := NewResponseContext(models.Provider{}, "", nil, tt.statusCode)
		assert.Equal(t, tt.expected, ctx.IsSuccess(), "status code: %d", tt.statusCode)
	}
}

// ==================================================
// Test Pipeline Steps
// ==================================================

func TestJSONPathExtractStep(t *testing.T) {
	logger := newTestLogger()
	step := NewJSONPathExtractStep(logger)

	t.Run("extracts values successfully", func(t *testing.T) {
		rawResponse := map[string]interface{}{
			"data": map[string]interface{}{
				"order": map[string]interface{}{
					"id":     "order-123",
					"status": "pending",
					"amount": 100.50,
				},
			},
		}

		ctx := NewResponseContext(models.Provider{ID: "p1"}, models.ActionCreateOrder, rawResponse, 200)
		ctx.SetMappings([]*models.ResponseMapping{
			{TargetField: "order_id", SourceJSONPath: "data.order.id"},
			{TargetField: "status", SourceJSONPath: "data.order.status"},
			{TargetField: "amount", SourceJSONPath: "data.order.amount"},
		})

		err := step.Execute(ctx)
		require.NoError(t, err)

		orderID, _ := ctx.GetExtractedValue("order_id")
		assert.Equal(t, "order-123", orderID)

		status, _ := ctx.GetExtractedValue("status")
		assert.Equal(t, "pending", status)

		amount, _ := ctx.GetExtractedValue("amount")
		assert.Equal(t, 100.50, amount)
	})

	t.Run("uses default value when path not found", func(t *testing.T) {
		rawResponse := map[string]interface{}{
			"data": map[string]interface{}{},
		}

		ctx := NewResponseContext(models.Provider{ID: "p1"}, models.ActionCreateOrder, rawResponse, 200)
		ctx.SetMappings([]*models.ResponseMapping{
			{TargetField: "status", SourceJSONPath: "data.status", DefaultValue: "unknown"},
		})

		err := step.Execute(ctx)
		require.NoError(t, err)

		status, _ := ctx.GetExtractedValue("status")
		assert.Equal(t, "unknown", status)
	})

	t.Run("fails on missing required field", func(t *testing.T) {
		rawResponse := map[string]interface{}{
			"data": map[string]interface{}{},
		}

		ctx := NewResponseContext(models.Provider{ID: "p1"}, models.ActionCreateOrder, rawResponse, 200)
		ctx.SetMappings([]*models.ResponseMapping{
			{TargetField: "order_id", SourceJSONPath: "data.order.id", IsRequired: true},
		})

		err := step.Execute(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required fields")
	})
}

func TestTransformStep(t *testing.T) {
	logger := newTestLogger()
	step := NewTransformStep(logger)

	t.Run("applies transformations", func(t *testing.T) {
		ctx := NewResponseContext(models.Provider{ID: "p1"}, models.ActionCreateOrder, nil, 200)
		ctx.SetMappings([]*models.ResponseMapping{
			{TargetField: "status", Transform: "uppercase"},
			{TargetField: "amount", Transform: "to_cents"},
			{TargetField: "order_id", Transform: ""}, // No transform
		})
		ctx.SetExtractedValue("status", "pending")
		ctx.SetExtractedValue("amount", 100.50)
		ctx.SetExtractedValue("order_id", "123")

		err := step.Execute(ctx)
		require.NoError(t, err)

		status, _ := ctx.GetTransformedValue("status")
		assert.Equal(t, "PENDING", status)

		amount, _ := ctx.GetTransformedValue("amount")
		assert.Equal(t, int64(10050), amount)

		orderID, _ := ctx.GetTransformedValue("order_id")
		assert.Equal(t, "123", orderID)
	})
}

// ==================================================
// Test Full Pipeline
// ==================================================

func TestDefaultResponsePipeline_NormalizeCreateOrder(t *testing.T) {
	logger := newTestLogger()

	t.Run("normalizes create order response", func(t *testing.T) {
		provider := models.Provider{
			ID:   "provider-banxa",
			Name: "banxa",
		}

		rawResponse := map[string]interface{}{
			"data": map[string]interface{}{
				"order": map[string]interface{}{
					"id":             "ORD-123",
					"status":         "pending",
					"checkout_url":   "https://banxa.com/checkout/123",
					"fiat_amount":    100.0,
					"fiat_code":      "USD",
					"crypto_amount":  0.0025,
					"coin_code":      "BTC",
					"wallet_address": "bc1qtest...",
				},
			},
		}

		mappings := []*models.ResponseMapping{
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "data.order.id"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "status", SourceJSONPath: "data.order.status"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "checkout_url", SourceJSONPath: "data.order.checkout_url"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "fiat_amount", SourceJSONPath: "data.order.fiat_amount"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "fiat_currency", SourceJSONPath: "data.order.fiat_code", Transform: "uppercase"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "crypto_amount", SourceJSONPath: "data.order.crypto_amount"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "crypto_currency", SourceJSONPath: "data.order.coin_code", Transform: "uppercase"},
			{ProviderID: provider.ID, Action: models.ActionCreateOrder, TargetField: "wallet_address", SourceJSONPath: "data.order.wallet_address"},
		}

		pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
		ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 201)
		ctx.RequestID = "req-123"
		ctx.ExecutionTime = 500 * time.Millisecond

		result, err := pipeline.NormalizeCreateOrder(ctx)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, 201, result.StatusCode)
		assert.Equal(t, "banxa", result.ProviderName)
		assert.Equal(t, models.ActionCreateOrder, result.Action)
		assert.Equal(t, "req-123", result.RequestID)
		assert.Equal(t, "ORD-123", result.OrderID)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, "https://banxa.com/checkout/123", result.CheckoutURL)
		assert.Equal(t, 100.0, result.FiatAmount)
		assert.Equal(t, "USD", result.FiatCurrency)
		assert.Equal(t, 0.0025, result.CryptoAmount)
		assert.Equal(t, "BTC", result.CryptoCurrency)
		assert.Equal(t, "bc1qtest...", result.WalletAddress)
		assert.NotNil(t, result.RawResponse)
	})
}

func TestDefaultResponsePipeline_NormalizeGetQuote(t *testing.T) {
	logger := newTestLogger()

	provider := models.Provider{
		ID:   "provider-transak",
		Name: "transak",
	}

	rawResponse := map[string]interface{}{
		"response": map[string]interface{}{
			"quoteId":         "Q-456",
			"fiatAmount":      100.0,
			"fiatCurrency":    "usd",
			"cryptoAmount":    0.0025,
			"cryptoCurrency":  "btc",
			"conversionPrice": 40000.0,
			"totalFee":        2.50,
			"totalFeePercent": 2.5,
			"isBuyOrSell":     "BUY",
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "quote_id", SourceJSONPath: "response.quoteId"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "fiat_amount", SourceJSONPath: "response.fiatAmount"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "fiat_currency", SourceJSONPath: "response.fiatCurrency", Transform: "uppercase"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "crypto_amount", SourceJSONPath: "response.cryptoAmount"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "crypto_currency", SourceJSONPath: "response.cryptoCurrency", Transform: "uppercase"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "exchange_rate", SourceJSONPath: "response.conversionPrice"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "fee", SourceJSONPath: "response.totalFee"},
		{ProviderID: provider.ID, Action: models.ActionGetQuote, TargetField: "fee_percentage", SourceJSONPath: "response.totalFeePercent"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionGetQuote, rawResponse, 200)

	result, err := pipeline.NormalizeGetQuote(ctx)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "Q-456", result.QuoteID)
	assert.Equal(t, 100.0, result.FiatAmount)
	assert.Equal(t, "USD", result.FiatCurrency)
	assert.Equal(t, 0.0025, result.CryptoAmount)
	assert.Equal(t, "BTC", result.CryptoCurrency)
	assert.Equal(t, 40000.0, result.ExchangeRate)
	assert.Equal(t, 2.50, result.Fee)
	assert.Equal(t, 2.5, result.FeePercentage)
}

func TestDefaultResponsePipeline_NormalizeGetOrder(t *testing.T) {
	logger := newTestLogger()

	provider := models.Provider{
		ID:   "provider-banxa",
		Name: "banxa",
	}

	rawResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"order": map[string]interface{}{
				"id":          "ORD-789",
				"status":      "complete",
				"fiat_amount": 100.0,
				"fiat_code":   "USD",
				"coin_amount": 0.0025,
				"coin_code":   "BTC",
				"tx_hash":     "0xabc123...",
				"fee":         2.50,
			},
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "order_id", SourceJSONPath: "data.order.id"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "status", SourceJSONPath: "data.order.status"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "fiat_amount", SourceJSONPath: "data.order.fiat_amount"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "fiat_currency", SourceJSONPath: "data.order.fiat_code"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "crypto_amount", SourceJSONPath: "data.order.coin_amount"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "crypto_currency", SourceJSONPath: "data.order.coin_code"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "transaction_hash", SourceJSONPath: "data.order.tx_hash"},
		{ProviderID: provider.ID, Action: models.ActionGetOrder, TargetField: "fee", SourceJSONPath: "data.order.fee"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionGetOrder, rawResponse, 200)

	result, err := pipeline.NormalizeGetOrder(ctx)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "ORD-789", result.OrderID)
	assert.Equal(t, "complete", result.Status)
	assert.Equal(t, 100.0, result.FiatAmount)
	assert.Equal(t, "USD", result.FiatCurrency)
	assert.Equal(t, 0.0025, result.CryptoAmount)
	assert.Equal(t, "BTC", result.CryptoCurrency)
	assert.Equal(t, "0xabc123...", result.TransactionHash)
	assert.Equal(t, 2.50, result.Fee)
}

func TestDefaultResponsePipeline_Normalize(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "p1", Name: "test"}

	mappings := []*models.ResponseMapping{
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "id"},
		{ProviderID: "p1", Action: models.ActionGetOrder, TargetField: "order_id", SourceJSONPath: "id"},
		{ProviderID: "p1", Action: models.ActionGetQuote, TargetField: "quote_id", SourceJSONPath: "id"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)

	tests := []struct {
		action       string
		expectedType string
	}{
		{models.ActionCreateOrder, "*models.UnifiedCreateOrderResponse"},
		{models.ActionGetOrder, "*models.UnifiedGetOrderResponse"},
		{models.ActionGetQuote, "*models.UnifiedGetQuoteResponse"},
		{models.ActionGetCountries, "*models.UnifiedGetCountriesResponse"},
		{models.ActionGetPaymentMethods, "*models.UnifiedGetPaymentMethodsResponse"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			rawResponse := map[string]interface{}{"id": "123"}
			ctx := NewResponseContext(provider, tt.action, rawResponse, 200)

			result, err := pipeline.Normalize(ctx)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestDefaultResponsePipeline_ErrorResponse(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "p1", Name: "test"}

	rawResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "INVALID_REQUEST",
			"message": "Missing required field",
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "data.id"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 400)
	ctx.Error = "Bad Request"

	result, err := pipeline.NormalizeCreateOrder(ctx)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Equal(t, 400, result.StatusCode)
	assert.Equal(t, "Bad Request", result.Error)
	assert.NotNil(t, result.RawResponse)
}

func TestDefaultResponsePipeline_AdditionalData(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "p1", Name: "test"}

	rawResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"id":           "123",
			"custom_field": "custom_value",
			"extra_info":   42,
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "data.id"},
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "custom_field", SourceJSONPath: "data.custom_field"},
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "extra_info", SourceJSONPath: "data.extra_info"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 200)

	result, err := pipeline.NormalizeCreateOrder(ctx)
	require.NoError(t, err)

	assert.Equal(t, "123", result.OrderID)
	assert.Equal(t, "custom_value", result.AdditionalData["custom_field"])
	assert.Equal(t, float64(42), result.AdditionalData["extra_info"])
}

func TestQuickNormalize(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "p1", Name: "test"}

	rawBytes := json.RawMessage(`{"data":{"id":"ORD-123","status":"pending"}}`)

	mappings := []*models.ResponseMapping{
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "data.id"},
		{ProviderID: "p1", Action: models.ActionCreateOrder, TargetField: "status", SourceJSONPath: "data.status"},
	}

	result, err := QuickNormalize(provider, models.ActionCreateOrder, rawBytes, 200, mappings, logger)
	require.NoError(t, err)

	createOrderResp, ok := result.(*models.UnifiedCreateOrderResponse)
	require.True(t, ok)
	assert.Equal(t, "ORD-123", createOrderResp.OrderID)
	assert.Equal(t, "pending", createOrderResp.Status)
}

// ==================================================
// Test Different Providers
// ==================================================

func TestPipeline_BanxaCreateOrder(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "banxa", Name: "banxa"}

	// Simulate real Banxa response structure
	rawResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"order": map[string]interface{}{
				"id":                 "e7f3e6c3-4d5a-4b6c-8a9b-1234567890ab",
				"account_id":         "acc-123",
				"account_reference":  "ref-456",
				"order_type":         "BUY",
				"payment_type":       "card",
				"fiat_code":          "USD",
				"fiat_amount":        100,
				"coin_code":          "BTC",
				"coin_amount":        0.00245,
				"wallet_address":     "bc1qtest123456789",
				"wallet_address_tag": nil,
				"fee":                2.5,
				"fee_tax":            0,
				"payment_fee":        1.5,
				"payment_fee_tax":    0,
				"commission":         0,
				"tx_hash":            nil,
				"tx_confirms":        0,
				"created_at":         "2024-01-15T10:30:00Z",
				"checkout_url":       "https://checkout.banxa.com/order/abc123",
			},
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "data.order.id"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "checkout_url", SourceJSONPath: "data.order.checkout_url"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "status", SourceJSONPath: "data.order.order_type", Transform: "lowercase"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "fiat_amount", SourceJSONPath: "data.order.fiat_amount"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "fiat_currency", SourceJSONPath: "data.order.fiat_code"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "crypto_amount", SourceJSONPath: "data.order.coin_amount"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "crypto_currency", SourceJSONPath: "data.order.coin_code"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "wallet_address", SourceJSONPath: "data.order.wallet_address"},
		{ProviderID: "banxa", Action: models.ActionCreateOrder, TargetField: "payment_method", SourceJSONPath: "data.order.payment_type"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 200)

	result, err := pipeline.NormalizeCreateOrder(ctx)
	require.NoError(t, err)

	assert.Equal(t, "e7f3e6c3-4d5a-4b6c-8a9b-1234567890ab", result.OrderID)
	assert.Equal(t, "https://checkout.banxa.com/order/abc123", result.CheckoutURL)
	assert.Equal(t, "buy", result.Status)
	assert.Equal(t, float64(100), result.FiatAmount)
	assert.Equal(t, "USD", result.FiatCurrency)
	assert.Equal(t, 0.00245, result.CryptoAmount)
	assert.Equal(t, "BTC", result.CryptoCurrency)
	assert.Equal(t, "card", result.PaymentMethod)
}

func TestPipeline_TransakCreateOrder(t *testing.T) {
	logger := newTestLogger()
	provider := models.Provider{ID: "transak", Name: "transak"}

	// Simulate real Transak response structure
	rawResponse := map[string]interface{}{
		"response": map[string]interface{}{
			"id":              "tr-789xyz",
			"status":          "AWAITING_PAYMENT_FROM_USER",
			"fiatAmount":      100,
			"fiatCurrency":    "USD",
			"cryptoAmount":    0.00245,
			"cryptocurrency":  "BTC",
			"walletAddress":   "bc1qtest123456789",
			"network":         "mainnet",
			"paymentOptionId": "credit_debit_card",
			"redirectURL":     "https://global.transak.com/order/tr-789xyz",
			"createdAt":       "2024-01-15T10:30:00.000Z",
			"conversionPrice": 40816.33,
			"totalFeeInFiat":  3.99,
		},
	}

	mappings := []*models.ResponseMapping{
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "order_id", SourceJSONPath: "response.id"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "checkout_url", SourceJSONPath: "response.redirectURL"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "status", SourceJSONPath: "response.status", Transform: "lowercase"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "fiat_amount", SourceJSONPath: "response.fiatAmount"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "fiat_currency", SourceJSONPath: "response.fiatCurrency"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "crypto_amount", SourceJSONPath: "response.cryptoAmount"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "crypto_currency", SourceJSONPath: "response.cryptocurrency"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "wallet_address", SourceJSONPath: "response.walletAddress"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "network", SourceJSONPath: "response.network"},
		{ProviderID: "transak", Action: models.ActionCreateOrder, TargetField: "payment_method", SourceJSONPath: "response.paymentOptionId"},
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	ctx := NewResponseContext(provider, models.ActionCreateOrder, rawResponse, 200)

	result, err := pipeline.NormalizeCreateOrder(ctx)
	require.NoError(t, err)

	assert.Equal(t, "tr-789xyz", result.OrderID)
	assert.Equal(t, "https://global.transak.com/order/tr-789xyz", result.CheckoutURL)
	assert.Equal(t, "awaiting_payment_from_user", result.Status)
	assert.Equal(t, "mainnet", result.Network)
	assert.Equal(t, "credit_debit_card", result.PaymentMethod)
}
