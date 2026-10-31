package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
)

// ErrorResponse represents the standardized JSON payload for API errors.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON sends a standardized JSON response with the provided HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteCreated sends a standardized 201 Created response with the provided payload.
func WriteCreated(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusCreated, data)
}

// WriteNoContent sends a standardized 204 No Content response without a body.
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteError sends a standardized error JSON response with the provided HTTP status code.
func WriteError(w http.ResponseWriter, status int, message string) {
	_ = WriteJSON(w, status, ErrorResponse{Error: message})
}

// HandleError maps domain and system errors to standardized HTTP status codes.
// It securely masks unexpected 500 errors and logs them with structured context.
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrBarberNotFound),
		errors.Is(err, domain.ErrCustomerNotFound),
		errors.Is(err, domain.ErrClientNotFound),
		errors.Is(err, domain.ErrServiceNotFound),
		errors.Is(err, domain.ErrBillingNotFound):
		WriteError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, domain.ErrAppointmentConflict):
		WriteError(w, http.StatusConflict, err.Error())

	case errors.Is(err, domain.ErrInvalidTimeSlot),
		errors.Is(err, domain.ErrInvalidPaymentMethod),
		errors.Is(err, domain.ErrInvalidBillingStatus):
		WriteError(w, http.StatusBadRequest, err.Error())

	default:
		slog.Error("unhandled server error", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ReadJSON decodes the request body into the destination pointer (dst).
// It enforces a 1MB payload limit and rejects unknown JSON fields.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit request body to 1MB to prevent memory DoS attacks
	maxBytes := int64(1 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains malformed JSON")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("field %q has invalid type (expected %s)", unmarshalTypeError.Field, unmarshalTypeError.Type)
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("request body exceeds maximum limit of %d bytes", maxBytes)
		default:
			return err
		}
	}

	// Ensure there is only a single JSON value in the body
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}
