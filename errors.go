package monigo

import (
	"errors"
	"fmt"
)

// APIError is returned when the Monigo API responds with an HTTP 4xx or 5xx status.
type APIError struct {
	// StatusCode is the HTTP status code (e.g. 404, 422).
	StatusCode int `json:"-"`
	// Message is the human-readable error description from the API.
	Message string `json:"error"`
	// Details contains field-level validation errors when present.
	Details map[string]string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("monigo: HTTP %d: %s (%v)", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("monigo: HTTP %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if err is an APIError with status 404.
func IsNotFound(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 404
}

// IsUnauthorized returns true if err is an APIError with status 401.
func IsUnauthorized(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 401
}

// IsForbidden returns true if err is an APIError with status 403.
func IsForbidden(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 403
}

// IsConflict returns true if err is an APIError with status 409.
// Commonly returned when a subscription already exists for a customer.
func IsConflict(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 409
}

// IsRateLimited returns true if err is an APIError with status 429.
func IsRateLimited(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 429
}

// IsQuotaExceeded returns true if err is an APIError with status 402.
// This is returned when the organisation's event quota is exhausted.
func IsQuotaExceeded(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 402
}

// IsValidationError returns true if err is an APIError with status 400
// that includes field-level Details.
func IsValidationError(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 400 && len(e.Details) > 0
}
