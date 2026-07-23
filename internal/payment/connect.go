package payment

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/account"
	"github.com/stripe/stripe-go/v82/accountlink"
	"github.com/stripe/stripe-go/v82/transfer"
	"github.com/stripe/stripe-go/v82/transferreversal"
)

// ConnectService handles Stripe Connect marketplace operations:
// onboarding bakeries, creating transfers, and reversing them on refund.
type ConnectService struct {
	stripeKey      string
	platformAcctID string
}

// ConnectConfig holds configuration for the Connect service.
type ConnectConfig struct {
	StripeKey      string
	PlatformAcctID string
}

// NewConnectService creates a new ConnectService.
func NewConnectService(cfg ConnectConfig) *ConnectService {
	return &ConnectService{
		stripeKey:      cfg.StripeKey,
		platformAcctID: cfg.PlatformAcctID,
	}
}

// CreateExpressAccount creates a new Stripe Connect Express account for a bakery.
// Returns the Stripe account ID (acct_...).
func (s *ConnectService) CreateExpressAccount(bakeryName string) (string, error) {
	stripe.Key = s.stripeKey

	params := &stripe.AccountParams{
		Type:         stripe.String("express"),
		Country:      stripe.String("BE"),
		BusinessType: stripe.String("individual"),
		Capabilities: &stripe.AccountCapabilitiesParams{
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		BusinessProfile: &stripe.AccountBusinessProfileParams{
			Name: stripe.String(bakeryName),
			MCC:  stripe.String("5462"), // Bakeries
		},
	}

	acct, err := account.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe connect: failed to create express account: %w", err)
	}

	return acct.ID, nil
}

// CreateAccountLink generates a Stripe Connect Express onboarding link for a bakery.
// refreshURL is where users go if the link expires; returnURL is where they go after completion.
func (s *ConnectService) CreateAccountLink(stripeAccountID string, refreshURL, returnURL string) (string, error) {
	stripe.Key = s.stripeKey

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe connect: failed to create account link: %w", err)
	}

	return link.URL, nil
}

// GetAccountStatus checks whether a connected account has completed onboarding.
func (s *ConnectService) GetAccountStatus(stripeAccountID string) (chargesEnabled bool, payoutsEnabled bool, err error) {
	stripe.Key = s.stripeKey

	acct, err := account.GetByID(stripeAccountID, nil)
	if err != nil {
		return false, false, fmt.Errorf("stripe connect: failed to get account: %w", err)
	}

	return acct.ChargesEnabled, acct.PayoutsEnabled, nil
}

// CreateTransfer transfers the bakery's share from the platform to the connected account.
// Returns the Stripe transfer ID (tr_...).
func (s *ConnectService) CreateTransfer(orderID string, amount int64, connectedAccountID string) (string, error) {
	stripe.Key = s.stripeKey

	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String("eur"),
		Destination: stripe.String(connectedAccountID),
		Metadata: map[string]string{
			"order_id": orderID,
		},
	}

	tr, err := transfer.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe connect: failed to create transfer for order %s: %w", orderID, err)
	}

	return tr.ID, nil
}

// ReverseTransfer reverses a transfer on refund.
func (s *ConnectService) ReverseTransfer(transferID string) error {
	stripe.Key = s.stripeKey

	params := &stripe.TransferReversalParams{
		ID: stripe.String(transferID),
	}

	_, err := transferreversal.New(params)
	if err != nil {
		return fmt.Errorf("stripe connect: failed to reverse transfer %s: %w", transferID, err)
	}

	return nil
}
