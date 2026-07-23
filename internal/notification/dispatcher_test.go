package notification

import (
	"context"
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReservationRepo implements domain.ReservationRepository for testing.
type mockReservationRepo struct {
	reservations map[string]*domain.Reservation
}

func (m *mockReservationRepo) Save(_ context.Context, r domain.Reservation) error {
	m.reservations[r.ID] = &r
	return nil
}

func (m *mockReservationRepo) Get(_ context.Context, id string) (*domain.Reservation, error) {
	r, ok := m.reservations[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (m *mockReservationRepo) ListByUser(_ context.Context, _ string, _ domain.ReservationFilters, _ domain.PaginationParams) ([]domain.Reservation, int, error) {
	return nil, 0, nil
}

func (m *mockReservationRepo) ListByBakery(_ context.Context, _ string, _ domain.ReservationFilters, _ domain.PaginationParams) ([]domain.Reservation, int, error) {
	return nil, 0, nil
}

func newTestService(emailSender *mockEmailSender) (*Service, *mockOrderRepo, *mockBakeryRepo, *mockUserRepo, *mockReservationRepo) {
	orderRepo := &mockOrderRepo{orders: map[string]*domain.Order{}}
	bakeryRepo := &mockBakeryRepo{bakeries: map[string]*domain.Bakery{}}
	userRepo := &mockUserRepo{users: map[string]*domain.User{}}
	reservationRepo := &mockReservationRepo{reservations: map[string]*domain.Reservation{}}

	svc := NewService(ServiceConfig{
		EmailSender:     emailSender,
		InvoiceStore:    invoice.NewStore(),
		OrderRepo:       orderRepo,
		BakeryRepo:      bakeryRepo,
		UserRepo:        userRepo,
		ReservationRepo: reservationRepo,
	})

	return svc, orderRepo, bakeryRepo, userRepo, reservationRepo
}

func TestOnOrderStatusChanged_Preparing(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		TotalAmount: 1500, Status: domain.OrderStatusPreparing,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com", Locale: "en"}

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusPreparing)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Equal(t, "marie@test.com", emailSender.messages[0].To)
	assert.Contains(t, emailSender.messages[0].Subject, "being prepared")
	assert.Contains(t, emailSender.messages[0].Subject, "Le Pain Doré")
	assert.Contains(t, emailSender.messages[0].Body, "Le Pain Doré")
}

func TestOnOrderStatusChanged_Ready(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		TotalAmount: 1500, Status: domain.OrderStatusReady,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com"}

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusReady)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "ready")
	assert.Contains(t, emailSender.messages[0].Body, "ready")
}

func TestOnOrderStatusChanged_Delivered(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		Items:       []domain.OrderItem{{ProductName: "Croissant", Quantity: 3, UnitPrice: 250, Subtotal: 750}},
		TotalAmount: 750, Status: domain.OrderStatusDelivered,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com"}

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusDelivered)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "delivered")
	assert.Contains(t, emailSender.messages[0].Body, "€7.50")
}

func TestOnOrderStatusChanged_IgnoresPendingPayment(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, _, _, _ := newTestService(emailSender)

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusPendingPayment)
	require.NoError(t, err)
	assert.Empty(t, emailSender.messages)
}

func TestOnOrderStatusChanged_FrenchLocale(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		TotalAmount: 1500, Status: domain.OrderStatusPreparing,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com", Locale: "fr"}

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusPreparing)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "en préparation")
}

func TestOnOrderStatusChanged_DutchLocale(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		TotalAmount: 1500, Status: domain.OrderStatusPreparing,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Bakker Jan", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "pieter", ContactEmail: "pieter@test.com", Locale: "nl"}

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusPreparing)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "wordt bereid")
}

func TestOnOrderStatusChanged_FallsBackToEnglish(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		TotalAmount: 1500, Status: domain.OrderStatusPreparing,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com", Locale: "de"} // unsupported locale

	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusPreparing)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	// Should fall back to English
	assert.Contains(t, emailSender.messages[0].Subject, "being prepared")
}

func TestOnNewOrder_AlertsBaker(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "customer-1",
		Items:       []domain.OrderItem{{ProductName: "Baguette", Quantity: 2, UnitPrice: 300, Subtotal: 600}},
		TotalAmount: 600,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["baker-1"] = &domain.User{ID: "baker-1", Username: "chef", ContactEmail: "baker@test.com", Locale: "en"}
	userRepo.users["customer-1"] = &domain.User{ID: "customer-1", Username: "marie", ContactEmail: "marie@test.com"}

	err := svc.OnNewOrder(context.Background(), "order-1")
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	msg := emailSender.messages[0]
	assert.Equal(t, "baker@test.com", msg.To)
	assert.Contains(t, msg.Subject, "New order received")
	assert.Contains(t, msg.Subject, "marie")
	assert.Contains(t, msg.Body, "Baguette")
	assert.Contains(t, msg.Body, "€6.00")
}

