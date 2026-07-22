package payment

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

const (
	// LinkExpiry is the duration after which a payment link becomes invalid.
	LinkExpiry = 30 * time.Minute
)

var (
	ErrLinkNotFound = errors.New("payment link not found")
	ErrLinkExpired  = errors.New("payment link has expired")
	ErrLinkUsed     = errors.New("payment link has already been used")
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrderStatus = errors.New("order is not in pending_payment status")
)

// paymentLinkEntry tracks a generated payment link with its metadata.
type paymentLinkEntry struct {
	URL       string
	OrderID   string
	ExpiresAt time.Time
	Used      bool
}

// ServiceConfig holds dependencies for creating a PaymentService.
type ServiceConfig struct {
	Gateway          PaymentGateway
	OrderRepo        domain.OrderRepository
	Clock            func() time.Time // optional, defaults to time.Now
	OnOrderConfirmed func(ctx context.Context, orderID string) error // optional callback after payment confirmation
}

// paymentService is the concrete implementation of domain.PaymentService.
type paymentService struct {
	gateway          PaymentGateway
	orderRepo        domain.OrderRepository
	clock            func() time.Time
	onOrderConfirmed func(ctx context.Context, orderID string) error

	mu    sync.RWMutex
	links map[string]*paymentLinkEntry // keyed by orderID
}

// NewPaymentService creates a new PaymentService with the given configuration.
func NewPaymentService(cfg ServiceConfig) domain.PaymentService {
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	return &paymentService{
		gateway:          cfg.Gateway,
		orderRepo:        cfg.OrderRepo,
		clock:            clock,
		onOrderConfirmed: cfg.OnOrderConfirmed,
		links:            make(map[string]*paymentLinkEntry),
	}
}

// InitiatePayment generates a single-use payment link for an order.
// The link expires after 30 minutes.
func (s *paymentService) InitiatePayment(ctx context.Context, orderID string, amount int64) (*domain.PaymentLink, error) {
	url, err := s.gateway.CreateCheckoutURL(ctx, orderID, amount)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	entry := &paymentLinkEntry{
		URL:       url,
		OrderID:   orderID,
		ExpiresAt: now.Add(LinkExpiry),
		Used:      false,
	}

	s.mu.Lock()
	s.links[orderID] = entry
	s.mu.Unlock()

	return &domain.PaymentLink{
		URL:       url,
		ExpiresIn: int(LinkExpiry.Seconds()),
	}, nil
}

// ProcessPaymentCallback handles a payment gateway callback to confirm payment.
// It verifies the link exists, hasn't expired, and hasn't been used, then updates
// the order status from PendingPayment to Confirmed.
func (s *paymentService) ProcessPaymentCallback(ctx context.Context, orderID string, paymentRef string) error {
	s.mu.Lock()
	entry, exists := s.links[orderID]
	if !exists {
		s.mu.Unlock()
		return ErrLinkNotFound
	}

	now := s.clock()
	if now.After(entry.ExpiresAt) {
		s.mu.Unlock()
		return ErrLinkExpired
	}

	if entry.Used {
		s.mu.Unlock()
		return ErrLinkUsed
	}

	// Mark as used (single-use enforcement)
	entry.Used = true
	s.mu.Unlock()

	// Fetch the order and verify it's in the expected state
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPendingPayment {
		return ErrInvalidOrderStatus
	}

	// Update order status to Confirmed
	order.Status = domain.OrderStatusConfirmed
	order.UpdatedAt = now

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	// Notify listeners (e.g., send confirmation email, generate invoice)
	if s.onOrderConfirmed != nil {
		if err := s.onOrderConfirmed(ctx, orderID); err != nil {
			// Log but don't fail the payment — the order is already confirmed
			log.Printf("[PAYMENT] post-confirmation callback error for order %s: %v", orderID, err)
		}
	}

	return nil
}

// InitiateRefund initiates a refund for a cancelled order.
// In a real implementation this would call the payment gateway's refund API.
func (s *paymentService) InitiateRefund(_ context.Context, _ string, _ int64) error {
	// In production, this would call the payment gateway's refund endpoint.
	// For now, this is a no-op mock that always succeeds.
	return nil
}
