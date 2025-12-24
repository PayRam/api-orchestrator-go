package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"sort"
	"strings"
	"time"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

// Strategy types constants
const (
	StrategyTypeHMAC         = "HMAC"
	StrategyTypeSignature    = "SIGNATURE"
	StrategyTypeTokenRefresh = "TOKEN_REFRESH"
	StrategyTypePayload      = "PAYLOAD"
	StrategyTypeJWT          = "JWT"
	StrategyTypeBasicAuth    = "BASIC_AUTH"
	StrategyTypeAPIKey       = "API_KEY"
)

type strategyServiceImpl struct {
	repo   repositories.StrategyRepo
	logger *zap.Logger
}

// NewStrategyService creates a new strategy service implementation
func NewStrategyService(repo repositories.StrategyRepo, logger *zap.Logger) StrategyService {
	return &strategyServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *strategyServiceImpl) CreateStrategy(strategy *models.Strategy) error {
	s.logger.Info("Creating strategy", zap.String("name", strategy.Name))

	// Business logic validations
	if strategy.Name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}
	if strategy.StrategyType == "" {
		return fmt.Errorf("strategy type cannot be empty")
	}

	// Check if strategy already exists
	existing, _ := s.repo.FindByName(strategy.Name)
	if existing != nil {
		return fmt.Errorf("strategy with name %s already exists", strategy.Name)
	}

	err := s.repo.Create(strategy)
	if err != nil {
		s.logger.Error("Failed to create strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy created successfully", zap.String("id", strategy.ID))
	return nil
}