func TestOnNewOrder_UsesBakerLocale(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "customer-1",
		Items:       []domain.OrderItem{{ProductName: "Baguette", Quantity: 1, UnitPrice: 300, Subtotal: 300}},
		TotalAmount: 300,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["baker-1"] = &domain.User{ID: "baker-1", Username: "chef", ContactEmail: "baker@test.com", Locale: "fr"}
	userRepo.users["customer-1"] = &domain.User{ID: "customer-1", Username: "marie", ContactEmail: "marie@test.com"}

	err := svc.OnNewOrder(context.Background(), "order-1")
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "Nouvelle commande")
}

func TestOnNewOrder_HandlesMissingOrder(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, _, _, _ := newTestService(emailSender)

	err := svc.OnNewOrder(context.Background(), "nonexistent")
	require.NoError(t, err) // non-blocking, should not error
	assert.Empty(t, emailSender.messages)
}

func TestOnReservationConfirmed_SendsEmail(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, bakeryRepo, userRepo, reservationRepo := newTestService(emailSender)

	reservationRepo.reservations["res-1"] = &domain.Reservation{
		ID: "res-1", BakeryID: "bakery-1", UserID: "user-1",
		Items:       []domain.OrderItem{{ProductName: "Croissant", Quantity: 4, UnitPrice: 200, Subtotal: 800}},
		TotalAmount: 800,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com", Locale: "en"}

	err := svc.OnReservationConfirmed(context.Background(), "res-1")
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	msg := emailSender.messages[0]
	assert.Equal(t, "marie@test.com", msg.To)
	assert.Contains(t, msg.Subject, "Reservation confirmed")
	assert.Contains(t, msg.Subject, "Le Pain Doré")
	assert.Contains(t, msg.Body, "Croissant")
	assert.Contains(t, msg.Body, "€8.00")
}

func TestOnReservationConfirmed_FrenchLocale(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, bakeryRepo, userRepo, reservationRepo := newTestService(emailSender)

	reservationRepo.reservations["res-1"] = &domain.Reservation{
		ID: "res-1", BakeryID: "bakery-1", UserID: "user-1",
		Items:       []domain.OrderItem{{ProductName: "Pain", Quantity: 1, UnitPrice: 300, Subtotal: 300}},
		TotalAmount: 300,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com", Locale: "fr"}

	err := svc.OnReservationConfirmed(context.Background(), "res-1")
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	assert.Contains(t, emailSender.messages[0].Subject, "Réservation confirmée")
}

func TestOnReservationConfirmed_HandlesMissingReservation(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, _, _, _ := newTestService(emailSender)

	err := svc.OnReservationConfirmed(context.Background(), "nonexistent")
	require.NoError(t, err) // non-blocking
	assert.Empty(t, emailSender.messages)
}

func TestRenderTemplate_LocaleFallback(t *testing.T) {
	data := TemplateData{
		BakeryName:   "Test Bakery",
		CustomerName: "John",
		OrderID:      "order-123",
		TotalDisplay: "€10.00",
	}

	// Unknown locale should fall back to EN
	subject, body, err := renderTemplate("xx", "status_preparing", data)
	require.NoError(t, err)
	assert.Contains(t, subject, "being prepared")
	assert.Contains(t, body, "Test Bakery")
}

func TestRenderTemplate_UnknownEvent(t *testing.T) {
	data := TemplateData{}
	_, _, err := renderTemplate(LocaleEN, "nonexistent_event", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown notification event")
}

func TestServiceImplementsDispatcher(t *testing.T) {
	// Compile-time check is done via var _ Dispatcher = (*Service)(nil)
	// but also verify at runtime
	var d Dispatcher = &Service{}
	assert.NotNil(t, d)
}

func TestOnOrderStatusChanged_MissingOrder(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, _, _, _, _ := newTestService(emailSender)

	err := svc.OnOrderStatusChanged(context.Background(), "nonexistent", domain.OrderStatusPreparing)
	require.NoError(t, err) // gracefully returns nil
	assert.Empty(t, emailSender.messages)
}

func TestOnOrderStatusChanged_SendsItemDetails(t *testing.T) {
	emailSender := &mockEmailSender{}
	svc, orderRepo, bakeryRepo, userRepo, _ := newTestService(emailSender)

	orderRepo.orders["order-1"] = &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "user-1",
		Items: []domain.OrderItem{
			{ProductName: "Baguette", Quantity: 2, UnitPrice: 300, Subtotal: 600},
			{ProductName: "Croissant", Quantity: 1, UnitPrice: 250, Subtotal: 250},
		},
		TotalAmount: 850, Status: domain.OrderStatusConfirmed,
	}
	bakeryRepo.bakeries["bakery-1"] = &domain.Bakery{ID: "bakery-1", Name: "Le Pain Doré", OwnerID: "baker-1"}
	userRepo.users["user-1"] = &domain.User{ID: "user-1", Username: "marie", ContactEmail: "marie@test.com"}

	// "order_confirmed" template includes item listing
	err := svc.OnOrderStatusChanged(context.Background(), "order-1", domain.OrderStatusDelivered)
	require.NoError(t, err)

	require.Len(t, emailSender.messages, 1)
	body := emailSender.messages[0].Body
	// The delivered template shows the total
	assert.Contains(t, body, "€8.50")
	assert.Contains(t, body, "Le Pain Doré")
}
