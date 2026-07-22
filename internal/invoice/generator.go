package invoice

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

// InvoiceData holds all data needed to render an invoice.
type InvoiceData struct {
	InvoiceNumber string
	OrderID       string
	Date          time.Time
	CustomerName  string
	CustomerEmail string
	BakeryName    string
	BakeryAddress string
	Items         []domain.OrderItem
	TotalCents    int64
}

// GenerateInvoiceNumber creates a unique invoice number from the order ID and timestamp.
func GenerateInvoiceNumber(orderID string, t time.Time) string {
	return fmt.Sprintf("INV-%s-%d", orderID, t.Unix())
}

// Generate creates an HTML invoice string from the provided data.
func Generate(data InvoiceData) (string, error) {
	tmpl, err := template.New("invoice").Funcs(template.FuncMap{
		"formatCents": func(cents int64) string {
			return fmt.Sprintf("€%.2f", float64(cents)/100)
		},
	}).Parse(invoiceTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing invoice template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing invoice template: %w", err)
	}
	return buf.String(), nil
}

const invoiceTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Invoice {{.InvoiceNumber}}</title>
<style>
  body { font-family: Arial, sans-serif; margin: 40px; color: #333; }
  .header { display: flex; justify-content: space-between; margin-bottom: 40px; }
  .invoice-title { font-size: 24px; font-weight: bold; color: #2c3e50; }
  .meta { margin-bottom: 30px; }
  .meta p { margin: 4px 0; }
  table { width: 100%; border-collapse: collapse; margin: 20px 0; }
  th { background: #f8f9fa; text-align: left; padding: 10px; border-bottom: 2px solid #dee2e6; }
  td { padding: 10px; border-bottom: 1px solid #eee; }
  .total-row { font-weight: bold; font-size: 1.1em; }
  .total-row td { border-top: 2px solid #333; padding-top: 12px; }
  .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; color: #666; font-size: 0.9em; }
  .confirmation { background: #d4edda; border: 1px solid #c3e6cb; padding: 12px; border-radius: 4px; margin-top: 20px; }
</style>
</head>
<body>
  <div class="header">
    <div>
      <div class="invoice-title">INVOICE</div>
      <p><strong>{{.InvoiceNumber}}</strong></p>
    </div>
    <div>
      <p><strong>{{.BakeryName}}</strong></p>
      <p>{{.BakeryAddress}}</p>
    </div>
  </div>

  <div class="meta">
    <p><strong>Date:</strong> {{.Date.Format "2006-01-02"}}</p>
    <p><strong>Order ID:</strong> {{.OrderID}}</p>
    <p><strong>Customer:</strong> {{.CustomerName}}</p>
    <p><strong>Email:</strong> {{.CustomerEmail}}</p>
  </div>

  <table>
    <thead>
      <tr>
        <th>Product</th>
        <th>Qty</th>
        <th>Unit Price</th>
        <th>Subtotal</th>
      </tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td>{{.ProductName}}</td>
        <td>{{.Quantity}}</td>
        <td>{{formatCents .UnitPrice}}</td>
        <td>{{formatCents .Subtotal}}</td>
      </tr>
      {{end}}
      <tr class="total-row">
        <td colspan="3">Total</td>
        <td>{{formatCents .TotalCents}}</td>
      </tr>
    </tbody>
  </table>

  <div class="confirmation">
    <strong>✓ Payment Confirmed</strong> — Thank you for your order! Your payment has been received and your order is being prepared.
  </div>

  <div class="footer">
    <p>This invoice was generated automatically. If you have questions, please contact {{.BakeryName}}.</p>
  </div>
</body>
</html>`
