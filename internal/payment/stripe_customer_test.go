package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
)

func TestNewStripeCustomerService_StoresConfig(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_key123", repo)

	assert.Equal(t, "sk_test_key123", svc.secretKey)
	assert.NotNil(t, svc.userRepo)
}

func TestStripeCustomerService_GetOrCreateCustomer_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	_, err := svc.GetOrCreateCustomer(context.Background(), "nonexistent-user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_GetOrCreateCustomer_ReturnsExistingID(t *testing.T) {
	repo := memory.NewUserRepo()
	user := &domain.User{
		ID:               "user-1",
		Username:         "testuser",
		StripeCustomerID: "cus_existing123",
	}
	require.NoError(t, repo.Save(context.Background(), user))

	svc := NewStripeCustomerService("sk_test_invalid", repo)

	// Should return the existing customer ID without hitting Stripe
	customerID, err := svc.GetOrCreateCustomer(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "cus_existing123", customerID)
}

func TestStripeCustomerService_ListPaymentMethods_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	_, err := svc.ListPaymentMethods(context.Background(), "nonexistent-user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_CreateSetupIntent_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	_, err := svc.CreateSetupIntent(context.Background(), "nonexistent-user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_DetachPaymentMethod_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	err := svc.DetachPaymentMethod(context.Background(), "nonexistent-user", "pm_123")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_SetDefaultPaymentMethod_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	err := svc.SetDefaultPaymentMethod(context.Background(), "nonexistent-user", "pm_123")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_ChargeWithSavedMethod_UserNotFound(t *testing.T) {
	repo := memory.NewUserRepo()
	svc := NewStripeCustomerService("sk_test_invalid", repo)

	_, err := svc.ChargeWithSavedMethod(context.Background(), "nonexistent-user", "pm_123", 5000, "order-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestStripeCustomerService_CreateSetupIntent_InvalidKey(t *testing.T) {
	repo := memory.NewUserRepo()
	user := &domain.User{
		ID:               "user-1",
		Username:         "testuser",
		StripeCustomerID: "cus_123",
	}
	require.NoError(t, repo.Save(context.Background(), user))

	svc := NewStripeCustomerService("sk_test_invalid", repo)

	// With an invalid key, Stripe's API will return an auth error
	_, err := svc.CreateSetupIntent(context.Background(), "user-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stripe: failed to create setup intent")
}
