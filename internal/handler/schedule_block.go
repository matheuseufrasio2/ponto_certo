package handler

import (
	"context"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// ScheduleBlockManager defines operations for barber agenda blockage.
type ScheduleBlockManager interface {
	CreateBlock(ctx context.Context, barberID string, input domain.CreateScheduleBlockInput, claims *service.UserClaims) (*domain.BarberScheduleBlock, error)
	ListBlocks(ctx context.Context, barberID string, from, to time.Time, claims *service.UserClaims) ([]domain.BarberScheduleBlock, error)
	DeleteBlock(ctx context.Context, blockID string, claims *service.UserClaims) error
}

// ScheduleBlockHandler handles HTTP requests for managing barber agenda blockages.
type ScheduleBlockHandler struct {
	svc ScheduleBlockManager
}

// NewScheduleBlockHandler instantiates ScheduleBlockHandler.
func NewScheduleBlockHandler(svc ScheduleBlockManager) *ScheduleBlockHandler {
	return &ScheduleBlockHandler{svc: svc}
}

// CreateBlock godoc
// @Summary Create barber schedule block
// @Description Blocks a time interval on a barber's schedule (e.g. lunch, medical appointment, vacation)
// @Tags manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Barber UUID"
// @Param body body domain.CreateScheduleBlockInput true "Block details"
// @Success 201 {object} domain.BarberScheduleBlock
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /manager/barbers/{id}/blocks [post]
func (h *ScheduleBlockHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	barberID := chi.URLParam(r, "id")
	if barberID == "" || !domain.IsValidUUID(barberID) {
		WriteError(w, http.StatusBadRequest, "invalid barber id format, expected uuid")
		return
	}

	var req domain.CreateScheduleBlockInput
	if err := ReadJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	block, err := h.svc.CreateBlock(r.Context(), barberID, req, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteCreated(w, block)
}

// ListBlocks godoc
// @Summary List barber schedule blocks
// @Description Retrieves all agenda blockages for a barber, optionally filtered by date range
// @Tags manager
// @Produce json
// @Security BearerAuth
// @Param id path string true "Barber UUID"
// @Param from query string false "Filter start (RFC3339 format)"
// @Param to query string false "Filter end (RFC3339 format)"
// @Success 200 {array} domain.BarberScheduleBlock
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /manager/barbers/{id}/blocks [get]
func (h *ScheduleBlockHandler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	barberID := chi.URLParam(r, "id")
	if barberID == "" || !domain.IsValidUUID(barberID) {
		WriteError(w, http.StatusBadRequest, "invalid barber id format, expected uuid")
		return
	}

	var from, to time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	blocks, err := h.svc.ListBlocks(r.Context(), barberID, from, to, claims)
	if err != nil {
		HandleError(w, err)
		return
	}

	if blocks == nil {
		blocks = []domain.BarberScheduleBlock{}
	}
	_ = WriteJSON(w, http.StatusOK, blocks)
}

// DeleteBlock godoc
// @Summary Delete barber schedule block
// @Description Removes an agenda blockage, freeing up time slots for appointment scheduling
// @Tags manager
// @Security BearerAuth
// @Param block_id path string true "Schedule Block UUID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /manager/barbers/blocks/{block_id} [delete]
func (h *ScheduleBlockHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	blockID := chi.URLParam(r, "block_id")
	if blockID == "" || !domain.IsValidUUID(blockID) {
		WriteError(w, http.StatusBadRequest, "invalid block id format, expected uuid")
		return
	}

	if err := h.svc.DeleteBlock(r.Context(), blockID, claims); err != nil {
		HandleError(w, err)
		return
	}

	WriteNoContent(w)
}
