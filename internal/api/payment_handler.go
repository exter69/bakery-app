package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/payment"
)

const maxPaymentRetries = 3

// PaymentHandler handles HTTP requests related to payment callbacks.
type PaymentHandler struct {
	paymentSvc domain.PaymentService
	orderRepo  domain.OrderRepository

	mu      sync.Mutex
	retries map[string]int // orderID -> retry count
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentSvc domain.PaymentService, orderRepo domain.OrderRepository) *PaymentHandler {
	return &PaymentHandler{
		paymentSvc: paymentSvc,
		orderRepo:  orderRepo,
		retries:    make(map[string]int),
	}
}

// RegisterRoutes registers payment-related routes on the given chi router.
func (h *PaymentHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/payments/callback", h.HandleCallback)
}

// HandleCallback handles POST /api/payments/callback.
// It processes payment gateway webhook notifications for order payments.
func (h *PaymentHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	var req dto.PaymentCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "invalid JSON body",
		})
		return
	}

	if req.OrderID == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "orderId is required",
		})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "status is required",
		})
		return
	}

	if req.Status != "success" && req.Status != "failed" {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "status must be 'success' or 'failed'",
		})
		return
	}

	switch req.Status {
	case "success":
		h.handleSuccess(w, r, req)
	case "failed":
		h.handleFailure(w, r, req)
	}
}

// handleSuccess processes a successful payment callback.
func (h *PaymentHandler) handleSuccess(w http.ResponseWriter, r *http.Request, req dto.PaymentCallbackRequest) {
	err := h.paymentSvc.ProcessPaymentCallback(r.Context(), req.OrderID, req.PaymentRef)
	if err != nil {
		if errors.Is(err, payment.ErrLinkExpired) {
			// Payment link has expired — cancel the order
			h.cancelOrder(r, req.OrderID)
			writeJSON(w, http.StatusGone, dto.ErrorResponse{
				Code:    "PAYMENT_LINK_EXPIRED",
				Message: "payment link has expired, order has been cancelled",
			})
			return
		}

		if errors.Is(err, payment.ErrLinkNotFound) {
			writeJSON(w, http.StatusNotFound, dto.ErrorResponse{
				Code:    "PAYMENT_LINK_NOT_FOUND",
				Message: "no payment link found for this order",
			})
			return
		}

		if errors.Is(err, payment.ErrLinkUsed) {
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "PAYMENT_ALREADY_PROCESSED",
				Message: "payment has already been processed for this order",
			})
			return
		}

		if errors.Is(err, payment.ErrInvalidOrderStatus) {
			writeJSON(w, http.StatusConflict, dto.ErrorResponse{
				Code:    "INVALID_ORDER_STATUS",
				Message: "order is not in a state that can accept payment",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to process payment",
		})
		return
	}

	// Payment confirmed successfully
	writeJSON(w, http.StatusOK, dto.PaymentCallbackResponse{
		OrderID: req.OrderID,
		Status:  "confirmed",
		Message: "payment successful, order confirmed",
	})
}

// handleFailure processes a failed payment callback.
func (h *PaymentHandler) handleFailure(w http.ResponseWriter, r *http.Request, req dto.PaymentCallbackRequest) {
	h.mu.Lock()
	h.retries[req.OrderID]++
	retryCount := h.retries[req.OrderID]
	h.mu.Unlock()

	if retryCount >= maxPaymentRetries {
		// Max retries reached — cancel the order
		h.cancelOrder(r, req.OrderID)

		h.mu.Lock()
		delete(h.retries, req.OrderID)
		h.mu.Unlock()

		writeJSON(w, http.StatusOK, dto.PaymentCallbackResponse{
			OrderID: req.OrderID,
			Status:  "cancelled",
			Message: "maximum payment retries exceeded, order has been cancelled",
		})
		return
	}

	// Still has retries left — keep order in PendingPayment
	remaining := maxPaymentRetries - retryCount
	writeJSON(w, http.StatusOK, dto.PaymentCallbackResponse{
		OrderID: req.OrderID,
		Status:  "pending_payment",
		Message: fmt.Sprintf("payment failed, %d retry attempt(s) remaining", remaining),
	})
}

// cancelOrder sets the order status to Cancelled.
func (h *PaymentHandler) cancelOrder(r *http.Request, orderID string) {
	order, err := h.orderRepo.GetByID(r.Context(), orderID)
	if err != nil || order == nil {
		return
	}

	order.Status = domain.OrderStatusCancelled
	_ = h.orderRepo.Save(r.Context(), order)
}
