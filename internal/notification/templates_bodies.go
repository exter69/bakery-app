package notification

// ─── English templates ───────────────────────────────────────────────────────

const orderConfirmedBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Your order is confirmed! 🎉</h2>
  <p>Great news — <strong>{{.BakeryName}}</strong> has received your order and it will be prepared soon.</p>
  <div class="details">
    <p><strong>Order:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total: {{.TotalDisplay}}</strong></p>
  </div>
  <p>You'll receive updates as your order progresses.</p>
  <div class="footer"><p>Thank you for ordering with {{.BakeryName}}!</p></div>
</div>
</body></html>`

const statusPreparingBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Good news! 🍞</h2>
  <p><strong>{{.BakeryName}}</strong> is now preparing your order.</p>
  <div class="details">
    <p><strong>Order:</strong> {{.OrderID}}</p>
  </div>
  <p>We'll let you know when it's ready.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusReadyBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Your order is ready! ✅</h2>
  <p>Your order from <strong>{{.BakeryName}}</strong> is ready and will be delivered soon.</p>
  <div class="details">
    <p><strong>Order:</strong> {{.OrderID}}</p>
  </div>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusDeliveredBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Delivered! 🎉</h2>
  <p>Your order from <strong>{{.BakeryName}}</strong> has been delivered. Enjoy your fresh baked goods!</p>
  <div class="details">
    <p><strong>Order:</strong> {{.OrderID}}</p>
    <p class="total"><strong>Total charged: {{.TotalDisplay}}</strong></p>
  </div>
  <p>Thank you for your order — we hope you enjoy everything!</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const newOrderBakerBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>New order received! 📦</h2>
  <p>You have a new order from <strong>{{.CustomerName}}</strong>.</p>
  <div class="details">
    <p><strong>Order:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total: {{.TotalDisplay}}</strong></p>
  </div>
  <p>Head to your dashboard to manage this order.</p>
  <div class="footer"><p>{{.BakeryName}} — Seller Dashboard</p></div>
</div>
</body></html>`

const reservationConfirmedBodyEN = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Reservation confirmed! 📋</h2>
  <p>Your reservation at <strong>{{.BakeryName}}</strong> has been confirmed.</p>
  <div class="details">
    <p><strong>Reservation:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total (pay on pickup): {{.TotalDisplay}}</strong></p>
  </div>
  <p>Remember to bring payment when you pick up your order.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

// ─── French templates ────────────────────────────────────────────────────────

const orderConfirmedBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Votre commande est confirmée ! 🎉</h2>
  <p>Bonne nouvelle — <strong>{{.BakeryName}}</strong> a bien reçu votre commande et la préparera bientôt.</p>
  <div class="details">
    <p><strong>Commande :</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total : {{.TotalDisplay}}</strong></p>
  </div>
  <p>Vous recevrez des mises à jour au fur et à mesure de l'avancement de votre commande.</p>
  <div class="footer"><p>Merci d'avoir commandé chez {{.BakeryName}} !</p></div>
</div>
</body></html>`

const statusPreparingBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Bonne nouvelle ! 🍞</h2>
  <p><strong>{{.BakeryName}}</strong> prépare maintenant votre commande.</p>
  <div class="details">
    <p><strong>Commande :</strong> {{.OrderID}}</p>
  </div>
  <p>Nous vous préviendrons quand elle sera prête.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusReadyBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Votre commande est prête ! ✅</h2>
  <p>Votre commande de <strong>{{.BakeryName}}</strong> est prête et sera livrée bientôt.</p>
  <div class="details">
    <p><strong>Commande :</strong> {{.OrderID}}</p>
  </div>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusDeliveredBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Livrée ! 🎉</h2>
  <p>Votre commande de <strong>{{.BakeryName}}</strong> a été livrée. Bon appétit !</p>
  <div class="details">
    <p><strong>Commande :</strong> {{.OrderID}}</p>
    <p class="total"><strong>Total facturé : {{.TotalDisplay}}</strong></p>
  </div>
  <p>Merci pour votre commande — nous espérons que tout vous plaira !</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const newOrderBakerBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Nouvelle commande reçue ! 📦</h2>
  <p>Vous avez une nouvelle commande de <strong>{{.CustomerName}}</strong>.</p>
  <div class="details">
    <p><strong>Commande :</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total : {{.TotalDisplay}}</strong></p>
  </div>
  <p>Rendez-vous sur votre tableau de bord pour gérer cette commande.</p>
  <div class="footer"><p>{{.BakeryName}} — Tableau de bord vendeur</p></div>
</div>
</body></html>`

