package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	appmw "github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

// B2BHandler handles HTTP requests for the B2B Comptoir portal.
type B2BHandler struct {
	svc        domain.B2BService
	bakeryRepo domain.BakeryRepository
}

// NewB2BHandler creates a new B2BHandler.
func NewB2BHandler(svc domain.B2BService, bakeryRepo domain.BakeryRepository) *B2BHandler {
	return &B2BHandler{svc: svc, bakeryRepo: bakeryRepo}
}

// requireB2BRole is middleware that enforces RoleBusiness (role 3).
func requireB2BRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := appmw.GetUserRoleFromContext(r.Context())
		if role != int(domain.RoleBusiness) {
			writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
				Code:    "FORBIDDEN",
				Message: "business role required",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireSellerRole is middleware that enforces RoleSeller or RoleAdmin.
func requireSellerRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := appmw.GetUserRoleFromContext(r.Context())
		if role != int(domain.RoleSeller) && role != int(domain.RoleAdmin) {
			writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
				Code:    "FORBIDDEN",
				Message: "seller or admin role required",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes mounts B2B routes on the chi router.
func (h *B2BHandler) RegisterRoutes(r chi.Router, jwtSecret string, userRepo domain.UserRepository) {
	// Public: B2B registration (no auth required)
	r.Post("/api/comptoir/register", h.Register)

	// Protected: requires JWT + RoleBusiness (role 3)
	r.Route("/api/comptoir", func(r chi.Router) {
		r.Use(appmw.JWTAuth(jwtSecret, userRepo))
		r.Use(requireB2BRole)

		// Profile
		r.Get("/profile", h.GetProfile)
		r.Put("/profile", h.UpdateProfile)

		// Delivery Sites
		r.Post("/sites", h.CreateSite)
		r.Get("/sites", h.ListSites)
		r.Put("/sites/{siteId}", h.UpdateSite)
		r.Delete("/sites/{siteId}", h.DeleteSite)

		// Access Requests
		r.Post("/access/request/{bakeryId}", h.RequestAccess)
		r.Get("/bakeries", h.ListApprovedBakeries)

		// Product Catalog (access checked in handler)
		r.Get("/bakeries/{bakeryId}/products", h.GetProducts)

		// B2B Config (read-only for business users)
		r.Get("/bakeries/{bakeryId}/config", h.GetBakeryConfig)

		// Cart & Checkout
		r.Post("/checkout", h.Checkout)
		r.Put("/orders/{orderId}", h.EditOrder)
		r.Post("/pricing", h.ComputePricing)

		// Saved Lists
		r.Post("/lists", h.CreateSavedList)
		r.Get("/lists", h.ListSavedLists)
		r.Delete("/lists/{listId}", h.DeleteSavedList)

		// Deliveries
		r.Get("/deliveries", h.ListDeliveries)
		r.Get("/orders/{bakeryId}/last", h.GetLastOrder)

		// Invoices
		r.Get("/invoices", h.ListInvoices)
		r.Get("/invoices/{invoiceId}/pdf", h.DownloadInvoicePDF)
	})

	// Baker-facing B2B management (requires JWT + seller/admin role)
	r.Route("/api/dashboard/b2b", func(r chi.Router) {
		r.Use(appmw.JWTAuth(jwtSecret, userRepo))
		r.Use(requireSellerRole)

		r.Get("/access", h.ListAccessRequests)
		r.Post("/access/{accessId}/approve", h.ApproveAccess)
		r.Post("/access/{accessId}/reject", h.RejectAccess)
		r.Post("/access/{accessId}/revoke", h.RevokeAccess)

		r.Get("/config", h.GetSellerConfig)
		r.Put("/config", h.SaveConfig)
	})
}

// --- Registration ---

// Register handles POST /api/comptoir/register.
func (h *B2BHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	domainReq := domain.RegisterBusinessRequest{
		Username:           req.Username,
		Password:           req.Password,
		CompanyName:        req.CompanyName,
		VATSiret:           req.VATSiret,
		IBAN:               req.IBAN,
		BillingEmail:       req.BillingEmail,
		BillingContactName: req.BillingContactName,
	}

	profile, token, err := h.svc.RegisterBusiness(r.Context(), domainReq)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.RegisterBusinessResponse{
		Token:   token,
		Profile: toBusinessProfileResponse(profile),
	})
}

// --- Profile ---

// GetProfile handles GET /api/comptoir/profile.
func (h *B2BHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	profile, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBusinessProfileResponse(profile))
}

