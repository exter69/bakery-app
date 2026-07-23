package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/payment"
)

// Payout-specific errors.
var (
	ErrPayoutNotFound       = errors.New("payout not found")
	ErrNoConnectedAccount   = errors.New("bakery has no connected Stripe account")
	ErrPayoutAlreadyExists  = errors.New("payout already exists for this order")
)

// PayoutServiceConfig holds dependencies for the payout service.
type PayoutServiceConfig struct {
	ConnectSvc *payment.ConnectService
	PayoutRepo domain.PayoutRepository
	BakeryRepo domain.BakeryRepository
	OrderRepo  domain.OrderRepository
}

// PayoutService handles marketplace payout logic.
type PayoutService struct {
	connectSvc *payment.ConnectService
	payoutRepo domain.PayoutRepository
	bakeryRepo domain.BakeryRepository
	orderRepo  domain.OrderRepository
}

// NewPayoutService creates a new PayoutService.
func NewPayoutService(cfg PayoutServiceConfig) *PayoutService {
	return &PayoutService{
		connectSvc: cfg.ConnectSvc,
		payoutRepo: cfg.PayoutRepo,
		bakeryRepo: cfg.BakeryRepo,
		orderRepo:  cfg.OrderRepo,
	}
}

// OnOrderDelivered is called when an order transitions to "delivered".
// It computes the split (order total * commission_rate / 100 = platform fee),
// creates a payout record, and initiates the Stripe transfer.
func (s *PayoutService) OnOrderDelivered(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching order: %w", err)
	}
	if order == nil {
		return ErrOrderNotFound
	}

	// Check if payout already exists (idempotency)
	existing, err := s.payoutRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("checking existing payout: %w", err)
	}
	if existing != nil {
		return ErrPayoutAlreadyExists
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil {
		return fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return ErrBakeryNotFound
	}

	if bakery.StripeConnectID == "" {
		// No connected account — log and create a pending payout for manual resolution
		log.Printf("[PAYOUT] bakery %s has no connected Stripe account, creating pending payout for order %s", bakery.ID, orderID)
	}

	// Compute split
	commission := order.TotalAmount * int64(bakery.CommissionRate) / 100
	bakeryShare := order.TotalAmount - commission

	now := time.Now()
	payout := &domain.Payout{
		OrderID:    orderID,
		BakeryID:   bakery.ID,
		Amount:     bakeryShare,
		Commission: commission,
		Status:     domain.PayoutStatusPending,
		CreatedAt:  now,
	}

	if err := s.payoutRepo.Create(ctx, payout); err != nil {
		return fmt.Errorf("creating payout: %w", err)
	}

	// Only attempt the transfer if the bakery has a connected account
	if bakery.StripeConnectID != "" && s.connectSvc != nil {
		transferID, err := s.connectSvc.CreateTransfer(orderID, bakeryShare, bakery.StripeConnectID)
		if err != nil {
			// Mark as failed but don't fail the overall operation — the order is already delivered
			payout.Status = domain.PayoutStatusFailed
			if updateErr := s.payoutRepo.Update(ctx, payout); updateErr != nil {
				log.Printf("[PAYOUT] failed to mark payout as failed for order %s: %v", orderID, updateErr)
			}
			log.Printf("[PAYOUT] transfer failed for order %s: %v", orderID, err)
			return nil
		}

		transferredAt := time.Now()
		payout.StripeTransferID = transferID
		payout.Status = domain.PayoutStatusTransferred
		payout.TransferredAt = &transferredAt

		if err := s.payoutRepo.Update(ctx, payout); err != nil {
			log.Printf("[PAYOUT] failed to update payout status for order %s: %v", orderID, err)
		}
	}

	return nil
}

// OnOrderRefunded reverses the transfer for a refunded order.
func (s *PayoutService) OnOrderRefunded(ctx context.Context, orderID string) error {
	payout, err := s.payoutRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching payout: %w", err)
	}
	if payout == nil {
		// No payout exists for this order — nothing to reverse
		return nil
	}

	if payout.Status != domain.PayoutStatusTransferred {
		// Only reverse transferred payouts; pending/failed have no transfer to reverse
		payout.Status = domain.PayoutStatusRefunded
		return s.payoutRepo.Update(ctx, payout)
	}

	if payout.StripeTransferID != "" && s.connectSvc != nil {
		if err := s.connectSvc.ReverseTransfer(payout.StripeTransferID); err != nil {
			log.Printf("[PAYOUT] failed to reverse transfer %s for order %s: %v", payout.StripeTransferID, orderID, err)
			return fmt.Errorf("reversing transfer: %w", err)
		}
	}

	payout.Status = domain.PayoutStatusRefunded
	if err := s.payoutRepo.Update(ctx, payout); err != nil {
		return fmt.Errorf("updating payout status: %w", err)
	}

	return nil
}

// ListPayouts returns payout history for a bakery.
func (s *PayoutService) ListPayouts(ctx context.Context, bakeryID string, params domain.PaginationParams) (*domain.ListResult[domain.Payout], error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	payouts, total, err := s.payoutRepo.ListByBakery(ctx, bakeryID, params)
	if err != nil {
		return nil, fmt.Errorf("listing payouts: %w", err)
	}

	return &domain.ListResult[domain.Payout]{
		Items:    payouts,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

// GetConnectStatus returns the bakery's Stripe Connect status from persisted fields.
// The onboarding flags are synced via the account.updated webhook, avoiding a live Stripe API call.
func (s *PayoutService) GetConnectStatus(ctx context.Context, bakeryID string) (connected bool, chargesEnabled bool, payoutsEnabled bool, err error) {
	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return false, false, false, fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return false, false, false, ErrBakeryNotFound
	}

	if bakery.StripeConnectID == "" {
		return false, false, false, nil
	}

	return true, bakery.ChargesEnabled, bakery.PayoutsEnabled, nil
}

// Onboard generates a Stripe Connect Express onboarding link for the bakery.
func (s *PayoutService) Onboard(ctx context.Context, bakeryID string, refreshURL, returnURL string) (string, error) {
	if s.connectSvc == nil {
		return "", errors.New("Stripe Connect is not configured")
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, bakeryID)
	if err != nil {
		return "", fmt.Errorf("fetching bakery: %w", err)
	}
	if bakery == nil {
		return "", ErrBakeryNotFound
	}

	// Create a new Express account if the bakery doesn't have one yet
	if bakery.StripeConnectID == "" {
		accountID, err := s.connectSvc.CreateExpressAccount(bakery.Name)
		if err != nil {
			return "", fmt.Errorf("creating express account: %w", err)
		}
		bakery.StripeConnectID = accountID
		if err := s.bakeryRepo.UpdateBakery(ctx, bakery); err != nil {
			return "", fmt.Errorf("saving stripe connect ID: %w", err)
		}
	}

	link, err := s.connectSvc.CreateAccountLink(bakery.StripeConnectID, refreshURL, returnURL)
	if err != nil {
		return "", fmt.Errorf("creating account link: %w", err)
	}

	return link, nil
}
