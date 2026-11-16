package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/i18n"
	"github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

// ManagerCatalogManager defines the contract for catalog and inventory operations for managers.
type ManagerCatalogManager interface {
	ListServices(ctx context.Context, unitID string, category *domain.ServiceCategory, status *string, claims *service.UserClaims) ([]domain.Service, error)
	GetService(ctx context.Context, unitID, serviceID string, claims *service.UserClaims) (*domain.Service, error)
	CreateService(ctx context.Context, unitID string, input domain.CreateServiceInput, claims *service.UserClaims) (*domain.Service, error)
	UpdateService(ctx context.Context, unitID, serviceID string, input domain.UpdateServiceInput, claims *service.UserClaims) (*domain.Service, error)
	DeleteService(ctx context.Context, unitID, serviceID string, claims *service.UserClaims) (softDeleted bool, err error)
	ListProducts(ctx context.Context, unitID string, category *domain.ProductCategory, status *string, claims *service.UserClaims) ([]domain.Product, error)
	GetProduct(ctx context.Context, unitID, productID string, claims *service.UserClaims) (*domain.Product, error)
	CreateProduct(ctx context.Context, unitID string, input domain.CreateProductInput, claims *service.UserClaims) (*domain.Product, error)
	UpdateProduct(ctx context.Context, unitID, productID string, input domain.UpdateProductInput, claims *service.UserClaims) (*domain.Product, error)
	DeleteProduct(ctx context.Context, unitID, productID string, claims *service.UserClaims) (softDeleted bool, err error)
	AdjustStock(ctx context.Context, unitID, productID string, input domain.AdjustProductStockInput, claims *service.UserClaims) (*domain.Product, *domain.InventoryMovement, error)
	ListProductMovements(ctx context.Context, unitID, productID string, limit int, claims *service.UserClaims) ([]domain.InventoryMovement, error)
}

// AdjustStockResponse wraps product and movement response payload.
type AdjustStockResponse struct {
	Product  *domain.Product           `json:"product"`
	Movement *domain.InventoryMovement `json:"movement"`
}

// ManagerCatalogHandler handles HTTP endpoints for manager catalog and stock management.
type ManagerCatalogHandler struct {
	svc ManagerCatalogManager
}

// NewManagerCatalogHandler instantiates a ManagerCatalogHandler.
func NewManagerCatalogHandler(svc ManagerCatalogManager) *ManagerCatalogHandler {
	return &ManagerCatalogHandler{svc: svc}
}

func resolveManagerUnitID(r *http.Request, claims *service.UserClaims) string {
	unitID := r.URL.Query().Get("unit_id")
	if unitID != "" {
		return unitID
	}
	if claims != nil && claims.UnitID != nil {
		return *claims.UnitID
	}
	return ""
}

// ListServices godoc
// @Summary      List services for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Param        unit_id   query     string  false  "Unit UUID (defaults to manager's unit)"
// @Param        category  query     string  false  "Service category filter"
// @Param        status    query     string  false  "Service status filter (active, inactive, all)"
// @Success      200       {array}   domain.Service
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/services [get]
func (h *ManagerCatalogHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var catPtr *domain.ServiceCategory
	if c := r.URL.Query().Get("category"); c != "" {
		cat := domain.ServiceCategory(c)
		catPtr = &cat
	}

	var statusPtr *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusPtr = &s
	}

	services, err := h.svc.ListServices(r.Context(), unitID, catPtr, statusPtr, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if services == nil {
		services = []domain.Service{}
	}
	_ = WriteJSON(w, http.StatusOK, services)
}

// GetService godoc
// @Summary      Get branch service by ID
// @Tags         manager
// @Security     BearerAuth
// @Param        id       path      string  true   "Service UUID"
// @Param        unit_id  query     string  false  "Unit UUID"
// @Success      200      {object}  domain.Service
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/services/{id} [get]
func (h *ManagerCatalogHandler) GetService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	serviceID := chi.URLParam(r, "id")
	svc, err := h.svc.GetService(r.Context(), unitID, serviceID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, svc)
}

