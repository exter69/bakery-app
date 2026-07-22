package invoice

import (
	"strings"
	"testing"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_ProducesValidHTMLWithLineItems(t *testing.T) {
	// Arrange
	data := InvoiceData{
		InvoiceNumber: "INV-order123-1700000000",
		OrderID:       "order123",
		Date:          time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		CustomerName:  "Jan Janssen",
		CustomerEmail: "jan@example.com",
		BakeryName:    "Boulangerie Belle",
		BakeryAddress: "Rue du Pain 42, Brussels",
		Items: []domain.OrderItem{
			{ProductID: "p1", ProductName: "Croissant", Quantity: 3, UnitPrice: 250, Subtotal: 750},
			{ProductID: "p2", ProductName: "Pain au chocolat", Quantity: 2, UnitPrice: 300, Subtotal: 600},
		},
		TotalCents: 1350,
	}

	// Act
	html, err := Generate(data)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, html, "INV-order123-1700000000")
	assert.Contains(t, html, "order123")
	assert.Contains(t, html, "2024-01-15")
	assert.Contains(t, html, "Jan Janssen")
	assert.Contains(t, html, "jan@example.com")
	assert.Contains(t, html, "Boulangerie Belle")
	assert.Contains(t, html, "Rue du Pain 42, Brussels")
	assert.Contains(t, html, "Croissant")
	assert.Contains(t, html, "Pain au chocolat")
	assert.Contains(t, html, "€2.50")  // unit price croissant
	assert.Contains(t, html, "€7.50")  // subtotal croissant
	assert.Contains(t, html, "€3.00")  // unit price pain au chocolat
	assert.Contains(t, html, "€6.00")  // subtotal pain au chocolat
	assert.Contains(t, html, "€13.50") // total
	assert.Contains(t, html, "Payment Confirmed")
	assert.True(t, strings.HasPrefix(html, "<!DOCTYPE html>"))
}

func TestGenerate_HandlesEmptyItems(t *testing.T) {
	// Arrange
	data := InvoiceData{
		InvoiceNumber: "INV-empty-1700000000",
		OrderID:       "empty",
		Date:          time.Now(),
		CustomerName:  "Test User",
		CustomerEmail: "test@example.com",
		BakeryName:    "Test Bakery",
		BakeryAddress: "Test Address",
		Items:         []domain.OrderItem{},
		TotalCents:    0,
	}

	// Act
	html, err := Generate(data)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, html, "€0.00")
	assert.Contains(t, html, "INV-empty-1700000000")
}

func TestGenerateInvoiceNumber_Format(t *testing.T) {
	// Arrange
	orderID := "abc123"
	ts := time.Unix(1700000000, 0)

	// Act
	num := GenerateInvoiceNumber(orderID, ts)

	// Assert
	assert.Equal(t, "INV-abc123-1700000000", num)
}

func TestGenerateInvoiceNumber_DifferentTimestampsProduceDifferentNumbers(t *testing.T) {
	// Arrange
	orderID := "order1"
	t1 := time.Unix(1000, 0)
	t2 := time.Unix(2000, 0)

	// Act
	num1 := GenerateInvoiceNumber(orderID, t1)
	num2 := GenerateInvoiceNumber(orderID, t2)

	// Assert
	assert.NotEqual(t, num1, num2)
}