// UpdateProfile handles PUT /api/comptoir/profile.
func (h *B2BHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	domainReq := domain.UpdateProfileRequest{
		CompanyName:        req.CompanyName,
		IBAN:               req.IBAN,
		BillingEmail:       req.BillingEmail,
		BillingContactName: req.BillingContactName,
	}

	profile, err := h.svc.UpdateProfile(r.Context(), userID, domainReq)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toBusinessProfileResponse(profile))
}

// --- Delivery Sites ---

// CreateSite handles POST /api/comptoir/sites.
func (h *B2BHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	var req dto.DeliverySiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	site := domain.DeliverySite{
		Name:                 req.Name,
		StreetAddress:        req.StreetAddress,
		City:                 req.City,
		PostalCode:           req.PostalCode,
		Country:              req.Country,
		DeliveryInstructions: req.DeliveryInstructions,
	}

	created, err := h.svc.CreateSite(r.Context(), userID, site)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDeliverySiteResponse(created))
}

// ListSites handles GET /api/comptoir/sites.
func (h *B2BHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	sites, err := h.svc.ListSites(r.Context(), userID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	resp := make([]dto.DeliverySiteResponse, len(sites))
	for i, s := range sites {
		resp[i] = toDeliverySiteResponse(&s)
	}
	writeJSON(w, http.StatusOK, resp)
}

// UpdateSite handles PUT /api/comptoir/sites/{siteId}.
func (h *B2BHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	siteID := chi.URLParam(r, "siteId")

	var req dto.DeliverySiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	site := domain.DeliverySite{
		Name:                 req.Name,
		StreetAddress:        req.StreetAddress,
		City:                 req.City,
		PostalCode:           req.PostalCode,
		Country:              req.Country,
		DeliveryInstructions: req.DeliveryInstructions,
	}

	updated, err := h.svc.UpdateSite(r.Context(), userID, siteID, site)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDeliverySiteResponse(updated))
}

// DeleteSite handles DELETE /api/comptoir/sites/{siteId}.
func (h *B2BHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	siteID := chi.URLParam(r, "siteId")

	if err := h.svc.DeleteSite(r.Context(), userID, siteID); err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "delivery site deleted",
	})
}

// --- Access Management ---

// RequestAccess handles POST /api/comptoir/access/request/{bakeryId}.
func (h *B2BHandler) RequestAccess(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeryID := chi.URLParam(r, "bakeryId")

	access, err := h.svc.RequestAccess(r.Context(), userID, bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toB2BAccessResponse(access))
}

// ListApprovedBakeries handles GET /api/comptoir/bakeries.
func (h *B2BHandler) ListApprovedBakeries(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeries, err := h.svc.ListApprovedBakeries(r.Context(), userID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bakeries)
}

// GetProducts handles GET /api/comptoir/bakeries/{bakeryId}/products.
func (h *B2BHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeryID := chi.URLParam(r, "bakeryId")

	// Check access
	hasAccess, err := h.svc.HasApprovedAccess(r.Context(), userID, bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	if !hasAccess {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "ACCESS_DENIED",
			Message: "B2B access not approved for this bakery",
		})
		return
	}

	products, err := h.bakeryRepo.GetProductsByBakery(r.Context(), bakeryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch products",
		})
		return
	}

	// Group products by category
	grouped := make(map[string][]domain.Product)
	for _, p := range products {
		grouped[p.Category] = append(grouped[p.Category], p)
	}
	writeJSON(w, http.StatusOK, grouped)
}

// GetBakeryConfig handles GET /api/comptoir/bakeries/{bakeryId}/config.
func (h *B2BHandler) GetBakeryConfig(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeryID := chi.URLParam(r, "bakeryId")

	// Check access
	hasAccess, err := h.svc.HasApprovedAccess(r.Context(), userID, bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	if !hasAccess {
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "ACCESS_DENIED",
			Message: "B2B access not approved for this bakery",
		})
		return
	}

	config, err := h.svc.GetConfig(r.Context(), bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toB2BConfigResponse(config))
}

// --- Checkout & Orders ---