// CreateService godoc
// @Summary      Create a new service for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        unit_id  query     string                     false  "Unit UUID"
// @Param        request  body      domain.CreateServiceInput  true   "Service creation payload"
// @Success      201      {object}  domain.Service
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/services [post]
func (h *ManagerCatalogHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var input domain.CreateServiceInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	svc, err := h.svc.CreateService(r.Context(), unitID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteCreated(w, svc)
}

// UpdateService godoc
// @Summary      Update branch service details
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true   "Service UUID"
// @Param        unit_id  query     string                     false  "Unit UUID"
// @Param        request  body      domain.UpdateServiceInput  true   "Service update payload"
// @Success      200      {object}  domain.Service
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/services/{id} [put]
func (h *ManagerCatalogHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	serviceID := chi.URLParam(r, "id")
	var input domain.UpdateServiceInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	svc, err := h.svc.UpdateService(r.Context(), unitID, serviceID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, svc)
}

// DeleteService godoc
// @Summary      Delete or soft-deactivate a branch service
// @Tags         manager
// @Security     BearerAuth
// @Param        id       path      string  true   "Service UUID"
// @Param        unit_id  query     string  false  "Unit UUID"
// @Success      200      {object}  map[string]any  "Soft-deleted service"
// @Success      204      "Hard-deleted service"
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/services/{id} [delete]
func (h *ManagerCatalogHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	serviceID := chi.URLParam(r, "id")
	softDeleted, err := h.svc.DeleteService(r.Context(), unitID, serviceID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if softDeleted {
		_ = WriteJSON(w, http.StatusOK, map[string]any{
			"soft_deleted": true,
			"message":      "serviço inativado devido a agendamentos vinculados",
		})
		return
	}

	WriteNoContent(w)
}

// ListProducts godoc
// @Summary      List retail products and stock for branch unit
// @Tags         manager
// @Security     BearerAuth
// @Param        unit_id   query     string  false  "Unit UUID"
// @Param        category  query     string  false  "Product category filter"
// @Param        status    query     string  false  "Product status filter (active, inactive, low_stock)"
// @Success      200       {array}   domain.Product
// @Failure      400       {object}  ErrorResponse
// @Failure      401       {object}  ErrorResponse
// @Failure      403       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /manager/products [get]
func (h *ManagerCatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var catPtr *domain.ProductCategory
	if c := r.URL.Query().Get("category"); c != "" {
		cat := domain.ProductCategory(c)
		catPtr = &cat
	}

	var statusPtr *string
	if s := r.URL.Query().Get("status"); s != "" {
		statusPtr = &s
	}

	products, err := h.svc.ListProducts(r.Context(), unitID, catPtr, statusPtr, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if products == nil {
		products = []domain.Product{}
	}
	_ = WriteJSON(w, http.StatusOK, products)
}

// GetProduct godoc
// @Summary      Get branch product by ID
// @Tags         manager
// @Security     BearerAuth
// @Param        id       path      string  true   "Product UUID"
// @Param        unit_id  query     string  false  "Unit UUID"
// @Success      200      {object}  domain.Product
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products/{id} [get]
func (h *ManagerCatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	productID := chi.URLParam(r, "id")
	prod, err := h.svc.GetProduct(r.Context(), unitID, productID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, prod)
}

// CreateProduct godoc
// @Summary      Create a retail product in branch unit
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        unit_id  query     string                     false  "Unit UUID"
// @Param        request  body      domain.CreateProductInput  true   "Product payload"
// @Success      201      {object}  domain.Product
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products [post]
func (h *ManagerCatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	var input domain.CreateProductInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	prod, err := h.svc.CreateProduct(r.Context(), unitID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteCreated(w, prod)
}

// UpdateProduct godoc
// @Summary      Update branch product details
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                     true   "Product UUID"
// @Param        unit_id  query     string                     false  "Unit UUID"
// @Param        request  body      domain.UpdateProductInput  true   "Product update payload"
// @Success      200      {object}  domain.Product
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products/{id} [put]
func (h *ManagerCatalogHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	productID := chi.URLParam(r, "id")
	var input domain.UpdateProductInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	prod, err := h.svc.UpdateProduct(r.Context(), unitID, productID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, prod)
}

// DeleteProduct godoc
// @Summary      Delete or soft-deactivate a branch product
// @Tags         manager
// @Security     BearerAuth
// @Param        id       path      string  true   "Product UUID"
// @Param        unit_id  query     string  false  "Unit UUID"
// @Success      200      {object}  map[string]any  "Soft-deleted product"
// @Success      204      "Hard-deleted product"
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products/{id} [delete]
func (h *ManagerCatalogHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	productID := chi.URLParam(r, "id")
	softDeleted, err := h.svc.DeleteProduct(r.Context(), unitID, productID, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if softDeleted {
		_ = WriteJSON(w, http.StatusOK, map[string]any{
			"soft_deleted": true,
			"message":      "produto inativado devido a histórico de movimentações",
		})
		return
	}

	WriteNoContent(w)
}

// AdjustStock godoc
// @Summary      Adjust product inventory balance with audit movement
// @Tags         manager
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      string                           true   "Product UUID"
// @Param        unit_id  query     string                           false  "Unit UUID"
// @Param        request  body      domain.AdjustProductStockInput   true   "Stock adjustment payload"
// @Success      200      {object}  AdjustStockResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products/{id}/stock [patch]
func (h *ManagerCatalogHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	productID := chi.URLParam(r, "id")
	var input domain.AdjustProductStockInput
	if err := ReadJSON(w, r, &input); err != nil {
		WriteErrorWithLocale(w, http.StatusBadRequest, err.Error(), i18n.ResolveLocale(r), domain.ErrCodeValidationFailed)
		return
	}

	prod, mov, err := h.svc.AdjustStock(r.Context(), unitID, productID, input, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	_ = WriteJSON(w, http.StatusOK, AdjustStockResponse{
		Product:  prod,
		Movement: mov,
	})
}

// ListProductMovements godoc
// @Summary      List inventory movements history for a product
// @Tags         manager
// @Security     BearerAuth
// @Param        id       path      string  true   "Product UUID"
// @Param        unit_id  query     string  false  "Unit UUID"
// @Param        limit    query     int     false  "Maximum records to return (default 50)"
// @Success      200      {array}   domain.InventoryMovement
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /manager/products/{id}/movements [get]
func (h *ManagerCatalogHandler) ListProductMovements(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok || claims == nil {
		HandleError(w, domain.ErrUnauthorized, r)
		return
	}

	unitID := resolveManagerUnitID(r, claims)
	if unitID == "" {
		WriteError(w, http.StatusBadRequest, "unit_id parameter is required")
		return
	}

	productID := chi.URLParam(r, "id")
	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	movements, err := h.svc.ListProductMovements(r.Context(), unitID, productID, limit, claims)
	if err != nil {
		HandleError(w, err, r)
		return
	}

	if movements == nil {
		movements = []domain.InventoryMovement{}
	}
	_ = WriteJSON(w, http.StatusOK, movements)
}
