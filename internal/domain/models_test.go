package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeOfDay_String(t *testing.T) {
	tests := []struct {
		time     TimeOfDay
		expected string
	}{
		{TimeOfDay{Hour: 9, Minute: 0}, "09:00"},
		{TimeOfDay{Hour: 14, Minute: 30}, "14:30"},
		{TimeOfDay{Hour: 0, Minute: 0}, "00:00"},
		{TimeOfDay{Hour: 23, Minute: 59}, "23:59"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.time.String())
	}
}

func TestTimeOfDay_Comparisons(t *testing.T) {
	early := TimeOfDay{Hour: 8, Minute: 0}
	mid := TimeOfDay{Hour: 12, Minute: 30}
	late := TimeOfDay{Hour: 18, Minute: 0}

	assert.True(t, early.Before(mid))
	assert.True(t, mid.Before(late))
	assert.False(t, late.Before(early))
	assert.False(t, mid.Before(mid))

	assert.True(t, late.After(mid))
	assert.True(t, mid.After(early))
	assert.False(t, early.After(late))
	assert.False(t, mid.After(mid))

	assert.True(t, mid.Equal(TimeOfDay{Hour: 12, Minute: 30}))
	assert.False(t, mid.Equal(early))

	assert.True(t, mid.BeforeOrEqual(mid))
	assert.True(t, early.BeforeOrEqual(mid))
	assert.False(t, late.BeforeOrEqual(mid))

	assert.True(t, mid.AfterOrEqual(mid))
	assert.True(t, late.AfterOrEqual(mid))
	assert.False(t, early.AfterOrEqual(mid))
}

func TestTimeOfDay_Before_SameHour(t *testing.T) {
	a := TimeOfDay{Hour: 10, Minute: 15}
	b := TimeOfDay{Hour: 10, Minute: 45}

	assert.True(t, a.Before(b))
	assert.False(t, b.Before(a))
}

func TestOrder_JSONRoundTrip(t *testing.T) {
	order := Order{
		ID:       "order-123",
		BakeryID: "bakery-456",
		UserID:   "user-789",
		Items: []OrderItem{
			{
				ProductID:   "prod-1",
				ProductName: "Croissant",
				Quantity:    3,
				UnitPrice:   350,
				Subtotal:    1050,
			},
		},
		ScheduledDay:  Wednesday,
		ScheduledTime: TimeSlot{StartTime: TimeOfDay{Hour: 10, Minute: 0}, EndTime: TimeOfDay{Hour: 10, Minute: 30}},
		Status:        OrderStatusPendingPayment,
		TotalAmount:   1050,
		PaymentMethod: PaymentMethodOnline,
		CreatedAt:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(order)
	require.NoError(t, err)

	var decoded Order
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, order.ID, decoded.ID)
	assert.Equal(t, order.Status, decoded.Status)
	assert.Equal(t, order.PaymentMethod, decoded.PaymentMethod)
	assert.Equal(t, order.ScheduledDay, decoded.ScheduledDay)
	assert.Len(t, decoded.Items, 1)
	assert.Equal(t, 3, decoded.Items[0].Quantity)
	assert.Equal(t, int64(350), decoded.Items[0].UnitPrice)
	assert.Equal(t, int64(1050), decoded.Items[0].Subtotal)
}

func TestReservation_JSONRoundTrip(t *testing.T) {
	reservation := Reservation{
		ID:       "res-123",
		BakeryID: "bakery-456",
		UserID:   "user-789",
		Items: []OrderItem{
			{
				ProductID:   "prod-2",
				ProductName: "Birthday Cake",
				Quantity:    1,
				UnitPrice:   2500,
				Subtotal:    2500,
			},
		},
		ScheduledDay:  Friday,
		ScheduledTime: TimeSlot{StartTime: TimeOfDay{Hour: 14, Minute: 0}, EndTime: TimeOfDay{Hour: 14, Minute: 30}},
		Status:        ReservationStatusConfirmed,
		TotalAmount:   2500,
		PaymentMethod: PaymentMethodOnSpot,
		CreatedAt:     time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(reservation)
	require.NoError(t, err)

	var decoded Reservation
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, reservation.ID, decoded.ID)
	assert.Equal(t, ReservationStatusConfirmed, decoded.Status)
	assert.Equal(t, PaymentMethodOnSpot, decoded.PaymentMethod)
	assert.Equal(t, Friday, decoded.ScheduledDay)
}

func TestBakery_JSONRoundTrip(t *testing.T) {
	bakery := Bakery{
		ID:          "bakery-1",
		Name:        "La Boulangerie",
		PhotoURL:    "https://example.com/photo.jpg",
		Description: "Traditional French bakery",
		Address:     "123 Main St",
		Schedule: []DaySchedule{
			{Day: Monday, OpenTime: TimeOfDay{Hour: 7, Minute: 0}, CloseTime: TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: Tuesday, OpenTime: TimeOfDay{Hour: 7, Minute: 0}, CloseTime: TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: Wednesday, OpenTime: TimeOfDay{Hour: 7, Minute: 0}, CloseTime: TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: Thursday, OpenTime: TimeOfDay{Hour: 7, Minute: 0}, CloseTime: TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: Friday, OpenTime: TimeOfDay{Hour: 7, Minute: 0}, CloseTime: TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: Saturday, OpenTime: TimeOfDay{Hour: 8, Minute: 0}, CloseTime: TimeOfDay{Hour: 14, Minute: 0}, IsOpen: true},
			{Day: Sunday, OpenTime: TimeOfDay{}, CloseTime: TimeOfDay{}, IsOpen: false},
		},
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(bakery)
	require.NoError(t, err)

	var decoded Bakery
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, bakery.Name, decoded.Name)
	assert.Len(t, decoded.Schedule, 7)
	assert.Equal(t, Monday, decoded.Schedule[0].Day)
	assert.True(t, decoded.Schedule[0].IsOpen)
	assert.False(t, decoded.Schedule[6].IsOpen)
}

func TestAllDaysOfWeek(t *testing.T) {
	days := AllDaysOfWeek()
	assert.Len(t, days, 7)
	assert.Equal(t, Monday, days[0])
	assert.Equal(t, Sunday, days[6])
}
