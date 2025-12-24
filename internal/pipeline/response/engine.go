package response

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"go.uber.org/zap"
)

// ResponsePipeline defines the interface for the response normalization pipeline
type ResponsePipeline interface {
	// NormalizeCreateOrder normalizes a create order response
	NormalizeCreateOrder(ctx *ResponseContext) (*models.UnifiedCreateOrderResponse, error)

	// NormalizeGetOrder normalizes a get order response
	NormalizeGetOrder(ctx *ResponseContext) (*models.UnifiedGetOrderResponse, error)

	// NormalizeGetQuote normalizes a get quote response
	NormalizeGetQuote(ctx *ResponseContext) (*models.UnifiedGetQuoteResponse, error)

	// NormalizeGetCountries normalizes a get countries response
	NormalizeGetCountries(ctx *ResponseContext) (*models.UnifiedGetCountriesResponse, error)

	// NormalizeGetPaymentMethods normalizes a get payment methods response
	NormalizeGetPaymentMethods(ctx *ResponseContext) (*models.UnifiedGetPaymentMethodsResponse, error)

	// NormalizeGetCurrencies normalizes a get currencies response
	NormalizeGetCurrencies(ctx *ResponseContext) (*models.UnifiedGetCurrenciesResponse, error)

	// NormalizeGetLimits normalizes a get limits response
	NormalizeGetLimits(ctx *ResponseContext) (*models.UnifiedGetLimitsResponse, error)

	// Normalize normalizes any response based on action type
	Normalize(ctx *ResponseContext) (interface{}, error)
}

// DefaultResponsePipeline is the default implementation of ResponsePipeline
type DefaultResponsePipeline struct {
	loader MappingLoader
	steps  []PipelineStep
	logger *zap.Logger
}

// NewDefaultResponsePipeline creates a new DefaultResponsePipeline
func NewDefaultResponsePipeline(loader MappingLoader, logger *zap.Logger) *DefaultResponsePipeline {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	pipeline := &DefaultResponsePipeline{
		loader: loader,
		logger: logger,
	}

	// Initialize pipeline steps
	pipeline.steps = []PipelineStep{
		NewLoadMappingsStep(loader, logger),
		NewJSONPathExtractStep(logger),
		NewTransformStep(logger),
	}

	return pipeline
}

// NewDefaultResponsePipelineWithMappings creates a pipeline with pre-loaded mappings (for testing)
func NewDefaultResponsePipelineWithMappings(mappings []*models.ResponseMapping, logger *zap.Logger) *DefaultResponsePipeline {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	pipeline := &DefaultResponsePipeline{
		loader: &staticMappingLoader{mappings: mappings},
		logger: logger,
	}

	pipeline.steps = []PipelineStep{
		NewLoadMappingsStep(pipeline.loader, logger),
		NewJSONPathExtractStep(logger),
		NewTransformStep(logger),
	}

	return pipeline
}

// staticMappingLoader is a simple loader that returns pre-loaded mappings
type staticMappingLoader struct {
	mappings []*models.ResponseMapping
}

func (l *staticMappingLoader) LoadMappings(providerID, action string) ([]*models.ResponseMapping, error) {
	var result []*models.ResponseMapping
	for _, m := range l.mappings {
		if m.ProviderID == providerID && m.Action == action {
			result = append(result, m)
		}
	}
	return result, nil
}

// ExecuteSteps runs all pipeline steps on the context
func (p *DefaultResponsePipeline) ExecuteSteps(ctx *ResponseContext) error {
	p.logger.Debug("Executing response pipeline",
		zap.String("provider", ctx.Provider.Name),
		zap.String("action", ctx.Action),
	)

	for _, step := range p.steps {
		p.logger.Debug("Executing step", zap.String("step", step.Name()))
		if err := step.Execute(ctx); err != nil {
			p.logger.Error("Pipeline step failed",
				zap.String("step", step.Name()),
				zap.Error(err),
			)
			return fmt.Errorf("step %s failed: %w", step.Name(), err)
		}
	}

	p.logger.Debug("Response pipeline complete",
		zap.Int("extracted_fields", len(ctx.ExtractedValues)),
		zap.Int("transformed_fields", len(ctx.TransformedValues)),
	)

	return nil
}

