package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// DecodeAndValidate decodes the JSON request body into dst and runs struct
// validation. On failure it writes the appropriate error response and returns
// false, so handlers can simply `if !DecodeAndValidate(...) { return }`.
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1MB cap
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		BadRequest(w, "Request body is not valid JSON: "+err.Error())
		return false
	}

	if err := validate.Struct(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			ValidationError(w, fieldErrors(ve))
			return false
		}
		BadRequest(w, err.Error())
		return false
	}
	return true
}

// fieldErrors converts validator errors into a friendly field -> message map.
func fieldErrors(ve validator.ValidationErrors) map[string]string {
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		field := strings.ToLower(fe.Field())
		out[field] = messageFor(fe)
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required."
	case "email":
		return "Must be a valid email address."
	case "min":
		return "Must be at least " + fe.Param() + " characters."
	case "max":
		return "Must be at most " + fe.Param() + " characters."
	case "oneof":
		return "Must be one of: " + fe.Param() + "."
	default:
		return "Invalid value."
	}
}
