package notification

import (
	"bytes"
	"fmt"
	"html/template"
)

// Locale represents a supported language.
type Locale string

const (
	LocaleEN Locale = "en"
	LocaleFR Locale = "fr"
	LocaleNL Locale = "nl"
)

// emailTemplate holds the subject and body templates for a notification.
type emailTemplate struct {
	Subject string
	Body    string
}

// TemplateData holds the data available to all notification templates.
type TemplateData struct {
	BakeryName   string
	CustomerName string
	OrderID      string
	Items        []ItemData
	TotalDisplay string
}

// ItemData holds product info for template rendering.
type ItemData struct {
	ProductName string
	Quantity    int
	Subtotal    string
}

// templates maps locale → event type → template.
var templates = map[Locale]map[string]emailTemplate{
	LocaleEN: {
		"order_confirmed": {
			Subject: "Order Confirmed — {{.BakeryName}}",
			Body:    orderConfirmedBodyEN,
		},
		"status_preparing": {
			Subject: "Your order is being prepared — {{.BakeryName}}",
			Body:    statusPreparingBodyEN,
		},
		"status_ready": {
			Subject: "Your order is ready for delivery — {{.BakeryName}}",
			Body:    statusReadyBodyEN,
		},
		"status_delivered": {
			Subject: "Order delivered — {{.BakeryName}}",
			Body:    statusDeliveredBodyEN,
		},
		"new_order_baker": {
			Subject: "New order received! — {{.CustomerName}}",
			Body:    newOrderBakerBodyEN,
		},
		"reservation_confirmed": {
			Subject: "Reservation confirmed — {{.BakeryName}}",
			Body:    reservationConfirmedBodyEN,
		},
	},
	LocaleFR: {
		"order_confirmed": {
			Subject: "Commande confirmée — {{.BakeryName}}",
			Body:    orderConfirmedBodyFR,
		},
		"status_preparing": {
			Subject: "Votre commande est en préparation — {{.BakeryName}}",
			Body:    statusPreparingBodyFR,
		},
		"status_ready": {
			Subject: "Votre commande est prête — {{.BakeryName}}",
			Body:    statusReadyBodyFR,
		},
		"status_delivered": {
			Subject: "Commande livrée — {{.BakeryName}}",
			Body:    statusDeliveredBodyFR,
		},
		"new_order_baker": {
			Subject: "Nouvelle commande reçue ! — {{.CustomerName}}",
			Body:    newOrderBakerBodyFR,
		},
		"reservation_confirmed": {
			Subject: "Réservation confirmée — {{.BakeryName}}",
			Body:    reservationConfirmedBodyFR,
		},
	},
	LocaleNL: {
		"order_confirmed": {
			Subject: "Bestelling bevestigd — {{.BakeryName}}",
			Body:    orderConfirmedBodyNL,
		},
		"status_preparing": {
			Subject: "Je bestelling wordt bereid — {{.BakeryName}}",
			Body:    statusPreparingBodyNL,
		},
		"status_ready": {
			Subject: "Je bestelling is klaar — {{.BakeryName}}",
			Body:    statusReadyBodyNL,
		},
		"status_delivered": {
			Subject: "Bestelling afgeleverd — {{.BakeryName}}",
			Body:    statusDeliveredBodyNL,
		},
		"new_order_baker": {
			Subject: "Nieuwe bestelling ontvangen! — {{.CustomerName}}",
			Body:    newOrderBakerBodyNL,
		},
		"reservation_confirmed": {
			Subject: "Reservering bevestigd — {{.BakeryName}}",
			Body:    reservationConfirmedBodyNL,
		},
	},
}

// renderTemplate renders the given event template for the specified locale with data.
// Falls back to English if the locale or event is not found.
func renderTemplate(locale Locale, event string, data TemplateData) (subject string, body string, err error) {
	tmplSet, ok := templates[locale]
	if !ok {
		tmplSet = templates[LocaleEN]
	}
	tmpl, ok := tmplSet[event]
	if !ok {
		// Final fallback to EN
		tmpl, ok = templates[LocaleEN][event]
		if !ok {
			return "", "", fmt.Errorf("unknown notification event: %s", event)
		}
	}

	subject, err = executeTemplateString(tmpl.Subject, data)
	if err != nil {
		return "", "", fmt.Errorf("rendering subject for %s/%s: %w", locale, event, err)
	}

	body, err = executeTemplateString(tmpl.Body, data)
	if err != nil {
		return "", "", fmt.Errorf("rendering body for %s/%s: %w", locale, event, err)
	}

	return subject, body, nil
}

func executeTemplateString(tmplStr string, data TemplateData) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatCentsForDisplay formats cents as a €X.XX string.
func formatCentsForDisplay(cents int64) string {
	return fmt.Sprintf("€%.2f", float64(cents)/100)
}
