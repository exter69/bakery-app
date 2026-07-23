package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransitionOrder_ValidTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      OrderStatus
		to        OrderStatus
	}{
		{"PendingPayment to Confirmed", OrderStatusPendingPayment, OrderStatusConfirmed},
		{"PendingPayment to Cancelled", OrderStatusPendingPayment, OrderStatusCancelled},
		{"Confirmed to Preparing", OrderStatusConfirmed, OrderStatusPreparing},
		{"Confirmed to Cancelled", OrderStatusConfirmed, OrderStatusCancelled},
		{"Preparing to Ready", OrderStatusPreparing, OrderStatusReady},
		{"Preparing to Cancelled", OrderStatusPreparing, OrderStatusCancelled},
		{"Ready to Delivered", OrderStatusReady, OrderStatusDelivered},
		{"Ready to Capturing", OrderStatusReady, OrderStatusCapturing},
		{"Capturing to Delivered", OrderStatusCapturing, OrderStatusDelivered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Status: tt.from}
			err := TransitionOrder(order, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.to, order.Status)
		})
	}
}

func TestTransitionOrder_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from OrderStatus
		to   OrderStatus
	}{
		{"PendingPayment to Preparing", OrderStatusPendingPayment, OrderStatusPreparing},
		{"PendingPayment to Ready", OrderStatusPendingPayment, OrderStatusReady},
		{"PendingPayment to Delivered", OrderStatusPendingPayment, OrderStatusDelivered},
		{"Confirmed to PendingPayment", OrderStatusConfirmed, OrderStatusPendingPayment},
		{"Confirmed to Ready", OrderStatusConfirmed, OrderStatusReady},
		{"Confirmed to Delivered", OrderStatusConfirmed, OrderStatusDelivered},
		{"Preparing to PendingPayment", OrderStatusPreparing, OrderStatusPendingPayment},
		{"Preparing to Confirmed", OrderStatusPreparing, OrderStatusConfirmed},
		{"Preparing to Delivered", OrderStatusPreparing, OrderStatusDelivered},
		{"Ready to PendingPayment", OrderStatusReady, OrderStatusPendingPayment},
		{"Ready to Confirmed", OrderStatusReady, OrderStatusConfirmed},
		{"Ready to Preparing", OrderStatusReady, OrderStatusPreparing},
		{"Ready to Cancelled", OrderStatusReady, OrderStatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Status: tt.from}
			err := TransitionOrder(order, tt.to)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidTransition))
			// Status should remain unchanged
			assert.Equal(t, tt.from, order.Status)
		})
	}
}

func TestTransitionOrder_TerminalStatesRejectAll(t *testing.T) {
	allStatuses := []OrderStatus{
		OrderStatusPendingPayment,
		OrderStatusConfirmed,
		OrderStatusPreparing,
		OrderStatusReady,
		OrderStatusCapturing,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}

	terminalStatuses := []OrderStatus{OrderStatusDelivered, OrderStatusCancelled}

	for _, terminal := range terminalStatuses {
		for _, target := range allStatuses {
			t.Run(string(terminal)+" to "+string(target), func(t *testing.T) {
				order := &Order{Status: terminal}
				err := TransitionOrder(order, target)
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidTransition))
				assert.Equal(t, terminal, order.Status)
			})
		}
	}
}

func TestTransitionOrder_ErrorMessageDescriptive(t *testing.T) {
	order := &Order{Status: OrderStatusReady}
	err := TransitionOrder(order, OrderStatusPendingPayment)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ready")
	assert.Contains(t, err.Error(), "pending_payment")
}

func TestTransitionReservation_ValidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ReservationStatus
		to   ReservationStatus
	}{
		{"Confirmed to Ready", ReservationStatusConfirmed, ReservationStatusReady},
		{"Confirmed to Cancelled", ReservationStatusConfirmed, ReservationStatusCancelled},
		{"Ready to PickedUp", ReservationStatusReady, ReservationStatusPickedUp},
		{"Ready to Cancelled", ReservationStatusReady, ReservationStatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation := &Reservation{Status: tt.from}
			err := TransitionReservation(reservation, tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.to, reservation.Status)
		})
	}
}

func TestTransitionReservation_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ReservationStatus
		to   ReservationStatus
	}{
		{"Confirmed to PickedUp", ReservationStatusConfirmed, ReservationStatusPickedUp},
		{"Ready to Confirmed", ReservationStatusReady, ReservationStatusConfirmed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation := &Reservation{Status: tt.from}
			err := TransitionReservation(reservation, tt.to)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidTransition))
			assert.Equal(t, tt.from, reservation.Status)
		})
	}
}

func TestTransitionReservation_TerminalStatesRejectAll(t *testing.T) {
	allStatuses := []ReservationStatus{
		ReservationStatusConfirmed,
		ReservationStatusReady,
		ReservationStatusPickedUp,
		ReservationStatusCancelled,
	}

	terminalStatuses := []ReservationStatus{ReservationStatusPickedUp, ReservationStatusCancelled}

	for _, terminal := range terminalStatuses {
		for _, target := range allStatuses {
			t.Run(string(terminal)+" to "+string(target), func(t *testing.T) {
				reservation := &Reservation{Status: terminal}
				err := TransitionReservation(reservation, target)
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidTransition))
				assert.Equal(t, terminal, reservation.Status)
			})
		}
	}
}

func TestTransitionReservation_ErrorMessageDescriptive(t *testing.T) {
	reservation := &Reservation{Status: ReservationStatusCancelled}
	err := TransitionReservation(reservation, ReservationStatusReady)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.Contains(t, err.Error(), "ready")
}

func TestIsTerminalOrderStatus(t *testing.T) {
	assert.True(t, IsTerminalOrderStatus(OrderStatusDelivered))
	assert.True(t, IsTerminalOrderStatus(OrderStatusCancelled))
	assert.False(t, IsTerminalOrderStatus(OrderStatusPendingPayment))
	assert.False(t, IsTerminalOrderStatus(OrderStatusConfirmed))
	assert.False(t, IsTerminalOrderStatus(OrderStatusPreparing))
	assert.False(t, IsTerminalOrderStatus(OrderStatusReady))
	assert.False(t, IsTerminalOrderStatus(OrderStatusCapturing))
}

func TestIsTerminalReservationStatus(t *testing.T) {
	assert.True(t, IsTerminalReservationStatus(ReservationStatusPickedUp))
	assert.True(t, IsTerminalReservationStatus(ReservationStatusCancelled))
	assert.False(t, IsTerminalReservationStatus(ReservationStatusConfirmed))
	assert.False(t, IsTerminalReservationStatus(ReservationStatusReady))
}