// Normalize normalizes any response based on action type
func (p *DefaultResponsePipeline) Normalize(ctx *ResponseContext) (interface{}, error) {
	switch ctx.Action {
	case models.ActionCreateOrder, models.ActionCreateWidgetURL:
		return p.NormalizeCreateOrder(ctx)
	case models.ActionGetOrder, models.ActionGetTransactionStatus:
		return p.NormalizeGetOrder(ctx)
	case models.ActionGetQuote:
		return p.NormalizeGetQuote(ctx)
	case models.ActionGetCountries:
		return p.NormalizeGetCountries(ctx)
	case models.ActionGetPaymentMethods:
		return p.NormalizeGetPaymentMethods(ctx)
	case models.ActionGetCurrencies:
		return p.NormalizeGetCurrencies(ctx)
	case models.ActionGetLimits:
		return p.NormalizeGetLimits(ctx)
	default:
		return nil, fmt.Errorf("unsupported action: %s", ctx.Action)
	}
}

// buildBaseResponse builds the base unified response fields
func (p *DefaultResponsePipeline) buildBaseResponse(ctx *ResponseContext) models.BaseUnifiedResponse {
	return models.BaseUnifiedResponse{
		Success:       ctx.IsSuccess(),
		StatusCode:    ctx.StatusCode,
		ProviderName:  ctx.Provider.Name,
		Action:        ctx.Action,
		RequestID:     ctx.RequestID,
		ExecutionTime: ctx.ExecutionTime,
		RawResponse:   ctx.RawResponseBytes,
		Error:         ctx.Error,
		Timestamp:     time.Now(),
	}
}

