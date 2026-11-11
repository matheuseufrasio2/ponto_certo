package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
)

// ErrorResponse represents the standardized JSON payload for API errors.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    domain.ErrorCode  `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
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
func WriteError(w http.ResponseWriter, status int, message string, code ...domain.ErrorCode) {
	WriteErrorWithLocale(w, status, message, i18n.LocalePTBR, code...)
}

// WriteErrorWithLocale sends a standardized error JSON response resolving friendly message for the given locale.
func WriteErrorWithLocale(w http.ResponseWriter, status int, message string, loc i18n.Locale, code ...domain.ErrorCode) {
	var errCode domain.ErrorCode
	if len(code) > 0 && code[0] != "" {
		errCode = code[0]
	} else {
		errCode = domain.CodeForMessage(message)
		if errCode == "" && status == http.StatusInternalServerError {
			errCode = domain.ErrCodeInternalServerError
		}
	}

	friendlyMsg := i18n.Translate(loc, errCode)
	if friendlyMsg == "" || errCode == "" {
		friendlyMsg = message
	}

	_ = WriteJSON(w, status, ErrorResponse{
		Error:   message,
		Message: friendlyMsg,
		Code:    errCode,
	})
}

// WriteValidationError sends a standardized 400 Bad Request response with structured field errors.
func WriteValidationError(w http.ResponseWriter, valErr *domain.ValidationError, loc ...i18n.Locale) {
	targetLoc := i18n.LocalePTBR
	if len(loc) > 0 {
		targetLoc = loc[0]
	}

	_ = WriteJSON(w, http.StatusBadRequest, ErrorResponse{
		Error:   valErr.Error(),
		Message: i18n.Translate(targetLoc, domain.ErrCodeValidationFailed),
		Code:    domain.ErrCodeValidationFailed,
		Fields:  valErr.Fields,
	})
}

// HandleError maps domain and system errors to standardized HTTP status codes.
// It supports an optional *http.Request to automatically resolve the client locale via Accept-Language.
func HandleError(w http.ResponseWriter, err error, r ...*http.Request) {
	var req *http.Request
	if len(r) > 0 {
		req = r[0]
	}
	loc := i18n.ResolveLocale(req)

	var valErr *domain.ValidationError
	if errors.As(err, &valErr) {
		slog.Warn("validation error (400)", "fields", valErr.Fields)
		WriteValidationError(w, valErr, loc)
		return
	}

	code := domain.CodeForError(err)

	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrBarberNotFound),
		errors.Is(err, domain.ErrCustomerNotFound),
		errors.Is(err, domain.ErrClientNotFound),
		errors.Is(err, domain.ErrServiceNotFound),
		errors.Is(err, domain.ErrBillingNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrOrganizationNotFound),
		errors.Is(err, domain.ErrAppointmentNotFound),
		errors.Is(err, domain.ErrScheduleBlockNotFound),
		errors.Is(err, domain.ErrEmployeeLeaveNotFound),
		errors.Is(err, domain.ErrUnitNotFound):
		slog.Warn("resource not found (404)", "error", err.Error(), "code", code)
		WriteErrorWithLocale(w, http.StatusNotFound, err.Error(), loc, code)

	case errors.Is(err, domain.ErrUnauthorized):
		slog.Warn("unauthorized request (401)", "error", err.Error(), "code", code)
		WriteErrorWithLocale(w, http.StatusUnauthorized, err.Error(), loc, code)

	case errors.Is(err, domain.ErrForbidden),
		errors.Is(err, domain.ErrCannotImpersonateSelf),
		errors.Is(err, domain.ErrCannotImpersonateSuperAdmin),
		errors.Is(err, domain.ErrImpersonationActionBlocked),
		errors.Is(err, domain.ErrTenantInactive),
		errors.Is(err, domain.ErrUnitInactive),
		errors.Is(err, domain.ErrCannotDeactivateSuperAdmin),
		errors.Is(err, domain.ErrCannotDeactivateSelf),
		errors.Is(err, domain.ErrInsufficientRoleHierarchy):
		slog.Warn("forbidden request (403)", "error", err.Error(), "code", code)
		WriteErrorWithLocale(w, http.StatusForbidden, err.Error(), loc, code)

	case errors.Is(err, domain.ErrAppointmentConflict),
		errors.Is(err, domain.ErrBarberHasFutureAppointments),
		errors.Is(err, domain.ErrScheduleBlockConflict),
		errors.Is(err, domain.ErrEmployeeLeaveConflict),
		errors.Is(err, domain.ErrUnitNameAlreadyExists),
		errors.Is(err, domain.ErrBillingAlreadyProcessed):
		slog.Warn("conflict detected (409)", "error", err.Error(), "code", code)
		WriteErrorWithLocale(w, http.StatusConflict, err.Error(), loc, code)

	case errors.Is(err, domain.ErrInvalidTimeSlot),
		errors.Is(err, domain.ErrInvalidPaymentMethod),
		errors.Is(err, domain.ErrInvalidBillingStatus),
		errors.Is(err, domain.ErrInvalidBarberStatus),
		errors.Is(err, domain.ErrInvalidAppointmentTransition),
		errors.Is(err, domain.ErrInvalidAppointmentTime),
		errors.Is(err, domain.ErrInvalidDateRange),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidDocument),
		errors.Is(err, domain.ErrInvalidLogoURL),
		errors.Is(err, domain.ErrUnitNameRequired),
		errors.Is(err, domain.ErrInvalidChairCount),
		errors.Is(err, domain.ErrValidationFailed):
		slog.Warn("bad request validation failed (400)", "error", err.Error(), "code", code)
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), loc, code)

	default:
		slog.Error("unhandled server error (500)", "error", err)
		WriteErrorWithLocale(w, http.StatusInternalServerError, "internal server error", loc, domain.ErrCodeInternalServerError)
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

// IsValidUUID validates that a string is a standard 36-character canonical UUID (8-4-4-4-12 hex).
func IsValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}
	for i, r := range u {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				return false
			}
		}
	}
	return true
}

// GetClientIP extracts client IP address checking X-Forwarded-For, X-Real-IP or RemoteAddr.
func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// GetUserAgent extracts the User-Agent header string.
func GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}