func (s *strategyServiceImpl) GetStrategyByID(id string) (*models.Strategy, error) {
	s.logger.Debug("Fetching strategy by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *strategyServiceImpl) GetStrategyByName(name string) (*models.Strategy, error) {
	s.logger.Debug("Fetching strategy by name", zap.String("name", name))
	return s.repo.FindByName(name)
}

func (s *strategyServiceImpl) ListStrategies() ([]*models.Strategy, error) {
	s.logger.Debug("Listing all strategies")
	return s.repo.List()
}

func (s *strategyServiceImpl) GetStrategiesByType(strategyType string) ([]*models.Strategy, error) {
	s.logger.Debug("Fetching strategies by type", zap.String("type", strategyType))
	return s.repo.FindByType(strategyType)
}

func (s *strategyServiceImpl) UpdateStrategy(strategy *models.Strategy) error {
	s.logger.Info("Updating strategy", zap.String("id", strategy.ID))

	if strategy.Name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	err := s.repo.Update(strategy)
	if err != nil {
		s.logger.Error("Failed to update strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy updated successfully")
	return nil
}

func (s *strategyServiceImpl) DeleteStrategy(id string) error {
	s.logger.Info("Deleting strategy", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete strategy", zap.Error(err))
		return err
	}

	s.logger.Info("Strategy deleted successfully")
	return nil
}

// ExecuteStrategy executes a strategy by name and returns the computed values
func (s *strategyServiceImpl) ExecuteStrategy(strategyName string, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	s.logger.Debug("Executing strategy", zap.String("name", strategyName))

	strategy, err := s.repo.FindByName(strategyName)
	if err != nil {
		return nil, fmt.Errorf("strategy not found: %w", err)
	}

	return s.executeStrategyInternal(strategy, ctx)
}

// ExecuteStrategyByID executes a strategy by ID and returns the computed values
func (s *strategyServiceImpl) ExecuteStrategyByID(strategyID string, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	s.logger.Debug("Executing strategy by ID", zap.String("id", strategyID))

	strategy, err := s.repo.FindByID(strategyID)
	if err != nil {
		return nil, fmt.Errorf("strategy not found: %w", err)
	}

	return s.executeStrategyInternal(strategy, ctx)
}

// executeStrategyInternal handles the actual strategy execution
func (s *strategyServiceImpl) executeStrategyInternal(strategy *models.Strategy, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	s.logger.Debug("Executing strategy",
		zap.String("name", strategy.Name),
		zap.String("type", strategy.StrategyType))

	// Parse strategy config
	config, err := s.parseConfig(strategy.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse strategy config: %w", err)
	}

	// Execute based on strategy type
	switch strategy.StrategyType {
	case StrategyTypeHMAC:
		return s.executeHMAC(strategy, config, ctx)
	case StrategyTypeSignature:
		return s.executeSignature(strategy, config, ctx)
	case StrategyTypeBasicAuth:
		return s.executeBasicAuth(strategy, config, ctx)
	case StrategyTypeAPIKey:
		return s.executeAPIKey(strategy, config, ctx)
	case StrategyTypePayload:
		return s.executePayload(strategy, config, ctx)
	default:
		return nil, fmt.Errorf("unsupported strategy type: %s", strategy.StrategyType)
	}
}

// parseConfig parses the JSON config into a map
func (s *strategyServiceImpl) parseConfig(configJSON []byte) (map[string]interface{}, error) {
	if len(configJSON) == 0 {
		return make(map[string]interface{}), nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// executeHMAC executes an HMAC signing strategy
func (s *strategyServiceImpl) executeHMAC(strategy *models.Strategy, config map[string]interface{}, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	// Get algorithm from config (default: SHA256)
	algorithm := s.getConfigString(config, "algorithm", "SHA256")

	// Get secret key reference and resolve it
	secretKeyRef := s.getConfigString(config, "secret_key_ref", "api_secret")
	secretKey, ok := ctx.Credentials[secretKeyRef]
	if !ok {
		return nil, fmt.Errorf("secret key not found in credentials: %s", secretKeyRef)
	}

	// Get message template from config
	messageTemplate := s.getConfigString(config, "message_template", "")
	message := s.buildMessage(messageTemplate, config, ctx)

	// Get encoding from config (default: hex)
	encoding := s.getConfigString(config, "encoding", "hex")

	// Compute HMAC
	signature, err := s.computeHMAC([]byte(secretKey), []byte(message), algorithm, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to compute HMAC: %w", err)
	}

	// Get output key name from config
	outputKey := s.getConfigString(config, "output_key", "signature")

	result := &StrategyResult{
		Values:  map[string]interface{}{outputKey: signature},
		Headers: make(map[string]string),
	}

	// Add to headers if configured
	if headerName := s.getConfigString(config, "header_name", ""); headerName != "" {
		headerPrefix := s.getConfigString(config, "header_prefix", "")
		result.Headers[headerName] = headerPrefix + signature
	}

	s.logger.Debug("HMAC strategy executed successfully",
		zap.String("algorithm", algorithm),
		zap.String("output_key", outputKey))

	return result, nil
}

// executeSignature executes a signature generation strategy
func (s *strategyServiceImpl) executeSignature(strategy *models.Strategy, config map[string]interface{}, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	// Similar to HMAC but with different message construction
	algorithm := s.getConfigString(config, "algorithm", "SHA256")
	secretKeyRef := s.getConfigString(config, "secret_key_ref", "api_secret")

	secretKey, ok := ctx.Credentials[secretKeyRef]
	if !ok {
		return nil, fmt.Errorf("secret key not found: %s", secretKeyRef)
	}

	// Build signature message from parameters
	messageTemplate := s.getConfigString(config, "message_template", "")
	message := s.buildMessage(messageTemplate, config, ctx)

	encoding := s.getConfigString(config, "encoding", "base64")

	signature, err := s.computeHMAC([]byte(secretKey), []byte(message), algorithm, encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to compute signature: %w", err)
	}

	outputKey := s.getConfigString(config, "output_key", "signature")

	result := &StrategyResult{
		Values:  map[string]interface{}{outputKey: signature},
		Headers: make(map[string]string),
	}

	if headerName := s.getConfigString(config, "header_name", ""); headerName != "" {
		result.Headers[headerName] = signature
	}

	return result, nil
}

// executeBasicAuth generates Basic Auth header
func (s *strategyServiceImpl) executeBasicAuth(strategy *models.Strategy, config map[string]interface{}, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	usernameRef := s.getConfigString(config, "username_ref", "api_key")
	passwordRef := s.getConfigString(config, "password_ref", "api_secret")

	username, ok := ctx.Credentials[usernameRef]
	if !ok {
		return nil, fmt.Errorf("username not found: %s", usernameRef)
	}

	password, ok := ctx.Credentials[passwordRef]
	if !ok {
		return nil, fmt.Errorf("password not found: %s", passwordRef)
	}

	// Encode as Base64
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	result := &StrategyResult{
		Values: map[string]interface{}{
			"basic_auth": auth,
		},
		Headers: map[string]string{
			"Authorization": "Basic " + auth,
		},
	}

	return result, nil
}

// executeAPIKey generates API key header/param
func (s *strategyServiceImpl) executeAPIKey(strategy *models.Strategy, config map[string]interface{}, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	keyRef := s.getConfigString(config, "key_ref", "api_key")

	apiKey, ok := ctx.Credentials[keyRef]
	if !ok {
		return nil, fmt.Errorf("API key not found: %s", keyRef)
	}

	headerName := s.getConfigString(config, "header_name", "X-API-Key")
	headerPrefix := s.getConfigString(config, "header_prefix", "")

	result := &StrategyResult{
		Values: map[string]interface{}{
			"api_key": apiKey,
		},
		Headers: map[string]string{
			headerName: headerPrefix + apiKey,
		},
	}

	return result, nil
}

// executePayload transforms input parameters into a specific payload format
func (s *strategyServiceImpl) executePayload(strategy *models.Strategy, config map[string]interface{}, ctx *StrategyExecutionContext) (*StrategyResult, error) {
	// Get field mappings from config
	mappings, ok := config["mappings"].(map[string]interface{})
	if !ok {
		mappings = make(map[string]interface{})
	}

	payload := make(map[string]interface{})

	// Apply mappings
	for targetField, sourceExpr := range mappings {
		if sourceStr, ok := sourceExpr.(string); ok {
			value := s.resolveExpression(sourceStr, ctx)
			if value != nil {
				payload[targetField] = value
			}
		}
	}

	// Include additional fields from config
	if includes, ok := config["include_fields"].([]interface{}); ok {
		for _, field := range includes {
			if fieldName, ok := field.(string); ok {
				if value, exists := ctx.Input[fieldName]; exists {
					payload[fieldName] = value
				}
			}
		}
	}

	result := &StrategyResult{
		Values: map[string]interface{}{
			"payload": payload,
		},
		Headers: make(map[string]string),
	}

	return result, nil
}

// Helper methods

func (s *strategyServiceImpl) getConfigString(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func (s *strategyServiceImpl) buildMessage(template string, config map[string]interface{}, ctx *StrategyExecutionContext) string {
	if template == "" {
		// Default message construction: sort params and join
		return s.buildDefaultMessage(config, ctx)
	}

	// Replace placeholders in template
	result := template

	// Replace ${credential:key}
	for key, value := range ctx.Credentials {
		result = strings.ReplaceAll(result, "${credential:"+key+"}", value)
	}

	// Replace ${input:key}
	for key, value := range ctx.Input {
		result = strings.ReplaceAll(result, "${input:"+key+"}", fmt.Sprint(value))
	}

	// Replace ${intermediate:key}
	for key, value := range ctx.Intermediate {
		result = strings.ReplaceAll(result, "${intermediate:"+key+"}", fmt.Sprint(value))
	}

	// Replace ${timestamp}
	result = strings.ReplaceAll(result, "${timestamp}", fmt.Sprintf("%d", ctx.Timestamp))

	// Replace ${nonce}
	result = strings.ReplaceAll(result, "${nonce}", ctx.Nonce)

	return result
}

func (s *strategyServiceImpl) buildDefaultMessage(config map[string]interface{}, ctx *StrategyExecutionContext) string {
	// Get message parts from config
	var parts []string
	separator := s.getConfigString(config, "separator", "")

	// Add timestamp if configured
	if includeTimestamp, ok := config["include_timestamp"].(bool); ok && includeTimestamp {
		parts = append(parts, fmt.Sprintf("%d", ctx.Timestamp))
	}

	// Add nonce if configured
	if includeNonce, ok := config["include_nonce"].(bool); ok && includeNonce {
		parts = append(parts, ctx.Nonce)
	}

	// Add specified fields in order
	if fields, ok := config["message_fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldName, ok := field.(string); ok {
				if value, exists := ctx.Input[fieldName]; exists {
					parts = append(parts, fmt.Sprint(value))
				}
			}
		}
	} else {
		// Default: include all input params sorted alphabetically
		keys := make([]string, 0, len(ctx.Input))
		for k := range ctx.Input {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, fmt.Sprint(ctx.Input[k]))
		}
	}

	return strings.Join(parts, separator)
}

func (s *strategyServiceImpl) computeHMAC(key, message []byte, algorithm, encoding string) (string, error) {
	var h hash.Hash

	switch strings.ToUpper(algorithm) {
	case "SHA256":
		h = hmac.New(sha256.New, key)
	case "SHA512":
		h = hmac.New(sha512.New, key)
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	h.Write(message)
	signature := h.Sum(nil)

	switch strings.ToLower(encoding) {
	case "hex":
		return hex.EncodeToString(signature), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(signature), nil
	default:
		return hex.EncodeToString(signature), nil
	}
}

func (s *strategyServiceImpl) resolveExpression(expr string, ctx *StrategyExecutionContext) interface{} {
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) != 2 {
		return expr
	}

	source := parts[0]
	key := parts[1]

	switch source {
	case "credential":
		if val, ok := ctx.Credentials[key]; ok {
			return val
		}
	case "input":
		if val, ok := ctx.Input[key]; ok {
			return val
		}
	case "intermediate":
		if val, ok := ctx.Intermediate[key]; ok {
			return val
		}
	case "timestamp":
		return ctx.Timestamp
	case "nonce":
		return ctx.Nonce
	case "static":
		return key
	}

	return nil
}

// GenerateNonce generates a unique nonce for replay protection
func GenerateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
