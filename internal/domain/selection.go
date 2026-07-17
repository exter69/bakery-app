package domain

// AddProductToSelection adds a product to the selection list.
// If the product is already present (matched by productID), its quantity is
// incremented by 1. Otherwise a new OrderItem is appended with quantity 1.
func AddProductToSelection(items []OrderItem, productID string, productName string, unitPrice int64) []OrderItem {
	for i := range items {
		if items[i].ProductID == productID {
			items[i].Quantity++
			items[i].Subtotal = int64(items[i].Quantity) * items[i].UnitPrice
			return items
		}
	}
	return append(items, OrderItem{
		ProductID:   productID,
		ProductName: productName,
		Quantity:    1,
		UnitPrice:   unitPrice,
		Subtotal:    unitPrice,
	})
}
