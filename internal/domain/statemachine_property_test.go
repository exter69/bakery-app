package domain

import (
	"errors"
	"testing"

	"pgregory.net/rapid"
)

// allOrderStatuses lists every defined OrderStatus value.
var allOrderStatuses = []OrderStatus{
	OrderStatusPendingPayment,
	OrderStatusConfirmed,
	OrderStatusPreparing,
	OrderStatusReady,
	OrderStatusDelivered,
	OrderStatusCancelled,
}

// terminalOrderStatuses lists the terminal states from which no transitions are allowed.
var terminalOrderStatuses = []OrderStatus{
	OrderStatusDelivered,
	OrderStatusCancelled,
}

// isValidTransition returns true if transitioning from `from` to `to` is permitted.
func isValidTransition(from, to OrderStatus) bool {
	allowed, ok := validOrderTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// orderStatusGen returns a rapid generator that draws a random OrderStatus.
func orderStatusGen() *rapid.Generator[OrderStatus] {
	return rapid.SampledFrom(allOrderStatuses)
}

// TestOrderStateTransitions_ValidTransitionsSucceed verifies that for any valid
// (from, to) pair in the transition map, TransitionOrder succeeds and the order
// status is updated to the target state.
// **Validates: Requirements 8.1, 8.2**
func TestOrderStateTransitions_ValidTransitionsSucceed(t *testing.T) {
	// Build list of all valid (from, to) pairs
	type transition struct {
		from OrderStatus
		to   OrderStatus
	}
	var validPairs []transition
	for from, targets := range validOrderTransitions {
		for _, to := range targets {
			validPairs = append(validPairs, transition{from: from, to: to})
		}
	}

	rapid.Check(t, func(t *rapid.T) {
		pair := rapid.SampledFrom(validPairs).Draw(t, "transition")

		order := &Order{Status: pair.from}
		err := TransitionOrder(order, pair.to)

		if err != nil {
			t.Fatalf("expected valid transition from %q to %q to succeed, got error: %v",
				pair.from, pair.to, err)
		}
		if order.Status != pair.to {
			t.Fatalf("after valid transition from %q to %q, order status is %q, want %q",
				pair.from, pair.to, order.Status, pair.to)
		}
	})
}

// TestOrderStateTransitions_InvalidTransitionsRejected verifies that for any
// (from, to) pair that is NOT in the valid transition map, TransitionOrder
// returns ErrInvalidTransition and the order status remains unchanged.
// **Validates: Requirements 8.1, 8.2**
func TestOrderStateTransitions_InvalidTransitionsRejected(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		from := orderStatusGen().Draw(t, "fromStatus")
		to := orderStatusGen().Draw(t, "toStatus")

		// Skip if this happens to be a valid transition
		if isValidTransition(from, to) {
			t.Skip("drawn pair is a valid transition, skipping")
		}

		order := &Order{Status: from}
		err := TransitionOrder(order, to)

		if err == nil {
			t.Fatalf("expected invalid transition from %q to %q to fail, but got nil error",
				from, to)
		}
		if !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("expected ErrInvalidTransition for %q to %q, got: %v",
				from, to, err)
		}
		if order.Status != from {
			t.Fatalf("after rejected transition from %q to %q, status changed to %q, want %q",
				from, to, order.Status, from)
		}
	})
}

// TestOrderStateTransitions_TerminalStatesRejectAll verifies that for terminal
// states (Delivered, Cancelled) all transition attempts are rejected and the
// order status remains unchanged.
// **Validates: Requirements 8.1, 8.2**
func TestOrderStateTransitions_TerminalStatesRejectAll(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		terminalStatus := rapid.SampledFrom(terminalOrderStatuses).Draw(t, "terminalStatus")
		targetStatus := orderStatusGen().Draw(t, "targetStatus")

		order := &Order{Status: terminalStatus}
		err := TransitionOrder(order, targetStatus)

		if err == nil {
			t.Fatalf("expected transition from terminal state %q to %q to fail, but got nil error",
				terminalStatus, targetStatus)
		}
		if !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("expected ErrInvalidTransition from terminal state %q to %q, got: %v",
				terminalStatus, targetStatus, err)
		}
		if order.Status != terminalStatus {
			t.Fatalf("after rejected transition from terminal %q to %q, status changed to %q",
				terminalStatus, targetStatus, order.Status)
		}
	})
}
