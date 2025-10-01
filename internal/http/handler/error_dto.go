package handler

// ErrorResponse represents a generic error response.
// swagger:response errorResponse
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}
