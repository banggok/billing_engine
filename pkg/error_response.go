// pkg/error_response.go
package pkg

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}
