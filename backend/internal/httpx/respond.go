// Package httpx provides small helpers for writing consistent JSON responses
// and a single error shape used across every endpoint.
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse is the consistent error envelope returned by every endpoint.
//
//	{ "error": { "code": "validation_error", "message": "...", "fields": {...} } }
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// JSON writes v as JSON with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// Error writes the standard error envelope.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

// ValidationError writes a 422 with per-field messages.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Error: ErrorBody{
			Code:    "validation_error",
			Message: "One or more fields are invalid.",
			Fields:  fields,
		},
	})
}

// Common shorthand errors.
func BadRequest(w http.ResponseWriter, msg string) {
	Error(w, http.StatusBadRequest, "bad_request", msg)
}
func Unauthorized(w http.ResponseWriter, msg string) {
	Error(w, http.StatusUnauthorized, "unauthorized", msg)
}
func Forbidden(w http.ResponseWriter, msg string) { Error(w, http.StatusForbidden, "forbidden", msg) }
func NotFound(w http.ResponseWriter, msg string)  { Error(w, http.StatusNotFound, "not_found", msg) }
func Conflict(w http.ResponseWriter, msg string)  { Error(w, http.StatusConflict, "conflict", msg) }

func Internal(w http.ResponseWriter, err error) {
	slog.Error("internal server error", "error", err)
	Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong on our end.")
}
