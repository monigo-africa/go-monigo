package monigo_test

import (
	"errors"
	"testing"

	monigo "github.com/monigo-africa/go-monigo"
)

func apiErr(status int, msg string) *monigo.APIError {
	return &monigo.APIError{StatusCode: status, Message: msg}
}

func TestAPIError_Error(t *testing.T) {
	e := apiErr(404, "not found")
	got := e.Error()
	if got == "" {
		t.Error("Error() returned empty string")
	}
}

func TestAPIError_ErrorWithDetails(t *testing.T) {
	e := &monigo.APIError{
		StatusCode: 400,
		Message:    "validation failed",
		Details:    map[string]string{"name": "required"},
	}
	got := e.Error()
	if got == "" {
		t.Error("Error() returned empty string")
	}
}

func TestIsNotFound(t *testing.T) {
	if !monigo.IsNotFound(apiErr(404, "not found")) {
		t.Error("IsNotFound should be true for 404")
	}
	if monigo.IsNotFound(apiErr(400, "bad request")) {
		t.Error("IsNotFound should be false for 400")
	}
	if monigo.IsNotFound(errors.New("plain error")) {
		t.Error("IsNotFound should be false for non-APIError")
	}
	if monigo.IsNotFound(nil) {
		t.Error("IsNotFound should be false for nil")
	}
}

func TestIsUnauthorized(t *testing.T) {
	if !monigo.IsUnauthorized(apiErr(401, "unauthorized")) {
		t.Error("IsUnauthorized should be true for 401")
	}
	if monigo.IsUnauthorized(apiErr(403, "forbidden")) {
		t.Error("IsUnauthorized should be false for 403")
	}
}

func TestIsForbidden(t *testing.T) {
	if !monigo.IsForbidden(apiErr(403, "forbidden")) {
		t.Error("IsForbidden should be true for 403")
	}
	if monigo.IsForbidden(apiErr(401, "unauthorized")) {
		t.Error("IsForbidden should be false for 401")
	}
}

func TestIsConflict(t *testing.T) {
	if !monigo.IsConflict(apiErr(409, "conflict")) {
		t.Error("IsConflict should be true for 409")
	}
	if monigo.IsConflict(apiErr(400, "bad request")) {
		t.Error("IsConflict should be false for 400")
	}
}

func TestIsRateLimited(t *testing.T) {
	if !monigo.IsRateLimited(apiErr(429, "too many requests")) {
		t.Error("IsRateLimited should be true for 429")
	}
	if monigo.IsRateLimited(apiErr(500, "server error")) {
		t.Error("IsRateLimited should be false for 500")
	}
}

func TestIsQuotaExceeded(t *testing.T) {
	if !monigo.IsQuotaExceeded(apiErr(402, "quota exceeded")) {
		t.Error("IsQuotaExceeded should be true for 402")
	}
	if monigo.IsQuotaExceeded(apiErr(429, "rate limited")) {
		t.Error("IsQuotaExceeded should be false for 429")
	}
}

func TestIsValidationError(t *testing.T) {
	e := &monigo.APIError{
		StatusCode: 400,
		Message:    "validation failed",
		Details:    map[string]string{"email": "invalid"},
	}
	if !monigo.IsValidationError(e) {
		t.Error("IsValidationError should be true when Details is non-empty")
	}
	if monigo.IsValidationError(apiErr(400, "bad request")) {
		t.Error("IsValidationError should be false when Details is nil")
	}
}
