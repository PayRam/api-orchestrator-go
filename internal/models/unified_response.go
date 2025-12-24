package models

import (
	"encoding/json"
	"time"
)

// BaseUnifiedResponse contains common fields for all unified responses
type BaseUnifiedResponse struct {
	// Success indicates if the API call was successful
	Success bool `json:"success"`

	// StatusCode is the HTTP status code from the provider
	StatusCode int `json:"status_code"`

	// ProviderName is the name of the provider that handled the request
	ProviderName string `json:"provider_name"`

	// Action is the action that was performed
	Action string `json:"action"`

	// RequestID is the unique identifier for this request
	RequestID string `json:"request_id"`

	// ExecutionTime is how long the request took
	ExecutionTime time.Duration `json:"execution_time"`

	// RawResponse is the original unmodified response from the provider
	RawResponse json.RawMessage `json:"raw_response"`

	// Error contains error message if the request failed
	Error string `json:"error,omitempty"`

	// Timestamp is when the response was received
	Timestamp time.Time `json:"timestamp"`
}

// UnifiedCreateOrderResponse represents a normalized create order response
type UnifiedCreateOrderResponse struct {
	BaseUnifiedResponse

	// OrderID is the unique identifier for the created order
	OrderID string `json:"order_id,omitempty"`

	// ExternalOrderID is the provider's order reference
	ExternalOrderID string `json:"external_order_id,omitempty"`

	// Status is the order status (e.g., "pending", "processing", "completed")
	Status string `json:"status,omitempty"`

	// CheckoutURL is the URL to redirect user to complete payment
	CheckoutURL string `json:"checkout_url,omitempty"`

	// WidgetURL is the URL for embedded widget (if applicable)
	WidgetURL string `json:"widget_url,omitempty"`

	// FiatAmount is the fiat currency amount
	FiatAmount float64 `json:"fiat_amount,omitempty"`

	// FiatCurrency is the fiat currency code (e.g., "USD", "EUR")
	FiatCurrency string `json:"fiat_currency,omitempty"`

	// CryptoAmount is the cryptocurrency amount
	CryptoAmount float64 `json:"crypto_amount,omitempty"`

	// CryptoCurrency is the cryptocurrency code (e.g., "BTC", "ETH")
	CryptoCurrency string `json:"crypto_currency,omitempty"`

	// WalletAddress is the destination wallet address
	WalletAddress string `json:"wallet_address,omitempty"`

	// Network is the blockchain network (e.g., "ethereum", "polygon")
	Network string `json:"network,omitempty"`

	// PaymentMethod is the selected payment method
	PaymentMethod string `json:"payment_method,omitempty"`

	// ExpiresAt is when the order/checkout expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// CreatedAt is when the order was created
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// AdditionalData contains any extra fields not in the standard schema
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// UnifiedGetOrderResponse represents a normalized get order response
type UnifiedGetOrderResponse struct {
	BaseUnifiedResponse

	// OrderID is the unique identifier for the order
	OrderID string `json:"order_id,omitempty"`

	// ExternalOrderID is the provider's order reference
	ExternalOrderID string `json:"external_order_id,omitempty"`

	// Status is the order status
	Status string `json:"status,omitempty"`

	// StatusMessage provides additional status details
	StatusMessage string `json:"status_message,omitempty"`

	// FiatAmount is the fiat currency amount
	FiatAmount float64 `json:"fiat_amount,omitempty"`

	// FiatCurrency is the fiat currency code
	FiatCurrency string `json:"fiat_currency,omitempty"`

	// CryptoAmount is the cryptocurrency amount
	CryptoAmount float64 `json:"crypto_amount,omitempty"`

	// CryptoCurrency is the cryptocurrency code
	CryptoCurrency string `json:"crypto_currency,omitempty"`

	// WalletAddress is the destination wallet address
	WalletAddress string `json:"wallet_address,omitempty"`

	// Network is the blockchain network
	Network string `json:"network,omitempty"`

	// PaymentMethod is the payment method used
	PaymentMethod string `json:"payment_method,omitempty"`

	// TransactionHash is the blockchain transaction hash (if completed)
	TransactionHash string `json:"transaction_hash,omitempty"`

	// Fee is the transaction fee
	Fee float64 `json:"fee,omitempty"`

	// FeeCurrency is the fee currency
	FeeCurrency string `json:"fee_currency,omitempty"`

	// ExchangeRate is the exchange rate used
	ExchangeRate float64 `json:"exchange_rate,omitempty"`

	// CreatedAt is when the order was created
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// UpdatedAt is when the order was last updated
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// CompletedAt is when the order was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// UnifiedGetQuoteResponse represents a normalized quote response
type UnifiedGetQuoteResponse struct {
	BaseUnifiedResponse

	// QuoteID is the unique identifier for the quote
	QuoteID string `json:"quote_id,omitempty"`

	// FiatAmount is the fiat currency amount
	FiatAmount float64 `json:"fiat_amount,omitempty"`

	// FiatCurrency is the fiat currency code
	FiatCurrency string `json:"fiat_currency,omitempty"`

	// CryptoAmount is the cryptocurrency amount
	CryptoAmount float64 `json:"crypto_amount,omitempty"`

	// CryptoCurrency is the cryptocurrency code
	CryptoCurrency string `json:"crypto_currency,omitempty"`

	// ExchangeRate is the exchange rate
	ExchangeRate float64 `json:"exchange_rate,omitempty"`

	// Fee is the transaction fee
	Fee float64 `json:"fee,omitempty"`

	// FeeCurrency is the fee currency
	FeeCurrency string `json:"fee_currency,omitempty"`

	// FeePercentage is the fee as a percentage
	FeePercentage float64 `json:"fee_percentage,omitempty"`

	// NetworkFee is the blockchain network fee
	NetworkFee float64 `json:"network_fee,omitempty"`

	// TotalAmount is the total amount including fees
	TotalAmount float64 `json:"total_amount,omitempty"`

	// PaymentMethod is the payment method for this quote
	PaymentMethod string `json:"payment_method,omitempty"`

	// Network is the blockchain network
	Network string `json:"network,omitempty"`

	// ExpiresAt is when the quote expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ValidForSeconds is how long the quote is valid
	ValidForSeconds int `json:"valid_for_seconds,omitempty"`

	// MinAmount is the minimum allowed amount
	MinAmount float64 `json:"min_amount,omitempty"`

	// MaxAmount is the maximum allowed amount
	MaxAmount float64 `json:"max_amount,omitempty"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// Country represents a supported country
type Country struct {
	// Code is the ISO country code (e.g., "US", "GB")
	Code string `json:"code"`

	// Name is the country name
	Name string `json:"name"`

	// IsSupported indicates if the country is currently supported
	IsSupported bool `json:"is_supported"`

	// SupportedPaymentMethods lists payment methods available in this country
	SupportedPaymentMethods []string `json:"supported_payment_methods,omitempty"`
}

// UnifiedGetCountriesResponse represents a normalized countries list response
type UnifiedGetCountriesResponse struct {
	BaseUnifiedResponse

	// Countries is the list of supported countries
	Countries []Country `json:"countries,omitempty"`

	// TotalCount is the total number of countries
	TotalCount int `json:"total_count"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	// ID is the unique identifier for the payment method
	ID string `json:"id"`

	// Name is the display name
	Name string `json:"name"`

	// Type is the payment method type (e.g., "bank_transfer", "card", "wallet")
	Type string `json:"type"`

	// Logo is the URL to the payment method logo
	Logo string `json:"logo,omitempty"`

	// MinAmount is the minimum transaction amount
	MinAmount float64 `json:"min_amount,omitempty"`

	// MaxAmount is the maximum transaction amount
	MaxAmount float64 `json:"max_amount,omitempty"`

	// Fee is the fee for this payment method
	Fee float64 `json:"fee,omitempty"`

	// FeeType is the fee type (e.g., "percentage", "fixed")
	FeeType string `json:"fee_type,omitempty"`

	// ProcessingTime is the estimated processing time
	ProcessingTime string `json:"processing_time,omitempty"`

	// SupportedCurrencies lists currencies supported by this method
	SupportedCurrencies []string `json:"supported_currencies,omitempty"`

	// SupportedCountries lists countries where this method is available
	SupportedCountries []string `json:"supported_countries,omitempty"`

	// IsEnabled indicates if the method is currently enabled
	IsEnabled bool `json:"is_enabled"`
}

// UnifiedGetPaymentMethodsResponse represents a normalized payment methods response
type UnifiedGetPaymentMethodsResponse struct {
	BaseUnifiedResponse

	// PaymentMethods is the list of available payment methods
	PaymentMethods []PaymentMethod `json:"payment_methods,omitempty"`

	// TotalCount is the total number of payment methods
	TotalCount int `json:"total_count"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// Currency represents a supported currency (fiat or crypto)
type Currency struct {
	// Code is the currency code (e.g., "USD", "BTC")
	Code string `json:"code"`

	// Name is the currency name
	Name string `json:"name"`

	// Symbol is the currency symbol (e.g., "$", "₿")
	Symbol string `json:"symbol,omitempty"`

	// Type is "fiat" or "crypto"
	Type string `json:"type"`

	// Network is the blockchain network (for crypto only)
	Network string `json:"network,omitempty"`

	// Decimals is the number of decimal places
	Decimals int `json:"decimals"`

	// MinAmount is the minimum transaction amount
	MinAmount float64 `json:"min_amount,omitempty"`

	// MaxAmount is the maximum transaction amount
	MaxAmount float64 `json:"max_amount,omitempty"`

	// IsEnabled indicates if the currency is currently enabled
	IsEnabled bool `json:"is_enabled"`

	// Logo is the URL to the currency logo
	Logo string `json:"logo,omitempty"`
}

// UnifiedGetCurrenciesResponse represents a normalized currencies response
type UnifiedGetCurrenciesResponse struct {
	BaseUnifiedResponse

	// FiatCurrencies is the list of fiat currencies
	FiatCurrencies []Currency `json:"fiat_currencies,omitempty"`

	// CryptoCurrencies is the list of cryptocurrencies
	CryptoCurrencies []Currency `json:"crypto_currencies,omitempty"`

	// TotalFiatCount is the total number of fiat currencies
	TotalFiatCount int `json:"total_fiat_count"`

	// TotalCryptoCount is the total number of cryptocurrencies
	TotalCryptoCount int `json:"total_crypto_count"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// UnifiedGetLimitsResponse represents a normalized limits response
type UnifiedGetLimitsResponse struct {
	BaseUnifiedResponse

	// MinFiatAmount is the minimum fiat amount
	MinFiatAmount float64 `json:"min_fiat_amount,omitempty"`

	// MaxFiatAmount is the maximum fiat amount
	MaxFiatAmount float64 `json:"max_fiat_amount,omitempty"`

	// MinCryptoAmount is the minimum crypto amount
	MinCryptoAmount float64 `json:"min_crypto_amount,omitempty"`

	// MaxCryptoAmount is the maximum crypto amount
	MaxCryptoAmount float64 `json:"max_crypto_amount,omitempty"`

	// DailyLimit is the daily transaction limit
	DailyLimit float64 `json:"daily_limit,omitempty"`

	// MonthlyLimit is the monthly transaction limit
	MonthlyLimit float64 `json:"monthly_limit,omitempty"`

	// RemainingDailyLimit is the remaining daily limit
	RemainingDailyLimit float64 `json:"remaining_daily_limit,omitempty"`

	// RemainingMonthlyLimit is the remaining monthly limit
	RemainingMonthlyLimit float64 `json:"remaining_monthly_limit,omitempty"`

	// LimitCurrency is the currency for the limits
	LimitCurrency string `json:"limit_currency,omitempty"`

	// AdditionalData contains any extra fields
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}
