package builder

import (
	"fmt"
	"strings"

	"github.com/PayRam/api-orchestrator-go/internal/context"
	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/services"
	"go.uber.org/zap"
)

// HeaderBuilder is responsible for building HTTP headers dynamically
type HeaderBuilder struct {
	credentialService services.CredentialService
	logger            *zap.Logger
}

// NewHeaderBuilder creates a new header builder
func NewHeaderBuilder(credentialService services.CredentialService, logger *zap.Logger) *HeaderBuilder {
	return &HeaderBuilder{
		credentialService: credentialService,
		logger:            logger,
	}
}

// BuildHeaders constructs HTTP headers based on header rules
func (hb *HeaderBuilder) BuildHeaders(
	ctx *context.OrchestratorContext,
	headerRules []*models.HeaderRule,
	providerID string,
) error {
	hb.logger.Debug("Building headers", zap.Int("rules_count", len(headerRules)))

	for _, rule := range headerRules {
		value, err := hb.resolveHeaderValue(ctx, rule, providerID)
		if err != nil {
			hb.logger.Warn("Failed to resolve header value",
				zap.String("header", rule.HeaderName),
				zap.Error(err))
			continue
		}

		ctx.SetHeader(rule.HeaderName, value)
		hb.logger.Debug("Header set",
			zap.String("header", rule.HeaderName),
			zap.String("expression", rule.ValueExpression))
	}

	return nil
}

// resolveHeaderValue resolves the header value using the ValueExpression
// Expression format examples:
// - "static:Bearer xyz123" - Static value
// - "credential:api_key" - Get from credentials
// - "strategy:signature" - Get from computed strategy values
// - "param:user_id" - Get from input parameters
func (hb *HeaderBuilder) resolveHeaderValue(
	ctx *context.OrchestratorContext,
	rule *models.HeaderRule,
	providerID string,
) (string, error) {

	expr := rule.ValueExpression
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) < 2 {
		// Treat as static value if no prefix
		return expr, nil
	}

	valueType := parts[0]
	valueSource := parts[1]

	switch valueType {
	case "static":
		return valueSource, nil

	case "credential":
		if val, ok := ctx.GetCredential(valueSource); ok {
			return val, nil
		}
		return "", fmt.Errorf("credential not found: %s", valueSource)

	case "strategy":
		if val, ok := ctx.GetOk(valueSource); ok {
			return fmt.Sprint(val), nil
		}
		return "", fmt.Errorf("computed value not found: %s", valueSource)

	case "param":
		if val, ok := ctx.GetInput(valueSource); ok {
			return fmt.Sprint(val), nil
		}
		return "", fmt.Errorf("input param not found: %s", valueSource)

	default:
		return expr, nil
	}
}

// applyTemplate applies a template to a value (e.g., "Bearer {token}" => "Bearer abc123")
func (hb *HeaderBuilder) applyTemplate(template, value string) string {
	if template == "" {
		return value
	}

	// Simple template replacement
	result := strings.ReplaceAll(template, "{value}", value)
	result = strings.ReplaceAll(result, "{token}", value)
	result = strings.ReplaceAll(result, "{api_key}", value)
	result = strings.ReplaceAll(result, "{api_secret}", value)

	return result
}
