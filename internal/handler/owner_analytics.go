package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// OwnerAnalyticsManager defines business operations for executive network analytics.
type OwnerAnalyticsManager interface {
	GetExecutiveKPIs(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) (*domain.OwnerExecutiveKPIs, error)
	GetRevenueComparisonByUnit(ctx context.Context, claims *service.UserClaims, start, end time.Time) ([]domain.UnitRevenueComparison, error)
	GetServicesMix(ctx context.Context, claims *service.UserClaims, unitID *string, start, end time.Time) ([]domain.ServiceMixAnalytics, error)
	ListConsolidatedBillings(ctx context.Context, claims *service.UserClaims, unitID *string, paymentMethod *string, page, perPage int) (*domain.PaginatedBillings, error)
	ExportBillingsCSV(ctx context.Context, claims *service.UserClaims, unitID *string, w io.Writer) error
}

// OwnerAnalyticsHandler manages HTTP endpoints for the organization owner BI.
type OwnerAnalyticsHandler struct {
	svc OwnerAnalyticsManager
}

// NewOwnerAnalyticsHandler instantiates a new OwnerAnalyticsHandler.
func NewOwnerAnalyticsHandler(svc OwnerAnalyticsManager) *OwnerAnalyticsHandler {
	return &OwnerAnalyticsHandler{svc: svc}
}

func parseDateRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now().UTC()
	start := now.AddDate(0, 0, -30)
	end := now

	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	} else if s := r.URL.Query().Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}

	if e := r.URL.Query().Get("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	} else if e := r.URL.Query().Get("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t
		}
	}

	return start, end
}

func parseUnitID(r *http.Request) *string {
	u := r.URL.Query().Get("unit_id")
	if u == "" || u == "ALL" {
		return nil
	}
	return &u
}

// GetMetrics godoc
// @Summary Get executive network KPIs
// @Description Returns gross revenue, net revenue, total commissions, average ticket and retention rate
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param unit_id query string false "Filter by unit UUID or ALL"
// @Param start query string false "Start date YYYY-MM-DD"
// @Param end query string false "End date YYYY-MM-DD"
// @Success 200 {object} domain.OwnerExecutiveKPIs
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/metrics [get]
func (h *OwnerAnalyticsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	start, end := parseDateRange(r)
	unitID := parseUnitID(r)

	kpis, err := h.svc.GetExecutiveKPIs(r.Context(), claims, unitID, start, end)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, kpis)
}

// GetRevenueByUnit godoc
// @Summary Get revenue comparison across all units
// @Description Aggregates and ranks revenue by physical branch
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param start query string false "Start date YYYY-MM-DD"
// @Param end query string false "End date YYYY-MM-DD"
// @Success 200 {array} domain.UnitRevenueComparison
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/revenue-by-unit [get]
func (h *OwnerAnalyticsHandler) GetRevenueByUnit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	start, end := parseDateRange(r)
	comparisons, err := h.svc.GetRevenueComparisonByUnit(r.Context(), claims, start, end)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, comparisons)
}

// GetServicesMix godoc
// @Summary Get service mix analytics
// @Description Ranks performed services by quantity and revenue generated
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param unit_id query string false "Filter by unit UUID or ALL"
// @Param start query string false "Start date YYYY-MM-DD"
// @Param end query string false "End date YYYY-MM-DD"
// @Success 200 {array} domain.ServiceMixAnalytics
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/services-mix [get]
func (h *OwnerAnalyticsHandler) GetServicesMix(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	start, end := parseDateRange(r)
	unitID := parseUnitID(r)

	mix, err := h.svc.GetServicesMix(r.Context(), claims, unitID, start, end)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, mix)
}

// ListBillings godoc
// @Summary List consolidated billings
// @Description Returns paginated financial transactions with optional unit and payment method filters
// @Tags owner
// @Produce json
// @Security BearerAuth
// @Param unit_id query string false "Filter by unit UUID or ALL"
// @Param method query string false "Payment method (PIX, CREDIT_CARD, DEBIT_CARD, CASH)"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} domain.PaginatedBillings
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/billings [get]
func (h *OwnerAnalyticsHandler) ListBillings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := parseUnitID(r)
	var paymentMethod *string
	if m := r.URL.Query().Get("method"); m != "" && m != "ALL" {
		paymentMethod = &m
	} else if m := r.URL.Query().Get("payment_method"); m != "" && m != "ALL" {
		paymentMethod = &m
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	perPage := 20
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if val, err := strconv.Atoi(pp); err == nil && val > 0 {
			perPage = val
		}
	}

	res, err := h.svc.ListConsolidatedBillings(r.Context(), claims, unitID, paymentMethod, page, perPage)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, res)
}

// ExportCSV godoc
// @Summary Export consolidated billings as CSV
// @Description Streams a CSV file containing all billing records for the organization
// @Tags owner
// @Produce text/csv
// @Security BearerAuth
// @Param unit_id query string false "Filter by unit UUID or ALL"
// @Success 200 {string} string "CSV file download"
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/export/csv [get]
func (h *OwnerAnalyticsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	unitID := parseUnitID(r)

	filename := fmt.Sprintf("faturamentos_consolidados_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)

	if err := h.svc.ExportBillingsCSV(r.Context(), claims, unitID, w); err != nil {
		// Response status is already written, log error
		return
	}
}
