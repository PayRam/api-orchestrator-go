package response

import (
	"fmt"
	"sort"

	"github.com/PayRam/api-orchestrator-go/internal/models"
	"github.com/PayRam/api-orchestrator-go/internal/utils"
	"go.uber.org/zap"
)

// PipelineStep defines the interface for a single step in the response pipeline
type PipelineStep interface {
	// Name returns the name of this step for logging
	Name() string

	// Execute runs this step on the context
	Execute(ctx *ResponseContext) error
}

// =====================================================
// LoadMappingsStep - Loads and sorts response mappings
// =====================================================

// MappingLoader is an interface for loading response mappings
type MappingLoader interface {
	LoadMappings(providerID, action string) ([]*models.ResponseMapping, error)
}

// LoadMappingsStep loads response mappings from a data source
type LoadMappingsStep struct {
	loader MappingLoader
	logger *zap.Logger
}

// NewLoadMappingsStep creates a new LoadMappingsStep
func NewLoadMappingsStep(loader MappingLoader, logger *zap.Logger) *LoadMappingsStep {
	return &LoadMappingsStep{
		loader: loader,
		logger: logger,
	}
}

// Name returns the step name
func (s *LoadMappingsStep) Name() string {
	return "LoadMappingsStep"
}

// Execute loads and sorts mappings by priority
func (s *LoadMappingsStep) Execute(ctx *ResponseContext) error {
	s.logger.Debug("Loading response mappings",
		zap.String("provider_id", ctx.Provider.ID),
		zap.String("action", ctx.Action),
	)

	mappings, err := s.loader.LoadMappings(ctx.Provider.ID, ctx.Action)
	if err != nil {
		return fmt.Errorf("failed to load mappings: %w", err)
	}

	// Sort by priority (higher priority first)
	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Priority > mappings[j].Priority
	})

	ctx.SetMappings(mappings)

	s.logger.Debug("Loaded response mappings",
		zap.Int("count", len(mappings)),
	)

	return nil
}

// =====================================================
// JSONPathExtractStep - Extracts values using JSONPath
// =====================================================

// JSONPathExtractStep extracts values from the raw response using JSONPath
type JSONPathExtractStep struct {
	logger *zap.Logger
}

// NewJSONPathExtractStep creates a new JSONPathExtractStep
func NewJSONPathExtractStep(logger *zap.Logger) *JSONPathExtractStep {
	return &JSONPathExtractStep{
		logger: logger,
	}
}

// Name returns the step name
func (s *JSONPathExtractStep) Name() string {
	return "JSONPathExtractStep"
}

// Execute extracts values from the raw response using JSONPath
func (s *JSONPathExtractStep) Execute(ctx *ResponseContext) error {
	s.logger.Debug("Extracting values using JSONPath",
		zap.Int("mappings_count", len(ctx.Mappings)),
	)

	extractedCount := 0
	missingRequired := []string{}

	for _, mapping := range ctx.Mappings {
		// Normalize the JSONPath
		jsonPath := mapping.NormalizeJSONPath()

		// Try to extract the value
		value, err := utils.ExtractJSONPath(ctx.RawResponse, jsonPath)
		if err != nil {
			s.logger.Debug("JSONPath extraction failed",
				zap.String("target_field", mapping.TargetField),
				zap.String("json_path", jsonPath),
				zap.Error(err),
			)

			// Use default value if available
			if mapping.HasDefaultValue() {
				ctx.SetExtractedValue(mapping.TargetField, mapping.DefaultValue)
				extractedCount++
				continue
			}

			// Track missing required fields
			if mapping.IsRequired {
				missingRequired = append(missingRequired, mapping.TargetField)
			}
			continue
		}

		ctx.SetExtractedValue(mapping.TargetField, value)
		extractedCount++

		s.logger.Debug("Value extracted",
			zap.String("target_field", mapping.TargetField),
			zap.Any("value", value),
		)
	}

	// Check for missing required fields
	if len(missingRequired) > 0 {
		return fmt.Errorf("missing required fields: %v", missingRequired)
	}

	s.logger.Debug("JSONPath extraction complete",
		zap.Int("extracted_count", extractedCount),
	)

	return nil
}

// =====================================================
// TransformStep - Applies transformations to values
// =====================================================

// TransformStep applies transformation functions to extracted values
type TransformStep struct {
	transformer *utils.Transformer
	logger      *zap.Logger
}

// NewTransformStep creates a new TransformStep
func NewTransformStep(logger *zap.Logger) *TransformStep {
	return &TransformStep{
		transformer: utils.NewTransformer(),
		logger:      logger,
	}
}

// Name returns the step name
func (s *TransformStep) Name() string {
	return "TransformStep"
}

// Execute applies transformations to extracted values
func (s *TransformStep) Execute(ctx *ResponseContext) error {
	s.logger.Debug("Applying transformations")

	transformedCount := 0

	for _, mapping := range ctx.Mappings {
		// Get the extracted value
		value, ok := ctx.GetExtractedValue(mapping.TargetField)
		if !ok {
			continue
		}

		// Apply transformation if specified
		if mapping.HasTransform() {
			transformedValue := s.transformer.Apply(value, mapping.Transform)
			ctx.SetTransformedValue(mapping.TargetField, transformedValue)
			transformedCount++

			s.logger.Debug("Transformation applied",
				zap.String("target_field", mapping.TargetField),
				zap.String("transform", mapping.Transform),
				zap.Any("original", value),
				zap.Any("transformed", transformedValue),
			)
		} else {
			// Copy value as-is to transformed values
			ctx.SetTransformedValue(mapping.TargetField, value)
		}
	}

	s.logger.Debug("Transformations complete",
		zap.Int("transformed_count", transformedCount),
	)

	return nil
}

// =====================================================
// ValidateStep - Validates required fields
// =====================================================

// ValidateStep validates that all required fields are present
type ValidateStep struct {
	logger *zap.Logger
}

// NewValidateStep creates a new ValidateStep
func NewValidateStep(logger *zap.Logger) *ValidateStep {
	return &ValidateStep{
		logger: logger,
	}
}

// Name returns the step name
func (s *ValidateStep) Name() string {
	return "ValidateStep"
}

// Execute validates that all required fields are present
func (s *ValidateStep) Execute(ctx *ResponseContext) error {
	s.logger.Debug("Validating required fields")

	missingFields := []string{}

	for _, mapping := range ctx.Mappings {
		if !mapping.IsRequired {
			continue
		}

		value, ok := ctx.GetFinalValue(mapping.TargetField)
		if !ok || value == nil {
			missingFields = append(missingFields, mapping.TargetField)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %v", missingFields)
	}

	s.logger.Debug("Validation complete")
	return nil
}