// Checkout handles POST /api/comptoir/checkout.
func (h *B2BHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	var req dto.B2BCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Enrich items with product prices
	products, err := h.bakeryRepo.GetProductsByBakery(r.Context(), req.BakeryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch products",
		})
		return
	}
	productMap := make(map[string]domain.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	for i := range items {
		if p, ok := productMap[items[i].ProductID]; ok {
			items[i].ProductName = p.Name
			items[i].UnitPrice = p.Price
			items[i].Subtotal = int64(items[i].Quantity) * p.Price
		}
	}

	checkoutReq := domain.CheckoutRequest{
		BakeryID:       req.BakeryID,
		DeliverySiteID: req.DeliverySiteID,
		Items:          items,
	}

	order, err := h.svc.CheckoutBakeryGroup(r.Context(), userID, checkoutReq)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

// EditOrder handles PUT /api/comptoir/orders/{orderId}.
func (h *B2BHandler) EditOrder(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	orderID := chi.URLParam(r, "orderId")

	var req dto.B2BEditOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	editReq := domain.EditOrderRequest{Items: items}
	order, err := h.svc.EditOrder(r.Context(), userID, orderID, editReq)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// ComputePricing handles POST /api/comptoir/pricing.
func (h *B2BHandler) ComputePricing(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	var req dto.B2BPricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Enrich items with prices
	products, err := h.bakeryRepo.GetProductsByBakery(r.Context(), req.BakeryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to fetch products",
		})
		return
	}
	productMap := make(map[string]domain.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}
	for i := range items {
		if p, ok := productMap[items[i].ProductID]; ok {
			items[i].UnitPrice = p.Price
		}
	}

	pricing, err := h.svc.ComputePricing(r.Context(), userID, req.BakeryID, items)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	resp := dto.B2BPricingResultResponse{
		SubtotalHT:      pricing.SubtotalHT,
		ProDiscountRate: pricing.ProDiscountRate,
		ProDiscountAmt:  pricing.ProDiscountAmt,
		VolDiscountRate: pricing.VolDiscountRate,
		VolDiscountAmt:  pricing.VolDiscountAmt,
		TVARate:         pricing.TVARate,
		TVAAmount:       pricing.TVAAmount,
		TotalTTC:        pricing.TotalTTC,
		MonthlySpend:    pricing.MonthlySpend,
		SpendToNextTier: pricing.SpendToNextTier,
	}
	if pricing.CurrentTier != nil {
		resp.CurrentTier = &dto.VolumeTierResponse{
			MinMonthlySpend: pricing.CurrentTier.MinMonthlySpend,
			DiscountPercent: pricing.CurrentTier.DiscountPercent,
		}
	}
	if pricing.NextTier != nil {
		resp.NextTier = &dto.VolumeTierResponse{
			MinMonthlySpend: pricing.NextTier.MinMonthlySpend,
			DiscountPercent: pricing.NextTier.DiscountPercent,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// --- Saved Lists ---

// CreateSavedList handles POST /api/comptoir/lists.
func (h *B2BHandler) CreateSavedList(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	var req dto.SavedListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	items := make([]domain.SavedListItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.SavedListItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	list := domain.SavedList{
		BakeryID: req.BakeryID,
		Name:     req.Name,
		Items:    items,
	}

	created, err := h.svc.CreateSavedList(r.Context(), userID, list)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSavedListResponse(created))
}

// ListSavedLists handles GET /api/comptoir/lists.
func (h *B2BHandler) ListSavedLists(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeryID := r.URL.Query().Get("bakeryId")

	lists, err := h.svc.ListSavedLists(r.Context(), userID, bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	resp := make([]dto.SavedListResponse, len(lists))
	for i, l := range lists {
		resp[i] = toSavedListResponse(&l)
	}
	writeJSON(w, http.StatusOK, resp)
}

// DeleteSavedList handles DELETE /api/comptoir/lists/{listId}.
func (h *B2BHandler) DeleteSavedList(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	listID := chi.URLParam(r, "listId")

	if err := h.svc.DeleteSavedList(r.Context(), userID, listID); err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "saved list deleted",
	})
}

// --- Deliveries ---

// ListDeliveries handles GET /api/comptoir/deliveries.
func (h *B2BHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 1 {
			page = parsed
		}
	}

	filters := domain.B2BOrderFilters{
		BakeryID: r.URL.Query().Get("bakeryId"),
		Status:   r.URL.Query().Get("status"),
	}

	params := domain.PaginationParams{Page: page, PageSize: 20}
	result, err := h.svc.ListDeliveries(r.Context(), userID, filters, params)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetLastOrder handles GET /api/comptoir/orders/{bakeryId}/last.
func (h *B2BHandler) GetLastOrder(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	bakeryID := chi.URLParam(r, "bakeryId")

	order, err := h.svc.GetLastOrder(r.Context(), userID, bakeryID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	if order == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// --- Invoices ---

// ListInvoices handles GET /api/comptoir/invoices.
func (h *B2BHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 1 {
			page = parsed
		}
	}

	params := domain.PaginationParams{Page: page, PageSize: 20}
	result, err := h.svc.ListInvoices(r.Context(), userID, params)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// DownloadInvoicePDF handles GET /api/comptoir/invoices/{invoiceId}/pdf.
func (h *B2BHandler) DownloadInvoicePDF(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	invoiceID := chi.URLParam(r, "invoiceId")

	pdf, err := h.svc.DownloadInvoicePDF(r.Context(), invoiceID, userID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=invoice.pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdf)
}

// --- Baker-facing B2B management ---

// ListAccessRequests handles GET /api/dashboard/b2b/access.
func (h *B2BHandler) ListAccessRequests(w http.ResponseWriter, r *http.Request) {
	sellerID := extractUserID(r)

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), sellerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to look up bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "no bakery found for this seller",
		})
		return
	}

	accesses, err := h.svc.ListAccessRequests(r.Context(), bakery.ID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}

	resp := make([]dto.B2BAccessResponse, len(accesses))
	for i, a := range accesses {
		resp[i] = toB2BAccessResponse(&a)
	}
	writeJSON(w, http.StatusOK, resp)
}

