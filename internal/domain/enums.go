package domain

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusConfirmed      OrderStatus = "confirmed"
	OrderStatusPreparing      OrderStatus = "preparing"
	OrderStatusReady          OrderStatus = "ready"
	OrderStatusCapturing      OrderStatus = "capturing"
	OrderStatusDelivered      OrderStatus = "delivered"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

// ReservationStatus represents the lifecycle state of a reservation.
type ReservationStatus string

const (
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusReady     ReservationStatus = "ready"
	ReservationStatusPickedUp  ReservationStatus = "picked_up"
	ReservationStatusCancelled ReservationStatus = "cancelled"
)

// PaymentMethod represents how an order or reservation is paid for.
type PaymentMethod string

const (
	PaymentMethodOnline PaymentMethod = "online"
	PaymentMethodOnSpot PaymentMethod = "on_spot"
)

// DayOfWeek represents days of the week.
type DayOfWeek string

const (
	Monday    DayOfWeek = "monday"
	Tuesday   DayOfWeek = "tuesday"
	Wednesday DayOfWeek = "wednesday"
	Thursday  DayOfWeek = "thursday"
	Friday    DayOfWeek = "friday"
	Saturday  DayOfWeek = "saturday"
	Sunday    DayOfWeek = "sunday"
)

// AllDaysOfWeek returns all days of the week in order.
func AllDaysOfWeek() []DayOfWeek {
	return []DayOfWeek{Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday}
}
