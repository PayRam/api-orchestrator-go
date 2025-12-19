package model

// PipelineRequest represents the complete context needed for orchestration at runtime
// This is NOT a database model - it's used only during pipeline execution
type PipelineRequest struct {
	ProviderName string                 // e.g., "banxa", "transak"
	Action       string                 // e.g., "get_quote", "create_order"
	Params       map[string]interface{} // User-provided parameters
	Headers      map[string]string      // Optional user headers
}
