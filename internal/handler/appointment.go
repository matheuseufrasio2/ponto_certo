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

// ListByUnit handles GET /api/v1/reception/appointments.
// @Summary Listar agendamentos da recepção
// @Description Retorna a lista de agendamentos de uma unidade para a data informada (padrão: hoje).
// @Tags Reception
// @Security BearerAuth
// @Produce json
// @Param unit_id query string false "ID da unidade (opcional se usuário já vinculado a uma unidade)"
// @Param date query string false "Data no formato YYYY-MM-DD (padrão: hoje)"
// @Success 200 {array} domain.AppointmentDetails "Lista de agendamentos detalhados"
// @Failure 401 {object} ErrorResponse "Não autenticado"
// @Failure 403 {object} ErrorResponse "Acesso não permitido a esta unidade"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /reception/appointments [get]
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
// @Summary Criar novo agendamento ou encaixe
// @Description Cria um agendamento na unidade para cliente novo ou existente, com validação de colisão de horário.
// @Tags Reception
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.CreateAppointmentInput true "Dados do agendamento"
// @Success 201 {object} domain.Appointment "Agendamento criado com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos ou corpo malformado"
// @Failure 401 {object} ErrorResponse "Não autenticado"
// @Failure 403 {object} ErrorResponse "Acesso não permitido"
// @Failure 409 {object} ErrorResponse "Barbeiro já possui agendamento neste horário"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /reception/appointments [post]
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
// @Summary Listar fila diária do barbeiro
// @Description Retorna todos os agendamentos atribuídos ao barbeiro autenticado no dia selecionado.
// @Tags Barber
// @Security BearerAuth
// @Produce json
// @Param date query string false "Data no formato YYYY-MM-DD (padrão: hoje)"
// @Success 200 {array} domain.AppointmentDetails "Fila de agendamentos do profissional"
// @Failure 401 {object} ErrorResponse "Não autenticado"
// @Failure 404 {object} ErrorResponse "Perfil de barbeiro não encontrado para o usuário"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /barber/my-queue [get]
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
// @Summary Concluir atendimento na cadeira
// @Description Finaliza a execução do serviço pelo barbeiro responsável, liberando a cadeira.
// @Tags Barber
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID do agendamento"
// @Success 200 {object} map[string]string "Mensagem de confirmação"
// @Failure 400 {object} ErrorResponse "ID inválido ou transição de status não permitida"
// @Failure 401 {object} ErrorResponse "Não autenticado"
// @Failure 403 {object} ErrorResponse "Atendimento não pertence a este barbeiro"
// @Failure 404 {object} ErrorResponse "Agendamento não encontrado"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /barber/appointments/{id}/complete [post]
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
