package notification

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/email"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
	"github.com/lucatorrekens/bakery-app/internal/push"
	"github.com/lucatorrekens/bakery-app/internal/ws"
)

// Dispatcher is an optional dependency for sending notifications from other services.
// Methods are non-blocking safe: callers should invoke them in goroutines.
type Dispatcher interface {
	OnOrderStatusChanged(ctx context.Context, orderID string, newStatus domain.OrderStatus) error
	OnNewOrder(ctx context.Context, orderID string) error
	OnReservationConfirmed(ctx context.Context, reservationID string) error
}

// ServiceConfig holds dependencies for creating a notification Service.
type ServiceConfig struct {
	EmailSender     email.Sender
	InvoiceStore    *invoice.Store
	OrderRepo       domain.OrderRepository
	BakeryRepo      domain.BakeryRepository
	UserRepo        domain.UserRepository
	ReservationRepo domain.ReservationRepository
	WSHub           *ws.Hub        // optional WebSocket hub for real-time push
	PushSender      *push.Sender   // optional Web Push sender for PWA notifications
	Clock           func() time.Time // optional, defaults to time.Now
}

// Service orchestrates sending confirmation emails and generating invoices after payment.
// It implements the Dispatcher interface.
type Service struct {
	emailSender     email.Sender
	invoiceStore    *invoice.Store
	orderRepo       domain.OrderRepository
	bakeryRepo      domain.BakeryRepository
	userRepo        domain.UserRepository
	reservationRepo domain.ReservationRepository
	wsHub           *ws.Hub
	pushSender      *push.Sender
	clock           func() time.Time
}

// Compile-time check that Service implements Dispatcher.
var _ Dispatcher = (*Service)(nil)

// NewService creates a new notification Service.
func NewService(cfg ServiceConfig) *Service {
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		emailSender:     cfg.EmailSender,
		invoiceStore:    cfg.InvoiceStore,
		orderRepo:       cfg.OrderRepo,
		bakeryRepo:      cfg.BakeryRepo,
		userRepo:        cfg.UserRepo,
		reservationRepo: cfg.ReservationRepo,
		wsHub:           cfg.WSHub,
		pushSender:      cfg.PushSender,
		clock:           clock,
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

// OnOrderCancelled sends a cancellation notification to the customer.
// If refunded is true, it informs the customer that a refund was issued.
func (s *Service) OnOrderCancelled(ctx context.Context, orderID string, refunded bool) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		return fmt.Errorf("fetching order %s for cancellation notification: %w", orderID, err)
	}

	user, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil || user == nil {
		return nil // can't send notification without user info
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil || bakery == nil {
		return nil
	}

	subject := fmt.Sprintf("Order Cancelled - %s", bakery.Name)
	body := fmt.Sprintf("<h2>Order Cancelled</h2><p>Your order from %s has been cancelled.</p>", bakery.Name)
	if refunded {
		body += "<p><strong>A refund has been issued to your original payment method.</strong> It may take 5-10 business days to appear on your statement.</p>"
	} else {
		body += "<p>No charge was made — the authorization hold on your card has been released.</p>"
	}

	return s.emailSender.Send(ctx, email.EmailMessage{
		To:      user.ContactEmail,
		Subject: subject,
		Body:    body,
	})
}