const reservationConfirmedBodyFR = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Réservation confirmée ! 📋</h2>
  <p>Votre réservation chez <strong>{{.BakeryName}}</strong> est confirmée.</p>
  <div class="details">
    <p><strong>Réservation :</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Total (à régler sur place) : {{.TotalDisplay}}</strong></p>
  </div>
  <p>N'oubliez pas d'apporter votre moyen de paiement lors du retrait.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

// ─── Dutch templates ─────────────────────────────────────────────────────────

const orderConfirmedBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Je bestelling is bevestigd! 🎉</h2>
  <p>Goed nieuws — <strong>{{.BakeryName}}</strong> heeft je bestelling ontvangen en zal deze binnenkort bereiden.</p>
  <div class="details">
    <p><strong>Bestelling:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Totaal: {{.TotalDisplay}}</strong></p>
  </div>
  <p>Je ontvangt updates naarmate je bestelling vordert.</p>
  <div class="footer"><p>Bedankt voor je bestelling bij {{.BakeryName}}!</p></div>
</div>
</body></html>`

const statusPreparingBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Goed nieuws! 🍞</h2>
  <p><strong>{{.BakeryName}}</strong> bereidt nu je bestelling.</p>
  <div class="details">
    <p><strong>Bestelling:</strong> {{.OrderID}}</p>
  </div>
  <p>We laten je weten wanneer het klaar is.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusReadyBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Je bestelling is klaar! ✅</h2>
  <p>Je bestelling van <strong>{{.BakeryName}}</strong> is klaar en wordt binnenkort bezorgd.</p>
  <div class="details">
    <p><strong>Bestelling:</strong> {{.OrderID}}</p>
  </div>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const statusDeliveredBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Afgeleverd! 🎉</h2>
  <p>Je bestelling van <strong>{{.BakeryName}}</strong> is bezorgd. Geniet van je verse bakproducten!</p>
  <div class="details">
    <p><strong>Bestelling:</strong> {{.OrderID}}</p>
    <p class="total"><strong>Totaal afgeschreven: {{.TotalDisplay}}</strong></p>
  </div>
  <p>Bedankt voor je bestelling — we hopen dat alles naar wens is!</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

const newOrderBakerBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Nieuwe bestelling ontvangen! 📦</h2>
  <p>Je hebt een nieuwe bestelling van <strong>{{.CustomerName}}</strong>.</p>
  <div class="details">
    <p><strong>Bestelling:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Totaal: {{.TotalDisplay}}</strong></p>
  </div>
  <p>Ga naar je dashboard om deze bestelling te beheren.</p>
  <div class="footer"><p>{{.BakeryName}} — Verkoper Dashboard</p></div>
</div>
</body></html>`

const reservationConfirmedBodyNL = `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><style>` + emailStyles + `</style></head><body>
<div class="container">
  <h2>Reservering bevestigd! 📋</h2>
  <p>Je reservering bij <strong>{{.BakeryName}}</strong> is bevestigd.</p>
  <div class="details">
    <p><strong>Reservering:</strong> {{.OrderID}}</p>
    {{range .Items}}<p>• {{.ProductName}} × {{.Quantity}} — {{.Subtotal}}</p>{{end}}
    <p class="total"><strong>Totaal (betaling bij afhaling): {{.TotalDisplay}}</strong></p>
  </div>
  <p>Vergeet niet om betaalmiddel mee te nemen bij het ophalen.</p>
  <div class="footer"><p>{{.BakeryName}}</p></div>
</div>
</body></html>`

// ─── Shared email styles ─────────────────────────────────────────────────────

const emailStyles = `
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
  .container { max-width: 600px; margin: 20px auto; background: #fff; border-radius: 8px; padding: 32px; box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
  h2 { color: #2c3e50; margin-top: 0; }
  p { color: #333; line-height: 1.6; }
  .details { background: #f8f9fa; border-radius: 6px; padding: 16px; margin: 20px 0; }
  .details p { margin: 6px 0; }
  .total { margin-top: 12px; font-size: 1.1em; }
  .footer { margin-top: 32px; padding-top: 16px; border-top: 1px solid #eee; }
  .footer p { color: #888; font-size: 0.85em; }
`
