package pipeline

import (
	"fmt"

	"github.com/PayRam/api-orchestrator-go/internal/context"
	"go.uber.org/zap"
)

// Step represents a single step in the orchestration pipeline
type Step interface {
	Execute(ctx *context.OrchestratorContext) error
	Name() string
}

// Pipeline represents an orchestration pipeline
type Pipeline struct {
	steps  []Step
	logger *zap.Logger
}

// NewPipeline creates a new pipeline
func NewPipeline(logger *zap.Logger) *Pipeline {
	return &Pipeline{
		steps:  make([]Step, 0),
		logger: logger,
	}
}

// AddStep adds a step to the pipeline
func (p *Pipeline) AddStep(step Step) {
	p.steps = append(p.steps, step)
}

// Execute executes all steps in the pipeline
func (p *Pipeline) Execute(ctx *context.OrchestratorContext) error {
	p.logger.Info("Starting pipeline execution",
		zap.Int("steps_count", len(p.steps)),
		zap.String("request_id", ctx.RequestID.String()))

	for i, step := range p.steps {
		p.logger.Debug("Executing step",
			zap.Int("step_number", i+1),
			zap.String("step_name", step.Name()))

		err := step.Execute(ctx)
		if err != nil {
			p.logger.Error("Pipeline step failed",
				zap.String("step_name", step.Name()),
				zap.Error(err))
			return fmt.Errorf("step '%s' failed: %w", step.Name(), err)
		}

		p.logger.Debug("Step completed", zap.String("step_name", step.Name()))
	}

	p.logger.Info("Pipeline execution completed successfully")
	return nil
}

// StrategyStep represents a strategy execution step
type StrategyStep struct {
	name    string
	handler func(ctx *context.OrchestratorContext) error
}

// NewStrategyStep creates a new strategy step
func NewStrategyStep(name string, handler func(ctx *context.OrchestratorContext) error) *StrategyStep {
	return &StrategyStep{
		name:    name,
		handler: handler,
	}
}

// Execute executes the strategy
func (s *StrategyStep) Execute(ctx *context.OrchestratorContext) error {
	return s.handler(ctx)
}

// Name returns the step name
func (s *StrategyStep) Name() string {
	return s.name
}
