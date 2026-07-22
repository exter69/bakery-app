package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/email"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
)

// ServiceConfig holds dependencies for creating a notification Service.
type ServiceConfig struct {
	EmailSender  email.Sender
	InvoiceStore *invoice.Store
	OrderRepo    domain.OrderRepository
	BakeryRepo   domain.BakeryRepository
	UserRepo     domain.UserRepository
	Clock        func() time.Time // optional, defaults to time.Now
}

// Service orchestrates sending confirmation emails and generating invoices after payment.
type Service struct {
	emailSender  email.Sender
	invoiceStore *invoice.Store
	orderRepo    domain.OrderRepository
	bakeryRepo   domain.BakeryRepository
	userRepo     domain.UserRepository
	clock        func() time.Time
}

// NewService creates a new notification Service.
func NewService(cfg ServiceConfig) *Service {
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		emailSender:  cfg.EmailSender,
		invoiceStore: cfg.InvoiceStore,
		orderRepo:    cfg.OrderRepo,
		bakeryRepo:   cfg.BakeryRepo,
		userRepo:     cfg.UserRepo,
		clock:        clock,
	}
}

// OnPaymentConfirmed is called after an order is confirmed via payment.
// It generates an invoice and sends a confirmation email to the customer.
func (s *Service) OnPaymentConfirmed(ctx context.Context, orderID string) error {
	// 1. Fetch order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching order %s: %w", orderID, err)
	}
	if order == nil {
		return fmt.Errorf("order %s not found", orderID)
	}

	// 2. Fetch bakery
	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil {
		return fmt.Errorf("fetching bakery %s: %w", order.BakeryID, err)
	}
	if bakery == nil {
		return fmt.Errorf("bakery %s not found", order.BakeryID)
	}

	// 3. Fetch user
	user, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil {
		return fmt.Errorf("fetching user %s: %w", order.UserID, err)
	}
	if user == nil {
		return fmt.Errorf("user %s not found", order.UserID)
	}

	// 4. Generate invoice
	now := s.clock()
	invData := invoice.InvoiceData{
		InvoiceNumber: invoice.GenerateInvoiceNumber(orderID, now),
		OrderID:       orderID,
		Date:          now,
		CustomerName:  user.Username,
		CustomerEmail: user.ContactEmail,
		BakeryName:    bakery.Name,
		BakeryAddress: bakery.Address,
		Items:         order.Items,
		TotalCents:    order.TotalAmount,
	}

	html, err := invoice.Generate(invData)
	if err != nil {
		return fmt.Errorf("generating invoice for order %s: %w", orderID, err)
	}

	// 5. Store invoice
	s.invoiceStore.Save(orderID, html)

	// 6. Send confirmation email
	emailMsg := email.EmailMessage{
		To:      user.ContactEmail,
		Subject: fmt.Sprintf("Order Confirmed - %s", bakery.Name),
		Body:    html,
	}

	if err := s.emailSender.Send(ctx, emailMsg); err != nil {
		// Log the error but don't fail the payment flow — the invoice is already stored
		log.Printf("[NOTIFICATION] failed to send confirmation email for order %s: %v", orderID, err)
		return nil
	}

	return nil
}
