package model

// FinalRequest represents the fully constructed HTTP request ready for execution
// This is NOT a database model - it's built dynamically by the request builder
type FinalRequest struct {
	URL     string                 // Complete URL with path and query parameters
	Method  string                 // HTTP method: GET, POST, PUT, DELETE, etc.
	Headers map[string]string      // All resolved headers (provider + rules + user)
	Body    map[string]interface{} // Request body for POST/PUT/PATCH
}
