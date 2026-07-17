package domain

// CalculateOrderTotal computes the total for an order by calculating each item's
// subtotal (quantity × unitPrice) and summing them. It returns the total in cents
// and updates each item's Subtotal field in place.
func CalculateOrderTotal(items []OrderItem) int64 {
	var total int64
	for i := range items {
		items[i].Subtotal = int64(items[i].Quantity) * items[i].UnitPrice
		total += items[i].Subtotal
	}
	return total
}
