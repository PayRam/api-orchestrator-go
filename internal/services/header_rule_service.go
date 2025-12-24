package services

import (
	"github.com/PayRam/api-orchestrator-go/internal/models"
)

// HeaderRuleService defines the business logic interface for header rules.
// It handles fetching header rules and evaluating value expressions to generate
// HTTP headers dynamically based on credentials, strategy values, and input parameters.
type HeaderRuleService interface {
	// CreateHeaderRule creates a new header rule for a provider
	CreateHeaderRule(rule *models.HeaderRule) error

	// GetHeaderRuleByID retrieves a header rule by its ID
	GetHeaderRuleByID(id string) (*models.HeaderRule, error)

	// GetHeaderRulesByProviderID retrieves all header rules for a provider, ordered by priority
	GetHeaderRulesByProviderID(providerID string) ([]*models.HeaderRule, error)

	// UpdateHeaderRule updates an existing header rule
	UpdateHeaderRule(rule *models.HeaderRule) error

	// DeleteHeaderRule soft deletes a header rule
	DeleteHeaderRule(id string) error

	// EvaluateHeaderRules evaluates all header rules for a provider and returns
	// a map of header name to resolved value. It uses the provided context
	// for resolving credentials, strategy values, and input parameters.
	EvaluateHeaderRules(
		providerID string,
		credentials map[string]string,
		strategyValues map[string]interface{},
		inputParams map[string]interface{},
	) (map[string]string, error)

	// EvaluateExpression evaluates a single value expression and returns the resolved value.
	// Expression format: "type:source" where type can be:
	//   - "static" - returns the source value directly
	//   - "credential" - looks up source in credentials map
	//   - "strategy" - looks up source in strategy values map
	//   - "param" - looks up source in input parameters map
	//   - "template" - supports interpolation like "Bearer ${credential:api_key}"
	EvaluateExpression(
		expression string,
		credentials map[string]string,
		strategyValues map[string]interface{},
		inputParams map[string]interface{},
	) (string, error)
}
