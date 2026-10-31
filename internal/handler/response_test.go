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
	}{
		{
			name:           "not found error",
			err:            domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    domain.ErrNotFound.Error(),
		},
		{
			name:           "billing not found error",
			err:            domain.ErrBillingNotFound,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    domain.ErrBillingNotFound.Error(),
		},
		{
			name:           "appointment conflict error",
			err:            domain.ErrAppointmentConflict,
			expectedStatus: http.StatusConflict,
			expectedMsg:    domain.ErrAppointmentConflict.Error(),
		},
		{
			name:           "invalid time slot error",
			err:            domain.ErrInvalidTimeSlot,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidTimeSlot.Error(),
		},
		{
			name:           "invalid payment method error",
			err:            domain.ErrInvalidPaymentMethod,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    domain.ErrInvalidPaymentMethod.Error(),
		},
		{
			name:           "unexpected server error (sanitized to 500)",
			err:            errors.New("raw postgres connection pool exhausted"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
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
		})
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

