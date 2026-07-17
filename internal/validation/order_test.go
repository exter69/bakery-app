package validation

import (
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOrderItems_EmptyItems(t *testing.T) {
	result := ValidateOrderItems(nil)
	require.True(t, result.HasErrors())
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "items", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "at least one item")
}

func TestValidateOrderItems_ValidItems(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 100},
		{ProductID: "p2", Quantity: 999, UnitPrice: 5000},
	}
	result := ValidateOrderItems(items)
	assert.False(t, result.HasErrors())
}

func TestValidateOrderItems_QuantityTooLow(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 0, UnitPrice: 100},
	}
	result := ValidateOrderItems(items)
	require.True(t, result.HasErrors())
	assert.Equal(t, "items[0].quantity", result.Errors[0].Field)
}

func TestValidateOrderItems_QuantityTooHigh(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1000, UnitPrice: 100},
	}
	result := ValidateOrderItems(items)
	require.True(t, result.HasErrors())
	assert.Equal(t, "items[0].quantity", result.Errors[0].Field)
}

func TestValidateOrderItems_ZeroUnitPrice(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 0},
	}
	result := ValidateOrderItems(items)
	require.True(t, result.HasErrors())
	assert.Equal(t, "items[0].unitPrice", result.Errors[0].Field)
}

func TestValidateOrderItems_NegativeUnitPrice(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: -50},
	}
	result := ValidateOrderItems(items)
	require.True(t, result.HasErrors())
	assert.Equal(t, "items[0].unitPrice", result.Errors[0].Field)
}

func TestValidateOrderItems_MultipleErrors(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 0, UnitPrice: -1},
		{ProductID: "p2", Quantity: 5, UnitPrice: 200},
	}
	result := ValidateOrderItems(items)
	require.True(t, result.HasErrors())
	assert.Len(t, result.Errors, 2) // quantity and unitPrice errors for items[0]
}

func TestValidateSchedule_BakeryOpen_TimeWithinHours(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Monday,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: 8, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0},
		},
	}
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 10, Minute: 0},
		EndTime:   domain.TimeOfDay{Hour: 11, Minute: 0},
	}

	result := ValidateSchedule(domain.Monday, slot, schedule)
	assert.False(t, result.HasErrors())
}

func TestValidateSchedule_BakeryClosed(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Sunday,
			IsOpen:    false,
			OpenTime:  domain.TimeOfDay{Hour: 0, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 0, Minute: 0},
		},
	}
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 10, Minute: 0},
		EndTime:   domain.TimeOfDay{Hour: 11, Minute: 0},
	}

	result := ValidateSchedule(domain.Sunday, slot, schedule)
	require.True(t, result.HasErrors())
	assert.Equal(t, "scheduledDay", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "closed")
}

func TestValidateSchedule_StartTimeBeforeOpening(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Tuesday,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: 9, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 17, Minute: 0},
		},
	}
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 7, Minute: 30},
		EndTime:   domain.TimeOfDay{Hour: 10, Minute: 0},
	}

	result := ValidateSchedule(domain.Tuesday, slot, schedule)
	require.True(t, result.HasErrors())
	assert.Equal(t, "scheduledTime.startTime", result.Errors[0].Field)
}

func TestValidateSchedule_EndTimeAfterClosing(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Wednesday,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: 8, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 16, Minute: 0},
		},
	}
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 15, Minute: 0},
		EndTime:   domain.TimeOfDay{Hour: 17, Minute: 30},
	}

	result := ValidateSchedule(domain.Wednesday, slot, schedule)
	require.True(t, result.HasErrors())
	assert.Equal(t, "scheduledTime.endTime", result.Errors[0].Field)
}

func TestValidateSchedule_ExactBoundary(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Friday,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: 8, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0},
		},
	}
	// Slot exactly at opening and closing is valid.
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 8, Minute: 0},
		EndTime:   domain.TimeOfDay{Hour: 18, Minute: 0},
	}

	result := ValidateSchedule(domain.Friday, slot, schedule)
	assert.False(t, result.HasErrors())
}

func TestValidateSchedule_DayNotInSchedule(t *testing.T) {
	schedule := []domain.DaySchedule{
		{
			Day:       domain.Monday,
			IsOpen:    true,
			OpenTime:  domain.TimeOfDay{Hour: 8, Minute: 0},
			CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0},
		},
	}
	slot := domain.TimeSlot{
		StartTime: domain.TimeOfDay{Hour: 10, Minute: 0},
		EndTime:   domain.TimeOfDay{Hour: 11, Minute: 0},
	}

	result := ValidateSchedule(domain.Saturday, slot, schedule)
	require.True(t, result.HasErrors())
	assert.Equal(t, "scheduledDay", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "no schedule found")
}

func TestValidateProductAvailability_AllAvailable(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 2, UnitPrice: 300},
		{ProductID: "p2", Quantity: 1, UnitPrice: 500},
	}
	products := []domain.Product{
		{ID: "p1", Name: "Croissant", IsAvailable: true},
		{ID: "p2", Name: "Baguette", IsAvailable: true},
	}

	result := ValidateProductAvailability(items, products)
	assert.False(t, result.HasErrors())
}

func TestValidateProductAvailability_ProductNotFound(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p-unknown", Quantity: 1, UnitPrice: 100},
	}
	products := []domain.Product{
		{ID: "p1", Name: "Croissant", IsAvailable: true},
	}

	result := ValidateProductAvailability(items, products)
	require.True(t, result.HasErrors())
	assert.Contains(t, result.Errors[0].Message, "not found")
}

func TestValidateProductAvailability_ProductUnavailable(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 300},
	}
	products := []domain.Product{
		{ID: "p1", Name: "Croissant", IsAvailable: false},
	}

	result := ValidateProductAvailability(items, products)
	require.True(t, result.HasErrors())
	assert.Contains(t, result.Errors[0].Message, "not available")
}

func TestValidateProductAvailability_MixedAvailability(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 300},
		{ProductID: "p2", Quantity: 2, UnitPrice: 500},
		{ProductID: "p3", Quantity: 1, UnitPrice: 200},
	}
	products := []domain.Product{
		{ID: "p1", Name: "Croissant", IsAvailable: true},
		{ID: "p2", Name: "Baguette", IsAvailable: false},
		{ID: "p3", Name: "Eclair", IsAvailable: true},
	}

	result := ValidateProductAvailability(items, products)
	require.True(t, result.HasErrors())
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "Baguette")
}

func TestValidationResult_Merge(t *testing.T) {
	r1 := ValidationResult{
		Errors: []ValidationError{{Field: "a", Message: "err1"}},
	}
	r2 := ValidationResult{
		Errors: []ValidationError{{Field: "b", Message: "err2"}},
	}
	r1.Merge(r2)
	assert.Len(t, r1.Errors, 2)
}
