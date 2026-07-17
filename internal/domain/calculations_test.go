package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateOrderTotal_SingleItem(t *testing.T) {
	items := []OrderItem{
		{ProductID: "p1", ProductName: "Croissant", Quantity: 3, UnitPrice: 350},
	}

	total := CalculateOrderTotal(items)

	assert.Equal(t, int64(1050), total)
	assert.Equal(t, int64(1050), items[0].Subtotal)
}

func TestCalculateOrderTotal_MultipleItems(t *testing.T) {
	items := []OrderItem{
		{ProductID: "p1", ProductName: "Croissant", Quantity: 3, UnitPrice: 350},
		{ProductID: "p2", ProductName: "Baguette", Quantity: 1, UnitPrice: 500},
		{ProductID: "p3", ProductName: "Eclair", Quantity: 6, UnitPrice: 275},
	}

	total := CalculateOrderTotal(items)

	// 3*350 = 1050, 1*500 = 500, 6*275 = 1650 → total = 3200
	assert.Equal(t, int64(3200), total)
	assert.Equal(t, int64(1050), items[0].Subtotal)
	assert.Equal(t, int64(500), items[1].Subtotal)
	assert.Equal(t, int64(1650), items[2].Subtotal)
}

func TestCalculateOrderTotal_EmptySlice(t *testing.T) {
	items := []OrderItem{}

	total := CalculateOrderTotal(items)

	assert.Equal(t, int64(0), total)
}

func TestCalculateOrderTotal_QuantityOne(t *testing.T) {
	items := []OrderItem{
		{ProductID: "p1", ProductName: "Birthday Cake", Quantity: 1, UnitPrice: 2500},
	}

	total := CalculateOrderTotal(items)

	assert.Equal(t, int64(2500), total)
	assert.Equal(t, int64(2500), items[0].Subtotal)
}

func TestCalculateOrderTotal_SetsSubtotalOnEachItem(t *testing.T) {
	items := []OrderItem{
		{ProductID: "p1", ProductName: "Item A", Quantity: 2, UnitPrice: 100, Subtotal: 0},
		{ProductID: "p2", ProductName: "Item B", Quantity: 5, UnitPrice: 200, Subtotal: 0},
	}

	total := CalculateOrderTotal(items)

	assert.Equal(t, int64(200), items[0].Subtotal)
	assert.Equal(t, int64(1000), items[1].Subtotal)
	assert.Equal(t, int64(1200), total)
}
