package handler

import (
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// BillingHandler manages HTTP endpoints for billing and checkout.
type BillingHandler struct {
	billingService *service.BillingService
}

// NewBillingHandler creates a new BillingHandler instance.
func NewBillingHandler(billingService *service.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

// Checkout handles POST /api/v1/reception/checkout.
func (h *BillingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized)
		return
	}

	var input domain.CheckoutInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	billing, err := h.billingService.Checkout(r.Context(), input, claims.UserID)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, billing)
}
