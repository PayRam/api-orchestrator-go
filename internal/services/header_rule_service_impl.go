package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/repositories"
	"go.uber.org/zap"
)

type headerRuleServiceImpl struct {
	repo            repositories.HeaderRuleRepo
	providerService ProviderService
	logger          *zap.Logger
}

// NewHeaderRuleService creates a new header rule service implementation
func NewHeaderRuleService(
	repo repositories.HeaderRuleRepo,
	providerService ProviderService,
	logger *zap.Logger,
) HeaderRuleService {
	return &headerRuleServiceImpl{
		repo:            repo,
		providerService: providerService,
		logger:          logger,
	}
}

func (s *headerRuleServiceImpl) CreateHeaderRule(rule *models.HeaderRule) error {
	s.logger.Info("Creating header rule",
		zap.String("header_name", rule.HeaderName),
		zap.String("provider_id", rule.ProviderID))

	// Validate provider exists
	_, err := s.providerService.GetProviderByID(rule.ProviderID)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Business logic validations
	if rule.HeaderName == "" {
		return fmt.Errorf("header name cannot be empty")
	}
	if rule.ValueExpression == "" {
		return fmt.Errorf("value expression cannot be empty")
	}

	err = s.repo.Create(rule)
	if err != nil {
		s.logger.Error("Failed to create header rule", zap.Error(err))
		return err
	}

	s.logger.Info("Header rule created successfully", zap.String("id", rule.ID))
	return nil
}

func (s *headerRuleServiceImpl) GetHeaderRuleByID(id string) (*models.HeaderRule, error) {
	s.logger.Debug("Fetching header rule by ID", zap.String("id", id))
	return s.repo.FindByID(id)
}

func (s *headerRuleServiceImpl) GetHeaderRulesByProviderID(providerID string) ([]*models.HeaderRule, error) {
	s.logger.Debug("Fetching header rules by provider ID", zap.String("provider_id", providerID))
	// Rules are already ordered by priority ASC in the repository
	return s.repo.FindByProviderID(providerID)
}

func (s *headerRuleServiceImpl) UpdateHeaderRule(rule *models.HeaderRule) error {
	s.logger.Info("Updating header rule", zap.String("id", rule.ID))

	// Validate header name
	if rule.HeaderName == "" {
		return fmt.Errorf("header name cannot be empty")
	}
	if rule.ValueExpression == "" {
		return fmt.Errorf("value expression cannot be empty")
	}

	err := s.repo.Update(rule)
	if err != nil {
		s.logger.Error("Failed to update header rule", zap.Error(err))
		return err
	}

	s.logger.Info("Header rule updated successfully")
	return nil
}

func (s *headerRuleServiceImpl) DeleteHeaderRule(id string) error {
	s.logger.Info("Deleting header rule", zap.String("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Failed to delete header rule", zap.Error(err))
		return err
	}

	s.logger.Info("Header rule deleted successfully")
	return nil
}

// EvaluateHeaderRules evaluates all header rules for a provider and returns
// a map of header name to resolved value.
func (s *headerRuleServiceImpl) EvaluateHeaderRules(
	providerID string,
	credentials map[string]string,
	strategyValues map[string]interface{},
	inputParams map[string]interface{},
) (map[string]string, error) {
	s.logger.Debug("Evaluating header rules", zap.String("provider_id", providerID))

	// Fetch rules ordered by priority
	rules, err := s.repo.FindByProviderID(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch header rules: %w", err)
	}

	headers := make(map[string]string)

	// Process rules in priority order (already sorted by repo)
	for _, rule := range rules {
		value, err := s.EvaluateExpression(
			rule.ValueExpression,
			credentials,
			strategyValues,
			inputParams,
		)
		if err != nil {
			s.logger.Warn("Failed to evaluate header expression",
				zap.String("header", rule.HeaderName),
				zap.String("expression", rule.ValueExpression),
				zap.Error(err))
			continue
		}

		headers[rule.HeaderName] = value
		s.logger.Debug("Header evaluated",
			zap.String("header", rule.HeaderName),
			zap.Int("priority", rule.Priority))
	}

	s.logger.Debug("Header rules evaluated",
		zap.Int("total_rules", len(rules)),
		zap.Int("headers_set", len(headers)))

	return headers, nil
}

// EvaluateExpression evaluates a single value expression and returns the resolved value.
// Supported expression formats:
//   - "static:value" - returns "value" directly
//   - "credential:key" - looks up key in credentials map
//   - "strategy:key" - looks up key in strategy values map
//   - "param:key" - looks up key in input parameters map
//   - "template:Bearer ${credential:api_key}" - supports interpolation
//   - Plain value without prefix - treated as static value
func (s *headerRuleServiceImpl) EvaluateExpression(
	expression string,
	credentials map[string]string,
	strategyValues map[string]interface{},
	inputParams map[string]interface{},
) (string, error) {
	// Handle template expressions with interpolation
	if strings.HasPrefix(expression, "template:") {
		template := strings.TrimPrefix(expression, "template:")
		return s.evaluateTemplate(template, credentials, strategyValues, inputParams)
	}

	// Split expression into type and source
	parts := strings.SplitN(expression, ":", 2)
	if len(parts) < 2 {
		// Treat as static value if no prefix
		return expression, nil
	}

	valueType := strings.ToLower(parts[0])
	valueSource := parts[1]

	switch valueType {
	case "static":
		return valueSource, nil

	case "credential":
		if credentials == nil {
			return "", fmt.Errorf("credentials map is nil")
		}
		if val, ok := credentials[valueSource]; ok {
			return val, nil
		}
		return "", fmt.Errorf("credential not found: %s", valueSource)

	case "strategy":
		if strategyValues == nil {
			return "", fmt.Errorf("strategy values map is nil")
		}
		if val, ok := strategyValues[valueSource]; ok {
			return fmt.Sprint(val), nil
		}
		return "", fmt.Errorf("strategy value not found: %s", valueSource)

	case "param":
		if inputParams == nil {
			return "", fmt.Errorf("input params map is nil")
		}
		if val, ok := inputParams[valueSource]; ok {
			return fmt.Sprint(val), nil
		}
		return "", fmt.Errorf("input param not found: %s", valueSource)

	default:
		// Unknown prefix, return as-is
		return expression, nil
	}
}

// evaluateTemplate handles template expressions with ${type:key} interpolation
// Example: "Bearer ${credential:api_key}" -> "Bearer abc123"
func (s *headerRuleServiceImpl) evaluateTemplate(
	template string,
	credentials map[string]string,
	strategyValues map[string]interface{},
	inputParams map[string]interface{},
) (string, error) {
	// Regex to match ${type:key} patterns
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	result := template
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[0] // e.g., "${credential:api_key}"
		expr := match[1]        // e.g., "credential:api_key"

		// Recursively evaluate the inner expression
		value, err := s.EvaluateExpression(expr, credentials, strategyValues, inputParams)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate template placeholder %s: %w", placeholder, err)
		}

		result = strings.Replace(result, placeholder, value, 1)
	}

	return result, nil
}
