package domain

import "fmt"

// ErrInvalidTransition is returned when an invalid state transition is attempted.
var ErrInvalidTransition = fmt.Errorf("invalid state transition")

// validOrderTransitions defines the allowed state transitions for orders.
// Terminal states (Delivered, Cancelled) have no outgoing transitions.
var validOrderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPendingPayment: {OrderStatusConfirmed, OrderStatusCancelled},
	OrderStatusConfirmed:      {OrderStatusPreparing, OrderStatusCancelled},
	OrderStatusPreparing:      {OrderStatusReady, OrderStatusCancelled},
	OrderStatusReady:          {OrderStatusDelivered},
	OrderStatusDelivered:      {},
	OrderStatusCancelled:      {},
}

// validReservationTransitions defines the allowed state transitions for reservations.
// Terminal states (PickedUp, Cancelled) have no outgoing transitions.
var validReservationTransitions = map[ReservationStatus][]ReservationStatus{
	ReservationStatusConfirmed: {ReservationStatusReady, ReservationStatusCancelled},
	ReservationStatusReady:     {ReservationStatusPickedUp, ReservationStatusCancelled},
	ReservationStatusPickedUp:  {},
	ReservationStatusCancelled: {},
}

// TransitionOrder attempts to transition an order to a new status.
// It returns nil if the transition is valid, or an error describing why it was rejected.
func TransitionOrder(order *Order, newStatus OrderStatus) error {
	allowed, ok := validOrderTransitions[order.Status]
	if !ok {
		return fmt.Errorf("%w: unknown current order status %q", ErrInvalidTransition, order.Status)
	}

	for _, s := range allowed {
		if s == newStatus {
			order.Status = newStatus
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition order from %q to %q", ErrInvalidTransition, order.Status, newStatus)
}

// TransitionReservation attempts to transition a reservation to a new status.
// It returns nil if the transition is valid, or an error describing why it was rejected.
func TransitionReservation(reservation *Reservation, newStatus ReservationStatus) error {
	allowed, ok := validReservationTransitions[reservation.Status]
	if !ok {
		return fmt.Errorf("%w: unknown current reservation status %q", ErrInvalidTransition, reservation.Status)
	}

	for _, s := range allowed {
		if s == newStatus {
			reservation.Status = newStatus
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition reservation from %q to %q", ErrInvalidTransition, reservation.Status, newStatus)
}

// IsTerminalOrderStatus returns true if the given order status is terminal
// (no further transitions are permitted).
func IsTerminalOrderStatus(status OrderStatus) bool {
	return status == OrderStatusDelivered || status == OrderStatusCancelled
}

// IsTerminalReservationStatus returns true if the given reservation status is terminal
// (no further transitions are permitted).
func IsTerminalReservationStatus(status ReservationStatus) bool {
	return status == ReservationStatusPickedUp || status == ReservationStatusCancelled
}
