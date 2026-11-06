package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// AppointmentHandler manages HTTP endpoints for appointments.
type AppointmentHandler struct {
	appointmentService *service.AppointmentService
}

// NewAppointmentHandler creates a new AppointmentHandler.
func NewAppointmentHandler(appointmentService *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointmentService: appointmentService}
}

// ListByUnit handles GET /api/v1/reception/appointments?unit_id={id}&date={YYYY-MM-DD}.
func (h *AppointmentHandler) ListByUnit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := r.URL.Query().Get("unit_id")
	if unitID == "" && claims.UnitID != nil {
		unitID = *claims.UnitID
	}

	if !middleware.RequireUnitAccess(unitID, claims) {
		HandleError(w, domain.ErrForbidden)
		return
	}

	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now().UTC()
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			targetDate = parsedDate
		}
	}

	appointments, err := h.appointmentService.ListUnitAppointments(r.Context(), unitID, targetDate)
	if err != nil {
		HandleError(w, err)
		return
	}

	if appointments == nil {
		appointments = []domain.AppointmentDetails{}
	}

	_ = WriteJSON(w, http.StatusOK, appointments)
}

// Create handles POST /api/v1/reception/appointments.
func (h *AppointmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	var input domain.CreateAppointmentInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.UnitID == "" && claims.UnitID != nil {
		input.UnitID = *claims.UnitID
	}

	if !middleware.RequireUnitAccess(input.UnitID, claims) {
		HandleError(w, domain.ErrForbidden)
		return
	}

	created, err := h.appointmentService.CreateAppointment(r.Context(), input)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteCreated(w, created)
}

// ListMyQueue handles GET /api/v1/barber/my-queue.
func (h *AppointmentHandler) ListMyQueue(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now().UTC()
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			targetDate = parsedDate
		}
	}

	queue, err := h.appointmentService.ListBarberQueue(r.Context(), claims.UserID, targetDate)
	if err != nil {
		HandleError(w, err)
		return
	}

	if queue == nil {
		queue = []domain.AppointmentDetails{}
	}

	_ = WriteJSON(w, http.StatusOK, queue)
}

// CompleteCut handles POST /api/v1/barber/appointments/{id}/complete.
func (h *AppointmentHandler) CompleteCut(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	appointmentID := chi.URLParam(r, "id")
	if appointmentID == "" {
		WriteError(w, http.StatusBadRequest, "appointment id is required")
		return
	}

	err := h.appointmentService.CompleteCut(r.Context(), appointmentID, claims.UserID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Atendimento concluído com sucesso",
	})
}