// OnOrderStatusChanged sends an email to the customer when their order status changes.
// It maps the status to the appropriate template event.
func (s *Service) OnOrderStatusChanged(ctx context.Context, orderID string, newStatus domain.OrderStatus) error {
	// Determine template event from status
	var event string
	switch newStatus {
	case domain.OrderStatusPreparing:
		event = "status_preparing"
	case domain.OrderStatusReady:
		event = "status_ready"
	case domain.OrderStatusDelivered:
		event = "status_delivered"
	default:
		// No notification for other statuses (pending_payment, confirmed handled by OnPaymentConfirmed, cancelled handled by OnOrderCancelled)
		return nil
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		log.Printf("[NOTIFICATION] failed to fetch order %s for status notification: %v", orderID, err)
		return nil
	}

	user, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil || user == nil {
		log.Printf("[NOTIFICATION] failed to fetch user for order %s: %v", orderID, err)
		return nil
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil || bakery == nil {
		log.Printf("[NOTIFICATION] failed to fetch bakery for order %s: %v", orderID, err)
		return nil
	}

	locale := userLocale(user)
	data := TemplateData{
		BakeryName:   bakery.Name,
		CustomerName: user.Username,
		OrderID:      order.ID,
		Items:        buildItemData(order.Items),
		TotalDisplay: formatCentsForDisplay(order.TotalAmount),
	}

	subject, body, err := renderTemplate(locale, event, data)
	if err != nil {
		log.Printf("[NOTIFICATION] failed to render template %s for order %s: %v", event, orderID, err)
		return nil
	}

	if err := s.emailSender.Send(ctx, email.EmailMessage{
		To:      user.ContactEmail,
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Printf("[NOTIFICATION] failed to send %s email for order %s: %v", event, orderID, err)
	}

	// Push real-time WebSocket event to the customer
	if s.wsHub != nil {
		s.wsHub.SendToUser(order.UserID, ws.Event{
			Type: "order_status",
			Payload: map[string]string{
				"orderID": orderID,
				"status":  string(newStatus),
			},
		})
	}

	// Send Web Push notification to the customer
	if s.pushSender != nil {
		s.pushSender.SendToUser(order.UserID, push.PushMessage{
			Title: subject,
			Body:  fmt.Sprintf("Order from %s — %s", bakery.Name, string(newStatus)),
			URL:   "/schedule",
		})
	}

	return nil
}

// OnNewOrder alerts the baker that a new order has been received.
func (s *Service) OnNewOrder(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		log.Printf("[NOTIFICATION] failed to fetch order %s for baker alert: %v", orderID, err)
		return nil
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, order.BakeryID)
	if err != nil || bakery == nil {
		log.Printf("[NOTIFICATION] failed to fetch bakery for baker alert, order %s: %v", orderID, err)
		return nil
	}

	// Get baker (owner) info to send the alert
	baker, err := s.userRepo.GetByID(ctx, bakery.OwnerID)
	if err != nil || baker == nil {
		log.Printf("[NOTIFICATION] failed to fetch baker %s for new order alert: %v", bakery.OwnerID, err)
		return nil
	}

	// Get customer info for the template
	customer, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil || customer == nil {
		log.Printf("[NOTIFICATION] failed to fetch customer for new order alert, order %s: %v", orderID, err)
		return nil
	}

	locale := userLocale(baker)
	data := TemplateData{
		BakeryName:   bakery.Name,
		CustomerName: customer.Username,
		OrderID:      order.ID,
		Items:        buildItemData(order.Items),
		TotalDisplay: formatCentsForDisplay(order.TotalAmount),
	}

	subject, body, err := renderTemplate(locale, "new_order_baker", data)
	if err != nil {
		log.Printf("[NOTIFICATION] failed to render new_order_baker template for order %s: %v", orderID, err)
		return nil
	}

	if err := s.emailSender.Send(ctx, email.EmailMessage{
		To:      baker.ContactEmail,
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Printf("[NOTIFICATION] failed to send baker alert for order %s: %v", orderID, err)
	}

	// Push real-time WebSocket event to the baker
	if s.wsHub != nil {
		s.wsHub.SendToUser(bakery.OwnerID, ws.Event{
			Type: "new_order",
			Payload: map[string]string{
				"orderID":      orderID,
				"customerName": customer.Username,
			},
		})
	}

	// Send Web Push notification to the baker
	if s.pushSender != nil {
		s.pushSender.SendToUser(bakery.OwnerID, push.PushMessage{
			Title: "New order received!",
			Body:  fmt.Sprintf("Order from %s", customer.Username),
			URL:   "/dashboard/orders",
		})
	}

	return nil
}

// OnReservationConfirmed sends a confirmation email to the customer after a reservation is created.
func (s *Service) OnReservationConfirmed(ctx context.Context, reservationID string) error {
	if s.reservationRepo == nil {
		log.Printf("[NOTIFICATION] reservation repo not configured, skipping notification for %s", reservationID)
		return nil
	}

	reservation, err := s.reservationRepo.Get(ctx, reservationID)
	if err != nil || reservation == nil {
		log.Printf("[NOTIFICATION] failed to fetch reservation %s: %v", reservationID, err)
		return nil
	}

	user, err := s.userRepo.GetByID(ctx, reservation.UserID)
	if err != nil || user == nil {
		log.Printf("[NOTIFICATION] failed to fetch user for reservation %s: %v", reservationID, err)
		return nil
	}

	bakery, err := s.bakeryRepo.GetBakery(ctx, reservation.BakeryID)
	if err != nil || bakery == nil {
		log.Printf("[NOTIFICATION] failed to fetch bakery for reservation %s: %v", reservationID, err)
		return nil
	}

	locale := userLocale(user)
	data := TemplateData{
		BakeryName:   bakery.Name,
		CustomerName: user.Username,
		OrderID:      reservation.ID,
		Items:        buildItemData(reservation.Items),
		TotalDisplay: formatCentsForDisplay(reservation.TotalAmount),
	}

	subject, body, err := renderTemplate(locale, "reservation_confirmed", data)
	if err != nil {
		log.Printf("[NOTIFICATION] failed to render reservation_confirmed template for %s: %v", reservationID, err)
		return nil
	}

	if err := s.emailSender.Send(ctx, email.EmailMessage{
		To:      user.ContactEmail,
		Subject: subject,
		Body:    body,
	}); err != nil {
		log.Printf("[NOTIFICATION] failed to send reservation confirmation for %s: %v", reservationID, err)
	}

	// Push real-time WebSocket event to the customer
	if s.wsHub != nil {
		s.wsHub.SendToUser(reservation.UserID, ws.Event{
			Type: "reservation_status",
			Payload: map[string]string{
				"reservationID": reservationID,
				"status":        "confirmed",
			},
		})
	}

	// Send Web Push notification to the customer
	if s.pushSender != nil {
		s.pushSender.SendToUser(reservation.UserID, push.PushMessage{
			Title: "Reservation confirmed",
			Body:  fmt.Sprintf("Your reservation at %s is confirmed", bakery.Name),
			URL:   "/schedule",
		})
	}

	return nil
}

// userLocale returns the user's preferred locale, defaulting to English.
func userLocale(user *domain.User) Locale {
	switch Locale(user.Locale) {
	case LocaleFR:
		return LocaleFR
	case LocaleNL:
		return LocaleNL
	default:
		return LocaleEN
	}
}

// buildItemData converts domain order items to template-friendly data.
func buildItemData(items []domain.OrderItem) []ItemData {
	result := make([]ItemData, len(items))
	for i, item := range items {
		result[i] = ItemData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Subtotal:    formatCentsForDisplay(item.Subtotal),
		}
	}
	return result
}
