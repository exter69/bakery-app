package notification

import (
	"context"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/email"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEmailSender records sent messages for assertions.
type mockEmailSender struct {
	messages []email.EmailMessage
}

func (m *mockEmailSender) Send(_ context.Context, msg email.EmailMessage) error {
	m.messages = append(m.messages, msg)
	return nil
}

// mockOrderRepo implements domain.OrderRepository for testing.
type mockOrderRepo struct {
	orders map[string]*domain.Order
}

func (m *mockOrderRepo) Save(_ context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (*domain.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return order, nil
}

func (m *mockOrderRepo) ListByUser(_ context.Context, _ string, _ domain.OrderFilters, _ domain.PaginationParams) ([]domain.Order, int, error) {
	return nil, 0, nil
}

func (m *mockOrderRepo) ListByBakery(_ context.Context, _ string, _ domain.OrderFilters, _ domain.PaginationParams) ([]domain.Order, int, error) {
	return nil, 0, nil
}

// mockBakeryRepo implements domain.BakeryRepository for testing.
type mockBakeryRepo struct {
	bakeries map[string]*domain.Bakery
}

func (m *mockBakeryRepo) ListBakeries(_ context.Context, _ domain.PaginationParams) ([]domain.Bakery, int, error) {
	return nil, 0, nil
}

func (m *mockBakeryRepo) GetBakery(_ context.Context, id string) (*domain.Bakery, error) {
	bakery, ok := m.bakeries[id]
	if !ok {
		return nil, nil
	}
	return bakery, nil
}

func (m *mockBakeryRepo) GetBakeryByOwner(_ context.Context, _ string) (*domain.Bakery, error) {
	return nil, nil
}

func (m *mockBakeryRepo) UpdateBakery(_ context.Context, _ *domain.Bakery) error {
	return nil
}

func (m *mockBakeryRepo) GetProductsByBakery(_ context.Context, _ string) ([]domain.Product, error) {
	return nil, nil
}

func (m *mockBakeryRepo) GetProductByID(_ context.Context, _ string) (*domain.Product, error) {
	return nil, nil
}

func (m *mockBakeryRepo) CreateProduct(_ context.Context, _ *domain.Product) error {
	return nil
}

func (m *mockBakeryRepo) UpdateProduct(_ context.Context, _ *domain.Product) error {
	return nil
}

func (m *mockBakeryRepo) DeleteProduct(_ context.Context, _ string) error {
	return nil
}

// mockUserRepo implements domain.UserRepository for testing.
type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Save(_ context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*domain.User, error) {
	return nil, nil
}

func TestOnPaymentConfirmed_GeneratesInvoiceAndSendsEmail(t *testing.T) {
	// Arrange
	fixedTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	emailSender := &mockEmailSender{}
	invStore := invoice.NewStore()

	orderRepo := &mockOrderRepo{orders: map[string]*domain.Order{
		"order-1": {
			ID:          "order-1",
			BakeryID:    "bakery-1",
			UserID:      "user-1",
			Items:       []domain.OrderItem{{ProductID: "p1", ProductName: "Baguette", Quantity: 2, UnitPrice: 350, Subtotal: 700}},
			TotalAmount: 700,
			Status:      domain.OrderStatusConfirmed,
		},
	}}

	bakeryRepo := &mockBakeryRepo{bakeries: map[string]*domain.Bakery{
		"bakery-1": {ID: "bakery-1", Name: "Le Pain Doré", Address: "123 Rue Neuve, Brussels"},
	}}

	userRepo := &mockUserRepo{users: map[string]*domain.User{
		"user-1": {ID: "user-1", Username: "marie", ContactEmail: "marie@example.com"},
	}}

	svc := NewService(ServiceConfig{
		EmailSender:  emailSender,
		InvoiceStore: invStore,
		OrderRepo:    orderRepo,
		BakeryRepo:   bakeryRepo,
		UserRepo:     userRepo,
		Clock:        func() time.Time { return fixedTime },
	})

	// Act
	err := svc.OnPaymentConfirmed(context.Background(), "order-1")

	// Assert
	require.NoError(t, err)

	// Invoice was stored
	html, found := invStore.Get("order-1")
	assert.True(t, found)
	assert.Contains(t, html, "INV-order-1-")
	assert.Contains(t, html, "Baguette")
	assert.Contains(t, html, "€7.00")

	// Email was sent
	require.Len(t, emailSender.messages, 1)
	assert.Equal(t, "marie@example.com", emailSender.messages[0].To)
	assert.Contains(t, emailSender.messages[0].Subject, "Le Pain Doré")
	assert.Contains(t, emailSender.messages[0].Body, "Baguette")
}

func TestOnPaymentConfirmed_HandlesMissingOrder(t *testing.T) {
	// Arrange
	emailSender := &mockEmailSender{}
	invStore := invoice.NewStore()
	orderRepo := &mockOrderRepo{orders: map[string]*domain.Order{}}
	bakeryRepo := &mockBakeryRepo{bakeries: map[string]*domain.Bakery{}}
	userRepo := &mockUserRepo{users: map[string]*domain.User{}}

	svc := NewService(ServiceConfig{
		EmailSender:  emailSender,
		InvoiceStore: invStore,
		OrderRepo:    orderRepo,
		BakeryRepo:   bakeryRepo,
		UserRepo:     userRepo,
	})

	// Act
	err := svc.OnPaymentConfirmed(context.Background(), "nonexistent")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Empty(t, emailSender.messages)
}

func TestOnPaymentConfirmed_HandlesMissingBakery(t *testing.T) {
	// Arrange
	emailSender := &mockEmailSender{}
	invStore := invoice.NewStore()
	orderRepo := &mockOrderRepo{orders: map[string]*domain.Order{
		"order-2": {ID: "order-2", BakeryID: "missing-bakery", UserID: "user-1"},
	}}
	bakeryRepo := &mockBakeryRepo{bakeries: map[string]*domain.Bakery{}}
	userRepo := &mockUserRepo{users: map[string]*domain.User{
		"user-1": {ID: "user-1", Username: "test", ContactEmail: "test@example.com"},
	}}

	svc := NewService(ServiceConfig{
		EmailSender:  emailSender,
		InvoiceStore: invStore,
		OrderRepo:    orderRepo,
		BakeryRepo:   bakeryRepo,
		UserRepo:     userRepo,
	})

	// Act
	err := svc.OnPaymentConfirmed(context.Background(), "order-2")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bakery")
	assert.Empty(t, emailSender.messages)
}