// ApproveAccess handles POST /api/dashboard/b2b/access/{accessId}/approve.
func (h *B2BHandler) ApproveAccess(w http.ResponseWriter, r *http.Request) {
	accessID := chi.URLParam(r, "accessId")
	if err := h.svc.ApproveAccess(r.Context(), accessID); err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "access approved"})
}

// RejectAccess handles POST /api/dashboard/b2b/access/{accessId}/reject.
func (h *B2BHandler) RejectAccess(w http.ResponseWriter, r *http.Request) {
	accessID := chi.URLParam(r, "accessId")
	if err := h.svc.RejectAccess(r.Context(), accessID); err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "access rejected"})
}

// RevokeAccess handles POST /api/dashboard/b2b/access/{accessId}/revoke.
func (h *B2BHandler) RevokeAccess(w http.ResponseWriter, r *http.Request) {
	accessID := chi.URLParam(r, "accessId")
	if err := h.svc.RevokeAccess(r.Context(), accessID); err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "access revoked"})
}

// GetSellerConfig handles GET /api/dashboard/b2b/config.
func (h *B2BHandler) GetSellerConfig(w http.ResponseWriter, r *http.Request) {
	sellerID := extractUserID(r)

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), sellerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to look up bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "no bakery found for this seller",
		})
		return
	}

	config, err := h.svc.GetConfig(r.Context(), bakery.ID)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toB2BConfigResponse(config))
}

// SaveConfig handles PUT /api/dashboard/b2b/config.
func (h *B2BHandler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	sellerID := extractUserID(r)

	bakery, err := h.bakeryRepo.GetBakeryByOwner(r.Context(), sellerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to look up bakery",
		})
		return
	}
	if bakery == nil {
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "BAKERY_NOT_FOUND",
			Message: "no bakery found for this seller",
		})
		return
	}

	var req dto.B2BConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid request body",
		})
		return
	}

	cutoff, err := parseTimeOfDayStr(req.CutoffTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid cutoffTime: expected HH:MM format",
		})
		return
	}
	windowStart, err := parseTimeOfDayStr(req.DeliveryWindowStart)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid deliveryWindowStart: expected HH:MM format",
		})
		return
	}
	windowEnd, err := parseTimeOfDayStr(req.DeliveryWindowEnd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid deliveryWindowEnd: expected HH:MM format",
		})
		return
	}

	config := domain.B2BConfig{
		CutoffTime:          cutoff,
		DeliveryWindowStart: windowStart,
		DeliveryWindowEnd:   windowEnd,
		OrderMinimum:        req.OrderMinimum,
		ProDiscount:         req.ProDiscount,
		VATRate:             req.VATRate,
	}

	saved, err := h.svc.SaveConfig(r.Context(), bakery.ID, config)
	if err != nil {
		h.handleB2BError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toB2BConfigResponse(saved))
}

// --- Error handling ---

