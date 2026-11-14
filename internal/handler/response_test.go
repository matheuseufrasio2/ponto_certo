package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
)

func TestWriteCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"id": "123"}

	err := handler.WriteCreated(rec, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestWriteNoContent(t *testing.T) {
	rec := httptest.NewRecorder()

	handler.WriteNoContent(rec)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for 204 No Content, got %s", rec.Body.String())
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
		expectedCode   domain.ErrorCode
	}{
		{
			name:           "not found error",
			err:            domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    domain.ErrNotFound.Error(),
			expectedCode:   domain.ErrCodeNotFound,
		},
		{
			name:           "billing not found error",
			err:            domain.ErrBillingNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    domain.ErrBillingNotFound.Error(),
			expectedCode:   domain.ErrCodeBillingNotFound,
		},
		{
			name:           "appointment conflict error",
			err:            domain.ErrAppointmentConflict,
			expectedStatus: http.StatusConflict,
			expectedMsg:    domain.ErrAppointmentConflict.Error(),
			expectedCode:   domain.ErrCodeAppointmentConflict,
		},
		{
			name:           "invalid time slot error",
			err:            domain.ErrInvalidTimeSlot,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidTimeSlot.Error(),
			expectedCode:   domain.ErrCodeInvalidTimeSlot,
		},
		{
			name:           "invalid payment method error",
			err:            domain.ErrInvalidPaymentMethod,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidPaymentMethod.Error(),
			expectedCode:   domain.ErrCodeInvalidPaymentMethod,
		},
		{
			name:           "user not found error",
			err:            domain.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    domain.ErrUserNotFound.Error(),
			expectedCode:   domain.ErrCodeUserNotFound,
		},
		{
			name:           "unauthorized error",
			err:            domain.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    domain.ErrUnauthorized.Error(),
			expectedCode:   domain.ErrCodeUnauthorized,
		},
		{
			name:           "forbidden error",
			err:            domain.ErrForbidden,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    domain.ErrForbidden.Error(),
			expectedCode:   domain.ErrCodeForbidden,
		},
		{
			name:           "invalid role error",
			err:            domain.ErrInvalidRole,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidRole.Error(),
			expectedCode:   domain.ErrCodeInvalidRole,
		},
		{
			name:           "invalid uuid error",
			err:            domain.ErrInvalidUUID,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidUUID.Error(),
			expectedCode:   domain.ErrCodeInvalidUUID,
		},
		{
			name:           "unit name required error",
			err:            domain.ErrUnitNameRequired,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrUnitNameRequired.Error(),
			expectedCode:   domain.ErrCodeUnitNameRequired,
		},
		{
			name:           "invalid chair count error",
			err:            domain.ErrInvalidChairCount,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidChairCount.Error(),
			expectedCode:   domain.ErrCodeInvalidChairCount,
		},
		{
			name:           "unit inactive error",
			err:            domain.ErrUnitInactive,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    domain.ErrUnitInactive.Error(),
			expectedCode:   domain.ErrCodeUnitInactive,
		},
		{
			name:           "unit name already exists error",
			err:            domain.ErrUnitNameAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedMsg:    domain.ErrUnitNameAlreadyExists.Error(),
			expectedCode:   domain.ErrCodeUnitNameAlreadyExists,
		},
		{
			name:           "manager email taken error",
			err:            domain.ErrManagerEmailTaken,
			expectedStatus: http.StatusConflict,
			expectedMsg:    domain.ErrManagerEmailTaken.Error(),
			expectedCode:   domain.ErrCodeManagerEmailTaken,
		},
		{
			name:           "unit not in org error",
			err:            domain.ErrUnitNotInOrg,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    domain.ErrUnitNotInOrg.Error(),
			expectedCode:   domain.ErrCodeUnitNotInOrg,
		},
		{
			name:           "barber not in org error",
			err:            domain.ErrBarberNotInOrg,
			expectedStatus: http.StatusForbidden,
			expectedMsg:    domain.ErrBarberNotInOrg.Error(),
			expectedCode:   domain.ErrCodeBarberNotInOrg,
		},
		{
			name:           "invalid commission scope error",
			err:            domain.ErrInvalidCommissionScope,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidCommissionScope.Error(),
			expectedCode:   domain.ErrCodeInvalidCommissionScope,
		},
		{
			name:           "unexpected server error (sanitized to 500)",
			err:            errors.New("raw postgres connection pool exhausted"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
			expectedCode:   domain.ErrCodeInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.HandleError(rec, tc.err)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			var errRes handler.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&errRes); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if errRes.Error != tc.expectedMsg {
				t.Errorf("expected error message %q, got %q", tc.expectedMsg, errRes.Error)
			}

			if errRes.Code != tc.expectedCode {
				t.Errorf("expected error code %q, got %q", tc.expectedCode, errRes.Code)
			}

			if errRes.Message == "" {
				t.Errorf("expected non-empty friendly message for %v", tc.err)
			}
		})
	}
}

func TestHandleErrorWithLocale(t *testing.T) {
	t.Run("resolves English message when Accept-Language is en-US", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		handler.HandleError(rec, domain.ErrBarberNotFound, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}

		var errRes handler.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errRes); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		if errRes.Error != "barber not found" {
			t.Errorf("expected technical error 'barber not found', got %q", errRes.Error)
		}

		if errRes.Message != "Barber professional not found." {
			t.Errorf("expected English message, got %q", errRes.Message)
		}

		if errRes.Code != domain.ErrCodeBarberNotFound {
			t.Errorf("expected code %q, got %q", domain.ErrCodeBarberNotFound, errRes.Code)
		}
	})

	t.Run("resolves Portuguese message when Accept-Language is pt-BR", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

		handler.HandleError(rec, domain.ErrBarberNotFound, req)

		var errRes handler.ErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&errRes)

		if errRes.Message != "Profissional barbeiro não encontrado." {
			t.Errorf("expected Portuguese message, got %q", errRes.Message)
		}
	})
}

func TestWriteValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	valErr := domain.NewValidationError()
	valErr.Add("price_in_cents", "min_value")
	valErr.Add("unit_id", "required")

	handler.HandleError(rec, valErr)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var errRes handler.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errRes); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errRes.Code != domain.ErrCodeValidationFailed {
		t.Errorf("expected code %q, got %q", domain.ErrCodeValidationFailed, errRes.Code)
	}

	if errRes.Fields["price_in_cents"] != "min_value" {
		t.Errorf("expected field price_in_cents min_value, got %s", errRes.Fields["price_in_cents"])
	}

	if errRes.Fields["unit_id"] != "required" {
		t.Errorf("expected field unit_id required, got %s", errRes.Fields["unit_id"])
	}
}

func TestReadJSON(t *testing.T) {
	type Sample struct {
		Name string `json:"name"`
	}

	t.Run("valid json body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"John"}`))
		rec := httptest.NewRecorder()

		var s Sample
		if err := handler.ReadJSON(rec, req, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Name != "John" {
			t.Errorf("expected name John, got %s", s.Name)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		rec := httptest.NewRecorder()

		var s Sample
		if err := handler.ReadJSON(rec, req, &s); err == nil {
			t.Fatal("expected error on empty body, got nil")
		}
	})

	t.Run("disallow unknown fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"John","extra":"field"}`))
		rec := httptest.NewRecorder()

		var s Sample
		if err := handler.ReadJSON(rec, req, &s); err == nil {
			t.Fatal("expected error on unknown field, got nil")
		}
	})
}