// NormalizeCreateOrder normalizes a create order response
func (p *DefaultResponsePipeline) NormalizeCreateOrder(ctx *ResponseContext) (*models.UnifiedCreateOrderResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedCreateOrderResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		AdditionalData:      make(map[string]interface{}),
	}

	// Map extracted values to response fields
	for field, value := range ctx.TransformedValues {
		switch field {
		case "order_id", "OrderID":
			response.OrderID = toString(value)
		case "external_order_id", "ExternalOrderID":
			response.ExternalOrderID = toString(value)
		case "status", "Status":
			response.Status = toString(value)
		case "checkout_url", "CheckoutURL":
			response.CheckoutURL = toString(value)
		case "widget_url", "WidgetURL":
			response.WidgetURL = toString(value)
		case "fiat_amount", "FiatAmount":
			response.FiatAmount = toFloat64(value)
		case "fiat_currency", "FiatCurrency":
			response.FiatCurrency = toString(value)
		case "crypto_amount", "CryptoAmount":
			response.CryptoAmount = toFloat64(value)
		case "crypto_currency", "CryptoCurrency":
			response.CryptoCurrency = toString(value)
		case "wallet_address", "WalletAddress":
			response.WalletAddress = toString(value)
		case "network", "Network":
			response.Network = toString(value)
		case "payment_method", "PaymentMethod":
			response.PaymentMethod = toString(value)
		case "expires_at", "ExpiresAt":
			response.ExpiresAt = toTimePtr(value)
		case "created_at", "CreatedAt":
			response.CreatedAt = toTimePtr(value)
		default:
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// NormalizeGetOrder normalizes a get order response
func (p *DefaultResponsePipeline) NormalizeGetOrder(ctx *ResponseContext) (*models.UnifiedGetOrderResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetOrderResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		AdditionalData:      make(map[string]interface{}),
	}

	// Map extracted values to response fields
	for field, value := range ctx.TransformedValues {
		switch field {
		case "order_id", "OrderID":
			response.OrderID = toString(value)
		case "external_order_id", "ExternalOrderID":
			response.ExternalOrderID = toString(value)
		case "status", "Status":
			response.Status = toString(value)
		case "status_message", "StatusMessage":
			response.StatusMessage = toString(value)
		case "fiat_amount", "FiatAmount":
			response.FiatAmount = toFloat64(value)
		case "fiat_currency", "FiatCurrency":
			response.FiatCurrency = toString(value)
		case "crypto_amount", "CryptoAmount":
			response.CryptoAmount = toFloat64(value)
		case "crypto_currency", "CryptoCurrency":
			response.CryptoCurrency = toString(value)
		case "wallet_address", "WalletAddress":
			response.WalletAddress = toString(value)
		case "network", "Network":
			response.Network = toString(value)
		case "payment_method", "PaymentMethod":
			response.PaymentMethod = toString(value)
		case "transaction_hash", "TransactionHash":
			response.TransactionHash = toString(value)
		case "fee", "Fee":
			response.Fee = toFloat64(value)
		case "fee_currency", "FeeCurrency":
			response.FeeCurrency = toString(value)
		case "exchange_rate", "ExchangeRate":
			response.ExchangeRate = toFloat64(value)
		case "created_at", "CreatedAt":
			response.CreatedAt = toTimePtr(value)
		case "updated_at", "UpdatedAt":
			response.UpdatedAt = toTimePtr(value)
		case "completed_at", "CompletedAt":
			response.CompletedAt = toTimePtr(value)
		default:
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// NormalizeGetQuote normalizes a get quote response
func (p *DefaultResponsePipeline) NormalizeGetQuote(ctx *ResponseContext) (*models.UnifiedGetQuoteResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetQuoteResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		AdditionalData:      make(map[string]interface{}),
	}

	// Map extracted values to response fields
	for field, value := range ctx.TransformedValues {
		switch field {
		case "quote_id", "QuoteID":
			response.QuoteID = toString(value)
		case "fiat_amount", "FiatAmount":
			response.FiatAmount = toFloat64(value)
		case "fiat_currency", "FiatCurrency":
			response.FiatCurrency = toString(value)
		case "crypto_amount", "CryptoAmount":
			response.CryptoAmount = toFloat64(value)
		case "crypto_currency", "CryptoCurrency":
			response.CryptoCurrency = toString(value)
		case "exchange_rate", "ExchangeRate":
			response.ExchangeRate = toFloat64(value)
		case "fee", "Fee":
			response.Fee = toFloat64(value)
		case "fee_currency", "FeeCurrency":
			response.FeeCurrency = toString(value)
		case "fee_percentage", "FeePercentage":
			response.FeePercentage = toFloat64(value)
		case "network_fee", "NetworkFee":
			response.NetworkFee = toFloat64(value)
		case "total_amount", "TotalAmount":
			response.TotalAmount = toFloat64(value)
		case "payment_method", "PaymentMethod":
			response.PaymentMethod = toString(value)
		case "network", "Network":
			response.Network = toString(value)
		case "expires_at", "ExpiresAt":
			response.ExpiresAt = toTimePtr(value)
		case "valid_for_seconds", "ValidForSeconds":
			response.ValidForSeconds = toInt(value)
		case "min_amount", "MinAmount":
			response.MinAmount = toFloat64(value)
		case "max_amount", "MaxAmount":
			response.MaxAmount = toFloat64(value)
		default:
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// NormalizeGetCountries normalizes a get countries response
func (p *DefaultResponsePipeline) NormalizeGetCountries(ctx *ResponseContext) (*models.UnifiedGetCountriesResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetCountriesResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		Countries:           make([]models.Country, 0),
		AdditionalData:      make(map[string]interface{}),
	}

	// Check for countries array in transformed values
	if countriesRaw, ok := ctx.TransformedValues["countries"]; ok {
		response.Countries = parseCountries(countriesRaw)
	}

	response.TotalCount = len(response.Countries)

	// Any non-countries fields go to additional data
	for field, value := range ctx.TransformedValues {
		if field != "countries" && field != "Countries" {
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// NormalizeGetPaymentMethods normalizes a get payment methods response
func (p *DefaultResponsePipeline) NormalizeGetPaymentMethods(ctx *ResponseContext) (*models.UnifiedGetPaymentMethodsResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetPaymentMethodsResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		PaymentMethods:      make([]models.PaymentMethod, 0),
		AdditionalData:      make(map[string]interface{}),
	}

	// Check for payment_methods array in transformed values
	if methodsRaw, ok := ctx.TransformedValues["payment_methods"]; ok {
		response.PaymentMethods = parsePaymentMethods(methodsRaw)
	}

	response.TotalCount = len(response.PaymentMethods)

	// Any non-payment_methods fields go to additional data
	for field, value := range ctx.TransformedValues {
		if field != "payment_methods" && field != "PaymentMethods" {
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// NormalizeGetCurrencies normalizes a get currencies response
func (p *DefaultResponsePipeline) NormalizeGetCurrencies(ctx *ResponseContext) (*models.UnifiedGetCurrenciesResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetCurrenciesResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		FiatCurrencies:      make([]models.Currency, 0),
		CryptoCurrencies:    make([]models.Currency, 0),
		AdditionalData:      make(map[string]interface{}),
	}

	// Check for currencies in transformed values
	if fiatRaw, ok := ctx.TransformedValues["fiat_currencies"]; ok {
		response.FiatCurrencies = parseCurrencies(fiatRaw)
	}
	if cryptoRaw, ok := ctx.TransformedValues["crypto_currencies"]; ok {
		response.CryptoCurrencies = parseCurrencies(cryptoRaw)
	}

	response.TotalFiatCount = len(response.FiatCurrencies)
	response.TotalCryptoCount = len(response.CryptoCurrencies)

	return response, nil
}

// NormalizeGetLimits normalizes a get limits response
func (p *DefaultResponsePipeline) NormalizeGetLimits(ctx *ResponseContext) (*models.UnifiedGetLimitsResponse, error) {
	if err := p.ExecuteSteps(ctx); err != nil {
		return nil, err
	}

	response := &models.UnifiedGetLimitsResponse{
		BaseUnifiedResponse: p.buildBaseResponse(ctx),
		AdditionalData:      make(map[string]interface{}),
	}

	// Map extracted values to response fields
	for field, value := range ctx.TransformedValues {
		switch field {
		case "min_fiat_amount", "MinFiatAmount":
			response.MinFiatAmount = toFloat64(value)
		case "max_fiat_amount", "MaxFiatAmount":
			response.MaxFiatAmount = toFloat64(value)
		case "min_crypto_amount", "MinCryptoAmount":
			response.MinCryptoAmount = toFloat64(value)
		case "max_crypto_amount", "MaxCryptoAmount":
			response.MaxCryptoAmount = toFloat64(value)
		case "daily_limit", "DailyLimit":
			response.DailyLimit = toFloat64(value)
		case "monthly_limit", "MonthlyLimit":
			response.MonthlyLimit = toFloat64(value)
		case "remaining_daily_limit", "RemainingDailyLimit":
			response.RemainingDailyLimit = toFloat64(value)
		case "remaining_monthly_limit", "RemainingMonthlyLimit":
			response.RemainingMonthlyLimit = toFloat64(value)
		case "limit_currency", "LimitCurrency":
			response.LimitCurrency = toString(value)
		default:
			response.AdditionalData[field] = value
		}
	}

	return response, nil
}

// Helper functions for type conversion

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
}

func toTimePtr(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		return &val
	case *time.Time:
		return val
	case string:
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return &t
		}
	}
	return nil
}

func parseCountries(v interface{}) []models.Country {
	countries := make([]models.Country, 0)
	if arr, ok := v.([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				country := models.Country{
					Code:        toString(m["code"]),
					Name:        toString(m["name"]),
					IsSupported: true,
				}
				countries = append(countries, country)
			}
		}
	}
	return countries
}

func parsePaymentMethods(v interface{}) []models.PaymentMethod {
	methods := make([]models.PaymentMethod, 0)
	if arr, ok := v.([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				method := models.PaymentMethod{
					ID:        toString(m["id"]),
					Name:      toString(m["name"]),
					Type:      toString(m["type"]),
					IsEnabled: true,
				}
				methods = append(methods, method)
			}
		}
	}
	return methods
}

func parseCurrencies(v interface{}) []models.Currency {
	currencies := make([]models.Currency, 0)
	if arr, ok := v.([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				currency := models.Currency{
					Code:      toString(m["code"]),
					Name:      toString(m["name"]),
					Symbol:    toString(m["symbol"]),
					Type:      toString(m["type"]),
					IsEnabled: true,
				}
				currencies = append(currencies, currency)
			}
		}
	}
	return currencies
}

// QuickNormalize is a convenience function for quick normalization without setting up a full pipeline
func QuickNormalize(
	provider models.Provider,
	action string,
	rawResponse json.RawMessage,
	statusCode int,
	mappings []*models.ResponseMapping,
	logger *zap.Logger,
) (interface{}, error) {
	ctx, err := NewResponseContextFromBytes(provider, action, rawResponse, statusCode)
	if err != nil {
		return nil, err
	}

	pipeline := NewDefaultResponsePipelineWithMappings(mappings, logger)
	return pipeline.Normalize(ctx)
}
