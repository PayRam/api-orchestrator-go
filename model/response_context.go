package model

// ResponseContext holds the normalized response after mapping transformations
// This is NOT a database model - it's created by the response mapper
type ResponseContext struct {
	Success      bool                   // Whether the API call was successful
	StatusCode   int                    // HTTP status code
	RawResponse  map[string]interface{} // Original response body
	MappedData   map[string]interface{} // Transformed data using response mappings
	ErrorMessage string                 // Error message if any
}
