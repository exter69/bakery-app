package postgres

import "github.com/lucatorrekens/bakery-app/internal/domain"

// Compile-time interface satisfaction checks.
var (
	_ domain.UserRepository              = (*UserRepo)(nil)
	_ domain.BakeryRepository            = (*BakeryRepo)(nil)
	_ domain.OrderRepository             = (*OrderRepo)(nil)
	_ domain.ReservationRepository       = (*ReservationRepo)(nil)
	_ domain.RecurringOrderRepository    = (*RecurringOrderRepo)(nil)
	_ domain.RegistrationTokenRepository = (*TokenRepo)(nil)
	_ domain.BundleRepository            = (*BundleRepo)(nil)
)