func (h *B2BHandler) handleB2BError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCompanyAlreadyExists):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "COMPANY_ALREADY_EXISTS",
			Message: "company is already registered",
		})
	case errors.Is(err, service.ErrAccessDenied):
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "ACCESS_DENIED",
			Message: "B2B access not approved for this bakery",
		})
	case errors.Is(err, service.ErrBelowMinimum):
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "BELOW_MINIMUM",
			Message: "order total is below the bakery minimum",
		})
	case errors.Is(err, service.ErrCutoffPassed):
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "CUTOFF_PASSED",
			Message: "cutoff time has passed for this bakery",
		})
	case errors.Is(err, service.ErrLastSite):
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "LAST_SITE",
			Message: "at least one delivery site is required",
		})
	case errors.Is(err, service.ErrNoDeliverySite):
		writeJSON(w, http.StatusUnprocessableEntity, dto.ErrorResponse{
			Code:    "NO_DELIVERY_SITE",
			Message: "a delivery site is required to place an order",
		})
	case errors.Is(err, service.ErrAccessExists):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "ACCESS_EXISTS",
			Message: "access request already exists for this bakery",
		})
	case errors.Is(err, service.ErrProfileNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "PROFILE_NOT_FOUND",
			Message: "business profile not found",
		})
	case errors.Is(err, service.ErrSiteNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "SITE_NOT_FOUND",
			Message: "delivery site not found",
		})
	case errors.Is(err, service.ErrInvoiceNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "INVOICE_NOT_FOUND",
			Message: "invoice not found",
		})
	case errors.Is(err, service.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
			Code:    "ORDER_NOT_FOUND",
			Message: "order not found",
		})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: "you do not own this resource",
		})
	case errors.Is(err, service.ErrUsernameExists):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{
			Code:    "USERNAME_EXISTS",
			Message: "username already exists",
		})
	default:
		// Check if it's a validation error (from domain.ValidateBusinessRegistration)
		if err != nil && isB2BValidationErr(err) {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an internal error occurred",
		})
	}
}

// isB2BValidationErr checks if an error is a validation error from
// domain.ValidateBusinessRegistration (returns semicolon-separated messages).
func isB2BValidationErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "is required") || strings.Contains(msg, "must") ||
		strings.Contains(msg, "not exceed") || strings.Contains(msg, "valid email")
}

// --- DTO conversion helpers ---

func toBusinessProfileResponse(p *domain.BusinessProfile) dto.BusinessProfileResponse {
	return dto.BusinessProfileResponse{
		ID:                 p.ID,
		UserID:             p.UserID,
		CompanyName:        p.CompanyName,
		VATSiret:           p.VATSiret,
		IBAN:               p.IBAN,
		BillingEmail:       p.BillingEmail,
		BillingContactName: p.BillingContactName,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func toDeliverySiteResponse(s *domain.DeliverySite) dto.DeliverySiteResponse {
	return dto.DeliverySiteResponse{
		ID:                   s.ID,
		UserID:               s.UserID,
		Name:                 s.Name,
		StreetAddress:        s.StreetAddress,
		City:                 s.City,
		PostalCode:           s.PostalCode,
		Country:              s.Country,
		DeliveryInstructions: s.DeliveryInstructions,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
}

func toB2BAccessResponse(a *domain.B2BAccess) dto.B2BAccessResponse {
	return dto.B2BAccessResponse{
		ID:             a.ID,
		BakeryID:       a.BakeryID,
		BusinessUserID: a.BusinessUserID,
		Status:         string(a.Status),
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func toB2BConfigResponse(c *domain.B2BConfig) dto.B2BConfigResponse {
	return dto.B2BConfigResponse{
		ID:                  c.ID,
		BakeryID:            c.BakeryID,
		CutoffTime:          c.CutoffTime.String(),
		DeliveryWindowStart: c.DeliveryWindowStart.String(),
		DeliveryWindowEnd:   c.DeliveryWindowEnd.String(),
		OrderMinimum:        c.OrderMinimum,
		ProDiscount:         c.ProDiscount,
		VATRate:             c.VATRate,
	}
}

func toSavedListResponse(l *domain.SavedList) dto.SavedListResponse {
	items := make([]dto.SavedListItemResponse, len(l.Items))
	for i, item := range l.Items {
		items[i] = dto.SavedListItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return dto.SavedListResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		BakeryID:  l.BakeryID,
		Name:      l.Name,
		Items:     items,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
