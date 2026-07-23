package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// B2B sentinel errors.
var (
	ErrCompanyAlreadyExists = errors.New("company with this VAT/SIRET already exists")
	ErrAccessDenied         = errors.New("access denied: no approved access to this bakery")
	ErrBelowMinimum         = errors.New("order total is below the bakery minimum")
	ErrCutoffPassed         = errors.New("cutoff time has passed for this bakery")
	ErrLastSite             = errors.New("cannot delete the only remaining delivery site")
	ErrNoDeliverySite       = errors.New("no delivery site selected")
	ErrAccessExists         = errors.New("access request already exists for this bakery")
	ErrProfileNotFound      = errors.New("business profile not found")
	ErrSiteNotFound         = errors.New("delivery site not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
)

// B2BServiceImpl implements domain.B2BService.
type B2BServiceImpl struct {
	b2bRepo    domain.B2BRepository
	userRepo   domain.UserRepository
	bakeryRepo domain.BakeryRepository
	orderRepo  domain.OrderRepository
	jwtSecret  string
	now        func() time.Time
}

// B2BServiceConfig holds dependencies for the B2B service.
type B2BServiceConfig struct {
	B2BRepo    domain.B2BRepository
	UserRepo   domain.UserRepository
	BakeryRepo domain.BakeryRepository
	OrderRepo  domain.OrderRepository
	JWTSecret  string
	Now        func() time.Time
}

// NewB2BService creates a new B2BServiceImpl.
func NewB2BService(cfg B2BServiceConfig) *B2BServiceImpl {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &B2BServiceImpl{
		b2bRepo:    cfg.B2BRepo,
		userRepo:   cfg.UserRepo,
		bakeryRepo: cfg.BakeryRepo,
		orderRepo:  cfg.OrderRepo,
		jwtSecret:  cfg.JWTSecret,
		now:        now,
	}
}

// --- Registration & Profile ---

func (s *B2BServiceImpl) RegisterBusiness(ctx context.Context, req domain.RegisterBusinessRequest) (*domain.BusinessProfile, string, error) {
	if err := domain.ValidateBusinessRegistration(req); err != nil {
		return nil, "", err
	}

	// Check VAT uniqueness
	existing, err := s.b2bRepo.GetProfileByVAT(ctx, req.VATSiret)
	if err != nil {
		return nil, "", fmt.Errorf("checking VAT: %w", err)
	}
	if existing != nil {
		return nil, "", ErrCompanyAlreadyExists
	}

	// Check username uniqueness
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("checking username: %w", err)
	}
	if existingUser != nil {
		return nil, "", ErrUsernameExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	now := s.now()
	userID := uuid.New().String()

	user := &domain.User{
		ID:           userID,
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         domain.RoleBusiness,
		ContactEmail: req.BillingEmail,
		CreatedAt:    now,
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, "", fmt.Errorf("saving user: %w", err)
	}

	profile := &domain.BusinessProfile{
		ID:                 uuid.New().String(),
		UserID:             userID,
		CompanyName:        req.CompanyName,
		VATSiret:           req.VATSiret,
		IBAN:               req.IBAN,
		BillingEmail:       req.BillingEmail,
		BillingContactName: req.BillingContactName,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.b2bRepo.CreateProfile(ctx, profile); err != nil {
		return nil, "", fmt.Errorf("saving profile: %w", err)
	}

	// Generate JWT for the new user
	token, err := generateB2BJWT(userID, int(domain.RoleBusiness), s.jwtSecret, now)
	if err != nil {
		return nil, "", fmt.Errorf("generating token: %w", err)
	}

	return profile, token, nil
}

func (s *B2BServiceImpl) GetProfile(ctx context.Context, userID string) (*domain.BusinessProfile, error) {
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}
	return profile, nil
}

func (s *B2BServiceImpl) UpdateProfile(ctx context.Context, userID string, req domain.UpdateProfileRequest) (*domain.BusinessProfile, error) {
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	if req.CompanyName != "" {
		profile.CompanyName = req.CompanyName
	}
	if req.IBAN != "" {
		profile.IBAN = req.IBAN
	}
	if req.BillingEmail != "" {
		profile.BillingEmail = req.BillingEmail
	}
	if req.BillingContactName != "" {
		profile.BillingContactName = req.BillingContactName
	}
	profile.UpdatedAt = s.now()

	if err := s.b2bRepo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// --- Delivery Sites ---

func (s *B2BServiceImpl) CreateSite(ctx context.Context, userID string, site domain.DeliverySite) (*domain.DeliverySite, error) {
	now := s.now()
	site.ID = uuid.New().String()
	site.UserID = userID
	site.CreatedAt = now
	site.UpdatedAt = now
	if site.Country == "" {
		site.Country = "BE"
	}

	if err := s.b2bRepo.CreateSite(ctx, &site); err != nil {
		return nil, err
	}
	return &site, nil
}

func (s *B2BServiceImpl) ListSites(ctx context.Context, userID string) ([]domain.DeliverySite, error) {
	return s.b2bRepo.ListSitesByUser(ctx, userID)
}

func (s *B2BServiceImpl) UpdateSite(ctx context.Context, userID string, siteID string, site domain.DeliverySite) (*domain.DeliverySite, error) {
	existing, err := s.b2bRepo.GetSiteByID(ctx, siteID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrSiteNotFound
	}
	if existing.UserID != userID {
		return nil, ErrForbidden
	}

	existing.Name = site.Name
	existing.StreetAddress = site.StreetAddress
	existing.City = site.City
	existing.PostalCode = site.PostalCode
	if site.Country != "" {
		existing.Country = site.Country
	}
	existing.DeliveryInstructions = site.DeliveryInstructions
	existing.UpdatedAt = s.now()

	if err := s.b2bRepo.UpdateSite(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *B2BServiceImpl) DeleteSite(ctx context.Context, userID string, siteID string) error {
	existing, err := s.b2bRepo.GetSiteByID(ctx, siteID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrSiteNotFound
	}
	if existing.UserID != userID {
		return ErrForbidden
	}

	count, err := s.b2bRepo.CountSitesByUser(ctx, userID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastSite
	}

	return s.b2bRepo.DeleteSite(ctx, siteID)
}

// --- Access Management ---

func (s *B2BServiceImpl) RequestAccess(ctx context.Context, userID string, bakeryID string) (*domain.B2BAccess, error) {
	// Check if access already exists
	existing, err := s.b2bRepo.GetAccess(ctx, bakeryID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAccessExists
	}

	now := s.now()
	access := &domain.B2BAccess{
		ID:             uuid.New().String(),
		BakeryID:       bakeryID,
		BusinessUserID: userID,
		Status:         domain.B2BAccessPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.b2bRepo.CreateAccess(ctx, access); err != nil {
		return nil, err
	}
	return access, nil
}

func (s *B2BServiceImpl) ApproveAccess(ctx context.Context, accessID string) error {
	return s.b2bRepo.UpdateAccessStatus(ctx, accessID, domain.B2BAccessApproved)
}

func (s *B2BServiceImpl) RejectAccess(ctx context.Context, accessID string) error {
	return s.b2bRepo.UpdateAccessStatus(ctx, accessID, domain.B2BAccessRejected)
}

func (s *B2BServiceImpl) RevokeAccess(ctx context.Context, accessID string) error {
	return s.b2bRepo.UpdateAccessStatus(ctx, accessID, domain.B2BAccessRevoked)
}

func (s *B2BServiceImpl) ListAccessRequests(ctx context.Context, bakeryID string) ([]domain.B2BAccess, error) {
	return s.b2bRepo.ListAccessByBakery(ctx, bakeryID, nil)
}

func (s *B2BServiceImpl) ListApprovedBakeries(ctx context.Context, userID string) ([]domain.Bakery, error) {
	ids, err := s.b2bRepo.ListApprovedBakeryIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	var bakeries []domain.Bakery
	for _, id := range ids {
		b, err := s.bakeryRepo.GetBakery(ctx, id)
		if err != nil {
			return nil, err
		}
		if b != nil {
			bakeries = append(bakeries, *b)
		}
	}
	return bakeries, nil
}

func (s *B2BServiceImpl) HasApprovedAccess(ctx context.Context, userID string, bakeryID string) (bool, error) {
	access, err := s.b2bRepo.GetAccess(ctx, bakeryID, userID)
	if err != nil {
		return false, err
	}
	return access != nil && access.Status == domain.B2BAccessApproved, nil
}

// --- B2B Config ---

func (s *B2BServiceImpl) GetConfig(ctx context.Context, bakeryID string) (*domain.B2BConfig, error) {
	config, err := s.b2bRepo.GetConfig(ctx, bakeryID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		// Return default config
		return &domain.B2BConfig{
			BakeryID:            bakeryID,
			CutoffTime:          domain.TimeOfDay{Hour: 18, Minute: 0},
			DeliveryWindowStart: domain.TimeOfDay{Hour: 6, Minute: 0},
			DeliveryWindowEnd:   domain.TimeOfDay{Hour: 9, Minute: 0},
			OrderMinimum:        2000,
			ProDiscount:         0,
			VATRate:             6,
		}, nil
	}
	return config, nil
}

func (s *B2BServiceImpl) SaveConfig(ctx context.Context, bakeryID string, config domain.B2BConfig) (*domain.B2BConfig, error) {
	now := s.now()
	config.BakeryID = bakeryID
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = now
	config.UpdatedAt = now

	if err := s.b2bRepo.SaveConfig(ctx, &config); err != nil {
		return nil, err
	}

	return s.b2bRepo.GetConfig(ctx, bakeryID)
}

// --- Cart & Checkout ---

func (s *B2BServiceImpl) CheckoutBakeryGroup(ctx context.Context, userID string, req domain.CheckoutRequest) (*domain.Order, error) {
	// Verify access
	hasAccess, err := s.HasApprovedAccess(ctx, userID, req.BakeryID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrAccessDenied
	}

	// Verify delivery site exists and belongs to user
	site, err := s.b2bRepo.GetSiteByID(ctx, req.DeliverySiteID)
	if err != nil {
		return nil, err
	}
	if site == nil {
		return nil, ErrNoDeliverySite
	}
	if site.UserID != userID {
		return nil, ErrForbidden
	}

	// Get config for cutoff, minimum, and VAT rate
	config, err := s.GetConfig(ctx, req.BakeryID)
	if err != nil {
		return nil, err
	}

	// Check cutoff time
	now := s.now()
	currentTime := domain.TimeOfDay{Hour: now.Hour(), Minute: now.Minute()}
	if currentTime.AfterOrEqual(config.CutoffTime) {
		return nil, ErrCutoffPassed
	}

	// Get profile for pro discount and monthly spend tracking
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	// Compute pricing using the account-level pro discount and config VAT rate
	pricing := computePricing(req.Items, profile.ProDiscount, config.VATRate)

	// Check minimum
	if pricing.SubtotalHT < config.OrderMinimum {
		return nil, ErrBelowMinimum
	}

	// Create the order
	orderID := uuid.New().String()
	order := &domain.Order{
		ID:            orderID,
		BakeryID:      req.BakeryID,
		UserID:        userID,
		Items:         req.Items,
		Status:        domain.OrderStatusConfirmed,
		TotalAmount:   pricing.TotalTTC,
		PaymentMethod: domain.PaymentMethodOnInvoice,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("saving order: %w", err)
	}

	// Update monthly spend tracking
	currentMonth := now.Format("2006-01")
	newSpend := pricing.SubtotalHT
	if profile.SpendMonth == currentMonth {
		newSpend += profile.CurrentMonthSpend
	}
	if err := s.b2bRepo.UpdateMonthlySpend(ctx, profile.ID, newSpend, currentMonth); err != nil {
		// Non-fatal: log but don't fail the order
		_ = err
	}

	return order, nil
}

func (s *B2BServiceImpl) EditOrder(ctx context.Context, userID string, orderID string, req domain.EditOrderRequest) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}

	// Check cutoff
	config, err := s.GetConfig(ctx, order.BakeryID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	currentTime := domain.TimeOfDay{Hour: now.Hour(), Minute: now.Minute()}
	if currentTime.AfterOrEqual(config.CutoffTime) {
		return nil, ErrCutoffPassed
	}

	// Get profile for pro discount
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	// Recompute pricing with account-level pro discount and config VAT rate
	pricing := computePricing(req.Items, profile.ProDiscount, config.VATRate)

	// Check minimum
	if pricing.SubtotalHT < config.OrderMinimum {
		return nil, ErrBelowMinimum
	}

	order.Items = req.Items
	order.TotalAmount = pricing.TotalTTC
	order.UpdatedAt = now

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *B2BServiceImpl) ComputePricing(ctx context.Context, userID string, bakeryID string, items []domain.OrderItem) (*domain.B2BPricingResult, error) {
	config, err := s.GetConfig(ctx, bakeryID)
	if err != nil {
		return nil, err
	}

	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	tiers, err := s.b2bRepo.ListVolumeTiers(ctx)
	if err != nil {
		return nil, err
	}

	// Reset spend if month has changed
	currentMonth := s.now().Format("2006-01")
	monthlySpend := profile.CurrentMonthSpend
	if profile.SpendMonth != currentMonth {
		monthlySpend = 0
	}

	result := computeFullPricing(items, profile.ProDiscount, config.VATRate, monthlySpend, tiers)
	return &result, nil
}

// computePricing is a pure function that calculates B2B order pricing (legacy, used for simple total calculation).
func computePricing(items []domain.OrderItem, proDiscountRate int, vatRate int) domain.B2BOrderPricing {
	var subtotalHT int64
	for _, item := range items {
		subtotalHT += int64(item.Quantity) * item.UnitPrice
	}

	discountAmount := subtotalHT * int64(proDiscountRate) / 100
	taxableAmount := subtotalHT - discountAmount
	tvaAmount := taxableAmount * int64(vatRate) / 100
	totalTTC := taxableAmount + tvaAmount

	return domain.B2BOrderPricing{
		SubtotalHT:     subtotalHT,
		DiscountRate:   proDiscountRate,
		DiscountAmount: discountAmount,
		TVARate:        vatRate,
		TVAAmount:      tvaAmount,
		TotalTTC:       totalTTC,
	}
}

// computeFullPricing calculates B2B pricing with pro discount, volume discount, and VAT.
// Pricing flow: Subtotal HT -> Pro discount -> Volume discount -> TVA -> Total TTC
func computeFullPricing(items []domain.OrderItem, proDiscountRate int, vatRate int, monthlySpend int64, tiers []domain.VolumeTier) domain.B2BPricingResult {
	var subtotalHT int64
	for _, item := range items {
		subtotalHT += int64(item.Quantity) * item.UnitPrice
	}

	// Step 1: Pro discount on subtotal
	proDiscountAmt := subtotalHT * int64(proDiscountRate) / 100
	afterProDiscount := subtotalHT - proDiscountAmt

	// Step 2: Determine volume tier based on monthly spend
	var currentTier *domain.VolumeTier
	var nextTier *domain.VolumeTier
	for i := range tiers {
		if monthlySpend >= tiers[i].MinMonthlySpend {
			currentTier = &tiers[i]
		} else {
			nextTier = &tiers[i]
			break
		}
	}
	// If we matched all tiers, there's no next tier
	if currentTier != nil && nextTier == nil {
		// Check if the last matched tier is the last in the list
		for i := range tiers {
			if tiers[i].ID == currentTier.ID && i < len(tiers)-1 {
				nextTier = &tiers[i+1]
			}
		}
	}
	// If no tier matched, next tier is the first one
	if currentTier == nil && len(tiers) > 0 {
		nextTier = &tiers[0]
	}

	volDiscountRate := 0
	if currentTier != nil {
		volDiscountRate = currentTier.DiscountPercent
	}

	// Step 3: Volume discount on (subtotal - pro discount)
	volDiscountAmt := afterProDiscount * int64(volDiscountRate) / 100
	taxableAmount := afterProDiscount - volDiscountAmt

	// Step 4: TVA on discounted amount
	tvaAmount := taxableAmount * int64(vatRate) / 100
	totalTTC := taxableAmount + tvaAmount

	// Spend to next tier
	var spendToNextTier int64
	if nextTier != nil {
		spendToNextTier = nextTier.MinMonthlySpend - monthlySpend
		if spendToNextTier < 0 {
			spendToNextTier = 0
		}
	}

	return domain.B2BPricingResult{
		SubtotalHT:      subtotalHT,
		ProDiscountRate: proDiscountRate,
		ProDiscountAmt:  proDiscountAmt,
		VolDiscountRate: volDiscountRate,
		VolDiscountAmt:  volDiscountAmt,
		TVARate:         vatRate,
		TVAAmount:       tvaAmount,
		TotalTTC:        totalTTC,
		CurrentTier:     currentTier,
		NextTier:        nextTier,
		MonthlySpend:    monthlySpend,
		SpendToNextTier: spendToNextTier,
	}
}

// --- Saved Lists ---

func (s *B2BServiceImpl) CreateSavedList(ctx context.Context, userID string, list domain.SavedList) (*domain.SavedList, error) {
	now := s.now()
	list.ID = uuid.New().String()
	list.UserID = userID
	list.CreatedAt = now
	list.UpdatedAt = now

	for i := range list.Items {
		list.Items[i].ID = uuid.New().String()
	}

	if err := s.b2bRepo.CreateSavedList(ctx, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (s *B2BServiceImpl) ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]domain.SavedList, error) {
	return s.b2bRepo.ListSavedLists(ctx, userID, bakeryID)
}

func (s *B2BServiceImpl) DeleteSavedList(ctx context.Context, userID string, listID string) error {
	list, err := s.b2bRepo.GetSavedListByID(ctx, listID)
	if err != nil {
		return err
	}
	if list == nil {
		return fmt.Errorf("saved list not found")
	}
	if list.UserID != userID {
		return ErrForbidden
	}
	return s.b2bRepo.DeleteSavedList(ctx, listID)
}

// --- Deliveries & Invoices ---

func (s *B2BServiceImpl) ListDeliveries(ctx context.Context, userID string, filters domain.B2BOrderFilters, params domain.PaginationParams) (*domain.ListResult[domain.Order], error) {
	// Use the order repo with the on_invoice payment method filter
	orderFilters := domain.OrderFilters{}
	if filters.Status != "" {
		status := domain.OrderStatus(filters.Status)
		orderFilters.Status = &status
	}
	orderFilters.SortBy = "createdAt"
	orderFilters.SortDir = "desc"

	orders, total, err := s.orderRepo.ListByUser(ctx, userID, orderFilters, params)
	if err != nil {
		return nil, err
	}

	// Filter to only on_invoice orders
	var b2bOrders []domain.Order
	for _, o := range orders {
		if o.PaymentMethod == domain.PaymentMethodOnInvoice {
			b2bOrders = append(b2bOrders, o)
		}
	}

	return &domain.ListResult[domain.Order]{
		Items:    b2bOrders,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

func (s *B2BServiceImpl) GetLastOrder(ctx context.Context, userID string, bakeryID string) (*domain.Order, error) {
	filters := domain.OrderFilters{
		SortBy:  "createdAt",
		SortDir: "desc",
	}
	params := domain.PaginationParams{Page: 1, PageSize: 1}

	orders, _, err := s.orderRepo.ListByUser(ctx, userID, filters, params)
	if err != nil {
		return nil, err
	}

	// Find the most recent order for this bakery with on_invoice payment
	for _, o := range orders {
		if o.BakeryID == bakeryID && o.PaymentMethod == domain.PaymentMethodOnInvoice {
			return &o, nil
		}
	}
	return nil, nil
}

func (s *B2BServiceImpl) ListInvoices(ctx context.Context, userID string, params domain.PaginationParams) (*domain.ListResult[domain.B2BInvoice], error) {
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	invoices, total, err := s.b2bRepo.ListInvoicesByUser(ctx, profile.ID, params)
	if err != nil {
		return nil, err
	}

	return &domain.ListResult[domain.B2BInvoice]{
		Items:    invoices,
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}, nil
}

func (s *B2BServiceImpl) GenerateInvoice(ctx context.Context, orderID string) (*domain.B2BInvoice, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	// Check if invoice already exists
	existing, err := s.b2bRepo.GetInvoiceByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Get business profile
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	// Get config for VAT rate
	config, err := s.GetConfig(ctx, order.BakeryID)
	if err != nil {
		return nil, err
	}

	pricing := computePricing(order.Items, profile.ProDiscount, config.VATRate)

	// Get next invoice number
	num, err := s.b2bRepo.NextInvoiceNumber(ctx, order.BakeryID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	invoice := &domain.B2BInvoice{
		ID:                uuid.New().String(),
		OrderID:           orderID,
		BakeryID:          order.BakeryID,
		BusinessProfileID: profile.ID,
		InvoiceNumber:     num,
		SubtotalHT:        pricing.SubtotalHT,
		DiscountAmount:    pricing.DiscountAmount,
		TVAAmount:         pricing.TVAAmount,
		TotalTTC:          pricing.TotalTTC,
		PaymentStatus:     "pending",
		IssuedAt:          now,
	}

	if err := s.b2bRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *B2BServiceImpl) DownloadInvoicePDF(ctx context.Context, invoiceID string, userID string) ([]byte, error) {
	invoice, err := s.b2bRepo.GetInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, ErrInvoiceNotFound
	}

	// Verify ownership
	profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.ID != invoice.BusinessProfileID {
		return nil, ErrForbidden
	}

	// Placeholder PDF generation
	pdf := generatePlaceholderPDF(invoice, profile)
	return pdf, nil
}

// generatePlaceholderPDF creates a minimal placeholder PDF for the invoice.
func generatePlaceholderPDF(invoice *domain.B2BInvoice, profile *domain.BusinessProfile) []byte {
	content := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 200 >>
stream
BT
/F1 16 Tf
50 700 Td
(Invoice #%d) Tj
0 -30 Td
/F1 12 Tf
(Company: %s) Tj
0 -20 Td
(Total TTC: %d cents) Tj
0 -20 Td
(Status: %s) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
trailer
<< /Size 6 /Root 1 0 R >>
startxref
0
%%%%EOF`, invoice.InvoiceNumber, profile.CompanyName, invoice.TotalTTC, invoice.PaymentStatus)

	return []byte(content)
}

// generateJWT generates a JWT token for a user (used during B2B registration).
func generateB2BJWT(userID string, role int, secret string, now time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  jwt.NewNumericDate(now.Add(24 * time.Hour)),
		"iat":  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
