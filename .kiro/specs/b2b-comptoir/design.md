# Design Document: B2B Comptoir Portal

## Overview

The B2B Comptoir Portal adds a dedicated professional ordering interface for business clients (restaurants, hotels, cafeterias) to the existing bakery platform. Business users register with company details, get whitelisted by bakeries, and access a streamlined portal for bulk ordering with invoice-based payment, recurring templates, and delivery site management.

## Architecture

The feature follows the same layered architecture: **Handler (API) → Service → Repository**, using Go + chi + pgx on the backend and React + TypeScript + Vite on the frontend.

### Key Design Decisions

1. **Separate role (RoleBusiness = 3)** rather than a permission flag on existing users. This keeps the JWT claims compatible with the existing middleware and allows clean route guarding.
2. **Reuse existing Order model** with `payment_method = "on_invoice"` rather than a separate B2B order table. B2B orders are just orders with different payment semantics and additional metadata (delivery_site_id, discount fields).
3. **Multi-bakery cart in frontend localStorage** — the cart is a client-side concern. The backend only sees individual per-bakery checkout requests.
4. **Per-bakery B2B config table** (`b2b_config`) rather than adding columns to the bakeries table. This keeps B2B concerns decoupled and allows bakeries to opt in incrementally.
5. **Access whitelisting via `bakery_b2b_access`** — explicit approval model where bakeries control which business clients can order from them.
6. **Invoice generation on delivery** — invoices are created when an order is marked delivered, with sequential numbering per bakery.
7. **Frontend as separate route tree** at `/comptoir/*` with its own layout, nav, and theme — no shared state with consumer portal beyond the auth layer.

---

## Database Migration: `019_create_b2b_tables.sql`

Migration follows the existing goose format (see `018_create_surplus_bundles.sql`).

```sql
-- +goose Up

-- Business profiles linked to users with role 3
CREATE TABLE business_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(200) NOT NULL,
    vat_siret VARCHAR(20) NOT NULL UNIQUE,
    iban VARCHAR(34) NOT NULL,
    billing_email VARCHAR(255) NOT NULL,
    billing_contact_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_business_profiles_user_id ON business_profiles(user_id);
CREATE UNIQUE INDEX idx_business_profiles_vat ON business_profiles(vat_siret);

-- Delivery sites per business user
CREATE TABLE delivery_sites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    street_address VARCHAR(300) NOT NULL,
    city VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10) NOT NULL,
    country VARCHAR(2) NOT NULL DEFAULT 'BE',
    delivery_instructions VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_sites_user_id ON delivery_sites(user_id);

-- Bakery-to-business-user access whitelisting
CREATE TABLE bakery_b2b_access (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    business_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'revoked')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_b2b_access_bakery_user ON bakery_b2b_access(bakery_id, business_user_id);
CREATE INDEX idx_b2b_access_business_user ON bakery_b2b_access(business_user_id);
CREATE INDEX idx_b2b_access_status ON bakery_b2b_access(bakery_id, status);

-- Per-bakery B2B configuration
CREATE TABLE b2b_config (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL UNIQUE REFERENCES bakeries(id) ON DELETE CASCADE,
    cutoff_time TIME NOT NULL DEFAULT '18:00',
    delivery_window_start TIME NOT NULL DEFAULT '06:00',
    delivery_window_end TIME NOT NULL DEFAULT '09:00',
    order_minimum BIGINT NOT NULL DEFAULT 2000,  -- cents HT, default 20 EUR
    pro_discount INT NOT NULL DEFAULT 0 CHECK (pro_discount >= 0 AND pro_discount <= 100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_b2b_config_bakery_id ON b2b_config(bakery_id);

-- Saved product lists for Commande Rapide
CREATE TABLE saved_lists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_lists_user_bakery ON saved_lists(user_id, bakery_id);

CREATE TABLE saved_list_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    saved_list_id UUID NOT NULL REFERENCES saved_lists(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity >= 1)
);

CREATE INDEX idx_saved_list_items_list_id ON saved_list_items(saved_list_id);

-- B2B invoices generated on delivery
CREATE TABLE b2b_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    business_profile_id UUID NOT NULL REFERENCES business_profiles(id),
    invoice_number INT NOT NULL,
    subtotal_ht BIGINT NOT NULL,
    discount_amount BIGINT NOT NULL DEFAULT 0,
    tva_amount BIGINT NOT NULL,
    total_ttc BIGINT NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (payment_status IN ('pending', 'paid', 'overdue')),
    issued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMP
);

-- Sequential invoice numbers per bakery
CREATE UNIQUE INDEX idx_b2b_invoices_bakery_number ON b2b_invoices(bakery_id, invoice_number);
CREATE INDEX idx_b2b_invoices_business_profile ON b2b_invoices(business_profile_id);
CREATE INDEX idx_b2b_invoices_order ON b2b_invoices(order_id);

-- Add delivery_site_id and B2B pricing columns to orders
ALTER TABLE orders ADD COLUMN delivery_site_id UUID REFERENCES delivery_sites(id);
ALTER TABLE orders ADD COLUMN subtotal_ht BIGINT;
ALTER TABLE orders ADD COLUMN discount_amount BIGINT DEFAULT 0;
ALTER TABLE orders ADD COLUMN tva_amount BIGINT;

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS tva_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS subtotal_ht;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_site_id;
DROP TABLE IF EXISTS b2b_invoices;
DROP TABLE IF EXISTS saved_list_items;
DROP TABLE IF EXISTS saved_lists;
DROP TABLE IF EXISTS b2b_config;
DROP TABLE IF EXISTS bakery_b2b_access;
DROP TABLE IF EXISTS delivery_sites;
DROP TABLE IF EXISTS business_profiles;
```

---

## Data Models

### Go Domain Types

New types added to `internal/domain/` (in a new `b2b_models.go` file):

```go
package domain

import "time"

const RoleBusiness UserRole = 3

// B2BAccessStatus represents the state of a bakery-business access request.
type B2BAccessStatus string

const (
    B2BAccessPending  B2BAccessStatus = "pending"
    B2BAccessApproved B2BAccessStatus = "approved"
    B2BAccessRejected B2BAccessStatus = "rejected"
    B2BAccessRevoked  B2BAccessStatus = "revoked"
)

// BusinessProfile holds company-level details for a B2B user.
type BusinessProfile struct {
    ID                 string    `json:"id"`
    UserID             string    `json:"userId"`
    CompanyName        string    `json:"companyName"`
    VATSiret           string    `json:"vatSiret"`
    IBAN               string    `json:"iban"`
    BillingEmail       string    `json:"billingEmail"`
    BillingContactName string    `json:"billingContactName"`
    CreatedAt          time.Time `json:"createdAt"`
    UpdatedAt          time.Time `json:"updatedAt"`
}

// DeliverySite represents a delivery address for a business user.
type DeliverySite struct {
    ID                   string    `json:"id"`
    UserID               string    `json:"userId"`
    Name                 string    `json:"name"`
    StreetAddress        string    `json:"streetAddress"`
    City                 string    `json:"city"`
    PostalCode           string    `json:"postalCode"`
    Country              string    `json:"country"`
    DeliveryInstructions string    `json:"deliveryInstructions,omitempty"`
    CreatedAt            time.Time `json:"createdAt"`
    UpdatedAt            time.Time `json:"updatedAt"`
}

// B2BAccess represents a bakery-to-business whitelisting record.
type B2BAccess struct {
    ID             string          `json:"id"`
    BakeryID       string          `json:"bakeryId"`
    BusinessUserID string          `json:"businessUserId"`
    Status         B2BAccessStatus `json:"status"`
    CreatedAt      time.Time       `json:"createdAt"`
    UpdatedAt      time.Time       `json:"updatedAt"`
}

// B2BConfig holds per-bakery B2B ordering rules.
type B2BConfig struct {
    ID                  string    `json:"id"`
    BakeryID            string    `json:"bakeryId"`
    CutoffTime          TimeOfDay `json:"cutoffTime"`
    DeliveryWindowStart TimeOfDay `json:"deliveryWindowStart"`
    DeliveryWindowEnd   TimeOfDay `json:"deliveryWindowEnd"`
    OrderMinimum        int64     `json:"orderMinimum"`    // cents HT
    ProDiscount         int       `json:"proDiscount"`     // percentage 0-100
    CreatedAt           time.Time `json:"createdAt"`
    UpdatedAt           time.Time `json:"updatedAt"`
}

// SavedList holds a named product list for Commande Rapide.
type SavedList struct {
    ID        string          `json:"id"`
    UserID    string          `json:"userId"`
    BakeryID  string          `json:"bakeryId"`
    Name      string          `json:"name"`
    Items     []SavedListItem `json:"items"`
    CreatedAt time.Time       `json:"createdAt"`
    UpdatedAt time.Time       `json:"updatedAt"`
}

// SavedListItem is a product-quantity pair within a SavedList.
type SavedListItem struct {
    ID        string `json:"id"`
    ProductID string `json:"productId"`
    Quantity  int    `json:"quantity"`
}

// B2BInvoice represents a generated invoice for a delivered B2B order.
type B2BInvoice struct {
    ID                string    `json:"id"`
    OrderID           string    `json:"orderId"`
    BakeryID          string    `json:"bakeryId"`
    BusinessProfileID string    `json:"businessProfileId"`
    InvoiceNumber     int       `json:"invoiceNumber"`
    SubtotalHT        int64     `json:"subtotalHt"`
    DiscountAmount    int64     `json:"discountAmount"`
    TVAAmount         int64     `json:"tvaAmount"`
    TotalTTC          int64     `json:"totalTtc"`
    PaymentStatus     string    `json:"paymentStatus"`
    IssuedAt          time.Time `json:"issuedAt"`
    PaidAt            *time.Time `json:"paidAt,omitempty"`
}

// B2BOrderPricing computes and holds the pricing breakdown for a B2B order.
type B2BOrderPricing struct {
    SubtotalHT     int64 `json:"subtotalHt"`     // sum of line items (qty * unit_price)
    DiscountRate   int   `json:"discountRate"`   // percentage
    DiscountAmount int64 `json:"discountAmount"` // subtotalHT * discountRate / 100
    TVARate        int   `json:"tvaRate"`        // percentage (6 for Belgium bakery products)
    TVAAmount      int64 `json:"tvaAmount"`      // (subtotalHT - discountAmount) * tvaRate / 100
    TotalTTC       int64 `json:"totalTtc"`       // subtotalHT - discountAmount + tvaAmount
}
```

---

## Components and Interfaces

### Service Interface

New file `internal/domain/b2b_services.go`:

```go
package domain

import "context"

// B2BService handles all B2B-specific business logic.
type B2BService interface {
    // --- Registration & Profile ---

    // RegisterBusiness creates a user with RoleBusiness and a linked BusinessProfile.
    RegisterBusiness(ctx context.Context, req RegisterBusinessRequest) (*BusinessProfile, error)

    // GetProfile returns the BusinessProfile for the authenticated user.
    GetProfile(ctx context.Context, userID string) (*BusinessProfile, error)

    // UpdateProfile updates mutable fields of the BusinessProfile (not VAT/SIRET).
    UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*BusinessProfile, error)

    // --- Delivery Sites ---

    // CreateSite adds a delivery site for the business user.
    CreateSite(ctx context.Context, userID string, site DeliverySite) (*DeliverySite, error)

    // ListSites returns all delivery sites for the business user.
    ListSites(ctx context.Context, userID string) ([]DeliverySite, error)

    // UpdateSite updates a delivery site. Verifies ownership.
    UpdateSite(ctx context.Context, userID string, siteID string, site DeliverySite) (*DeliverySite, error)

    // DeleteSite removes a delivery site. Rejects if it's the only one.
    DeleteSite(ctx context.Context, userID string, siteID string) error

    // --- Access Management ---

    // RequestAccess creates a pending access request for a bakery.
    RequestAccess(ctx context.Context, userID string, bakeryID string) (*B2BAccess, error)

    // ApproveAccess approves a pending access request (seller action).
    ApproveAccess(ctx context.Context, accessID string) error

    // RejectAccess rejects a pending access request (seller action).
    RejectAccess(ctx context.Context, accessID string) error

    // RevokeAccess revokes an approved access (seller action).
    RevokeAccess(ctx context.Context, accessID string) error

    // ListAccessRequests returns pending access requests for a bakery (seller view).
    ListAccessRequests(ctx context.Context, bakeryID string) ([]B2BAccess, error)

    // ListApprovedBakeries returns bakeries where the user has approved access.
    ListApprovedBakeries(ctx context.Context, userID string) ([]Bakery, error)

    // HasApprovedAccess checks if a user has approved access to a bakery.
    HasApprovedAccess(ctx context.Context, userID string, bakeryID string) (bool, error)

    // --- B2B Config ---

    // GetConfig returns the B2B config for a bakery.
    GetConfig(ctx context.Context, bakeryID string) (*B2BConfig, error)

    // SaveConfig creates or updates B2B config for a bakery (seller action).
    SaveConfig(ctx context.Context, bakeryID string, config B2BConfig) (*B2BConfig, error)

    // --- Cart & Checkout ---

    // CheckoutBakeryGroup validates and creates a B2B order for a single bakery group.
    // Enforces: access check, order minimum, cutoff time, delivery site requirement.
    CheckoutBakeryGroup(ctx context.Context, userID string, req CheckoutRequest) (*Order, error)

    // EditOrder modifies items on a submitted order (before cutoff).
    EditOrder(ctx context.Context, userID string, orderID string, req EditOrderRequest) (*Order, error)

    // ComputePricing calculates the full pricing breakdown for a set of items.
    ComputePricing(ctx context.Context, bakeryID string, items []OrderItem) (*B2BOrderPricing, error)

    // --- Saved Lists ---

    // CreateSavedList stores a named product list.
    CreateSavedList(ctx context.Context, userID string, list SavedList) (*SavedList, error)

    // ListSavedLists returns all saved lists for a user+bakery.
    ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]SavedList, error)

    // DeleteSavedList removes a saved list. Verifies ownership.
    DeleteSavedList(ctx context.Context, userID string, listID string) error

    // --- Deliveries & Invoices ---

    // ListDeliveries returns B2B orders for a user with filters and pagination.
    ListDeliveries(ctx context.Context, userID string, filters B2BOrderFilters, params PaginationParams) (*ListResult[Order], error)

    // GetLastOrder returns the most recent B2B order for a user at a bakery.
    GetLastOrder(ctx context.Context, userID string, bakeryID string) (*Order, error)

    // ListInvoices returns invoices for the business user with pagination.
    ListInvoices(ctx context.Context, userID string, params PaginationParams) (*ListResult[B2BInvoice], error)

    // GenerateInvoice creates an invoice record for a delivered B2B order.
    GenerateInvoice(ctx context.Context, orderID string) (*B2BInvoice, error)

    // DownloadInvoicePDF generates a PDF for the given invoice.
    DownloadInvoicePDF(ctx context.Context, invoiceID string, userID string) ([]byte, error)
}

// RegisterBusinessRequest is the input for B2B registration.
type RegisterBusinessRequest struct {
    Username           string `json:"username"`
    Password           string `json:"password"`
    CompanyName        string `json:"companyName"`
    VATSiret           string `json:"vatSiret"`
    IBAN               string `json:"iban"`
    BillingEmail       string `json:"billingEmail"`
    BillingContactName string `json:"billingContactName"`
}

// UpdateProfileRequest is the input for profile updates (VAT/SIRET excluded).
type UpdateProfileRequest struct {
    CompanyName        string `json:"companyName"`
    IBAN               string `json:"iban"`
    BillingEmail       string `json:"billingEmail"`
    BillingContactName string `json:"billingContactName"`
}

// CheckoutRequest is the input for per-bakery B2B checkout.
type CheckoutRequest struct {
    BakeryID       string      `json:"bakeryId"`
    DeliverySiteID string      `json:"deliverySiteId"`
    Items          []OrderItem `json:"items"`
}

// EditOrderRequest is the input for editing a submitted B2B order.
type EditOrderRequest struct {
    Items []OrderItem `json:"items"`
}

// B2BOrderFilters extends OrderFilters for B2B-specific filtering.
type B2BOrderFilters struct {
    BakeryID  string     `json:"bakeryId,omitempty"`
    Status    string     `json:"status,omitempty"`
    DateFrom  *time.Time `json:"dateFrom,omitempty"`
    DateTo    *time.Time `json:"dateTo,omitempty"`
}
```

---

### Repository Interface

New file `internal/domain/b2b_repository.go`:

```go
package domain

import "context"

// B2BRepository provides data access for B2B-specific tables.
type B2BRepository interface {
    // --- Business Profiles ---
    CreateProfile(ctx context.Context, profile *BusinessProfile) error
    GetProfileByUserID(ctx context.Context, userID string) (*BusinessProfile, error)
    GetProfileByVAT(ctx context.Context, vatSiret string) (*BusinessProfile, error)
    UpdateProfile(ctx context.Context, profile *BusinessProfile) error

    // --- Delivery Sites ---
    CreateSite(ctx context.Context, site *DeliverySite) error
    GetSiteByID(ctx context.Context, id string) (*DeliverySite, error)
    ListSitesByUser(ctx context.Context, userID string) ([]DeliverySite, error)
    UpdateSite(ctx context.Context, site *DeliverySite) error
    DeleteSite(ctx context.Context, id string) error
    CountSitesByUser(ctx context.Context, userID string) (int, error)

    // --- Access Whitelisting ---
    CreateAccess(ctx context.Context, access *B2BAccess) error
    GetAccessByID(ctx context.Context, id string) (*B2BAccess, error)
    GetAccess(ctx context.Context, bakeryID string, userID string) (*B2BAccess, error)
    UpdateAccessStatus(ctx context.Context, id string, status B2BAccessStatus) error
    ListAccessByBakery(ctx context.Context, bakeryID string, status *B2BAccessStatus) ([]B2BAccess, error)
    ListApprovedBakeryIDs(ctx context.Context, userID string) ([]string, error)

    // --- B2B Config ---
    GetConfig(ctx context.Context, bakeryID string) (*B2BConfig, error)
    SaveConfig(ctx context.Context, config *B2BConfig) error

    // --- Saved Lists ---
    CreateSavedList(ctx context.Context, list *SavedList) error
    GetSavedListByID(ctx context.Context, id string) (*SavedList, error)
    ListSavedLists(ctx context.Context, userID string, bakeryID string) ([]SavedList, error)
    DeleteSavedList(ctx context.Context, id string) error

    // --- Invoices ---
    CreateInvoice(ctx context.Context, invoice *B2BInvoice) error
    GetInvoiceByID(ctx context.Context, id string) (*B2BInvoice, error)
    GetInvoiceByOrder(ctx context.Context, orderID string) (*B2BInvoice, error)
    ListInvoicesByUser(ctx context.Context, profileID string, params PaginationParams) ([]B2BInvoice, int, error)
    NextInvoiceNumber(ctx context.Context, bakeryID string) (int, error)
}
```

---

### API Handler & Routes

New file `internal/api/b2b_handler.go`:

```go
package api

import "github.com/go-chi/chi/v5"

// B2BHandler handles HTTP requests for the B2B Comptoir portal.
type B2BHandler struct {
    svc domain.B2BService
    now func() time.Time
}

func NewB2BHandler(svc domain.B2BService) *B2BHandler {
    return &B2BHandler{svc: svc, now: time.Now}
}

// RegisterRoutes mounts B2B routes on the chi router.
// All routes require JWT auth; role enforcement handled by middleware.
func (h *B2BHandler) RegisterRoutes(r chi.Router) {
    // Public: B2B registration (no auth required)
    r.Post("/api/comptoir/register", h.Register)

    // Protected: requires RoleBusiness (role 3) JWT
    r.Route("/api/comptoir", func(r chi.Router) {
        r.Use(requireRole(RoleBusiness))

        // Profile
        r.Get("/profile", h.GetProfile)
        r.Put("/profile", h.UpdateProfile)

        // Delivery Sites
        r.Post("/sites", h.CreateSite)
        r.Get("/sites", h.ListSites)
        r.Put("/sites/{siteId}", h.UpdateSite)
        r.Delete("/sites/{siteId}", h.DeleteSite)

        // Access Requests
        r.Post("/access/request/{bakeryId}", h.RequestAccess)
        r.Get("/bakeries", h.ListApprovedBakeries)

        // Product Catalog (requires approved access, checked in handler)
        r.Get("/bakeries/{bakeryId}/products", h.GetProducts)

        // B2B Config (read-only for business users)
        r.Get("/bakeries/{bakeryId}/config", h.GetConfig)

        // Cart & Checkout
        r.Post("/checkout", h.Checkout)
        r.Put("/orders/{orderId}", h.EditOrder)
        r.Post("/pricing", h.ComputePricing)

        // Saved Lists
        r.Post("/lists", h.CreateSavedList)
        r.Get("/lists", h.ListSavedLists)
        r.Delete("/lists/{listId}", h.DeleteSavedList)

        // Deliveries
        r.Get("/deliveries", h.ListDeliveries)
        r.Get("/orders/{bakeryId}/last", h.GetLastOrder)

        // Invoices
        r.Get("/invoices", h.ListInvoices)
        r.Get("/invoices/{invoiceId}/pdf", h.DownloadInvoicePDF)
    })

    // Baker-facing B2B management (requires RoleSeller or RoleAdmin)
    r.Route("/api/dashboard/b2b", func(r chi.Router) {
        r.Use(requireRole(RoleSeller, RoleAdmin))

        r.Get("/access", h.ListAccessRequests)
        r.Post("/access/{accessId}/approve", h.ApproveAccess)
        r.Post("/access/{accessId}/reject", h.RejectAccess)
        r.Post("/access/{accessId}/revoke", h.RevokeAccess)

        r.Get("/config", h.GetBakerConfig)
        r.Put("/config", h.SaveConfig)
    })
}
```

### Route Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/comptoir/register` | None | B2B user registration |
| GET | `/api/comptoir/profile` | Business | Get business profile |
| PUT | `/api/comptoir/profile` | Business | Update profile |
| POST | `/api/comptoir/sites` | Business | Create delivery site |
| GET | `/api/comptoir/sites` | Business | List delivery sites |
| PUT | `/api/comptoir/sites/{siteId}` | Business | Update delivery site |
| DELETE | `/api/comptoir/sites/{siteId}` | Business | Delete delivery site |
| POST | `/api/comptoir/access/request/{bakeryId}` | Business | Request bakery access |
| GET | `/api/comptoir/bakeries` | Business | List approved bakeries |
| GET | `/api/comptoir/bakeries/{bakeryId}/products` | Business | Get product catalog |
| GET | `/api/comptoir/bakeries/{bakeryId}/config` | Business | Get bakery B2B config |
| POST | `/api/comptoir/checkout` | Business | Per-bakery checkout |
| PUT | `/api/comptoir/orders/{orderId}` | Business | Edit order before cutoff |
| POST | `/api/comptoir/pricing` | Business | Compute pricing breakdown |
| POST | `/api/comptoir/lists` | Business | Create saved list |
| GET | `/api/comptoir/lists` | Business | List saved lists |
| DELETE | `/api/comptoir/lists/{listId}` | Business | Delete saved list |
| GET | `/api/comptoir/deliveries` | Business | List deliveries |
| GET | `/api/comptoir/orders/{bakeryId}/last` | Business | Get last order for bakery |
| GET | `/api/comptoir/invoices` | Business | List invoices |
| GET | `/api/comptoir/invoices/{invoiceId}/pdf` | Business | Download invoice PDF |
| GET | `/api/dashboard/b2b/access` | Seller/Admin | List access requests |
| POST | `/api/dashboard/b2b/access/{accessId}/approve` | Seller/Admin | Approve access |
| POST | `/api/dashboard/b2b/access/{accessId}/reject` | Seller/Admin | Reject access |
| POST | `/api/dashboard/b2b/access/{accessId}/revoke` | Seller/Admin | Revoke access |
| GET | `/api/dashboard/b2b/config` | Seller/Admin | Get B2B config |
| PUT | `/api/dashboard/b2b/config` | Seller/Admin | Save B2B config |

---

### Frontend Architecture

### Component Tree

```
/comptoir/*
├── ComptoirLayout
│   ├── ComptoirNav (top bar)
│   │   ├── NavTabs: Commander | Recurrences | Livraisons | Factures
│   │   ├── SiteSwitcher (delivery site dropdown)
│   │   └── AccountMenu (company name, settings link)
│   └── <Outlet /> (page content)
├── Pages
│   ├── CommanderPage (Commande Rapide interface)
│   ├── RecurrencesPage (recurring order templates)
│   ├── LivraisonsPage (delivery history)
│   └── FacturesPage (invoice list)
└── Shared Components
    ├── CommandeRapide (spreadsheet grid)
    ├── DeliveryList (delivery entries with filters)
    ├── InvoiceList (invoice table with download links)
    ├── B2BCartSummary (per-bakery pricing breakdown)
    └── SavedListPicker (dropdown to select/save lists)
```

### Router Integration

Added to `App.tsx` alongside existing routes:

```tsx
import ComptoirLayout from './pages/comptoir/ComptoirLayout';
import CommanderPage from './pages/comptoir/CommanderPage';
import RecurrencesPage from './pages/comptoir/RecurrencesPage';
import LivraisonsPage from './pages/comptoir/LivraisonsPage';
import FacturesPage from './pages/comptoir/FacturesPage';
import ComptoirProfilePage from './pages/comptoir/ComptoirProfilePage';

// B2B Comptoir (role 3)
<Route path="/comptoir" element={<RoleRoute allowedRoles={[3]}><ComptoirLayout /></RoleRoute>}>
  <Route index element={<CommanderPage />} />
  <Route path="recurrences" element={<RecurrencesPage />} />
  <Route path="livraisons" element={<LivraisonsPage />} />
  <Route path="factures" element={<FacturesPage />} />
  <Route path="profile" element={<ComptoirProfilePage />} />
</Route>
```

### Frontend Types

New file `frontend/src/types/b2b.ts`:

```typescript
export interface BusinessProfile {
  id: string;
  userId: string;
  companyName: string;
  vatSiret: string;
  iban: string;
  billingEmail: string;
  billingContactName: string;
  createdAt: string;
  updatedAt: string;
}

export interface DeliverySite {
  id: string;
  userId: string;
  name: string;
  streetAddress: string;
  city: string;
  postalCode: string;
  country: string;
  deliveryInstructions?: string;
  createdAt: string;
  updatedAt: string;
}

export type B2BAccessStatus = 'pending' | 'approved' | 'rejected' | 'revoked';

export interface B2BAccess {
  id: string;
  bakeryId: string;
  businessUserId: string;
  status: B2BAccessStatus;
  createdAt: string;
  updatedAt: string;
}

export interface B2BConfig {
  id: string;
  bakeryId: string;
  cutoffTime: string;       // "HH:MM"
  deliveryWindowStart: string;
  deliveryWindowEnd: string;
  orderMinimum: number;     // cents HT
  proDiscount: number;      // percentage 0-100
}

export interface SavedList {
  id: string;
  userId: string;
  bakeryId: string;
  name: string;
  items: SavedListItem[];
  createdAt: string;
  updatedAt: string;
}

export interface SavedListItem {
  id: string;
  productId: string;
  quantity: number;
}

export interface B2BInvoice {
  id: string;
  orderId: string;
  bakeryId: string;
  businessProfileId: string;
  invoiceNumber: number;
  subtotalHt: number;
  discountAmount: number;
  tvaAmount: number;
  totalTtc: number;
  paymentStatus: 'pending' | 'paid' | 'overdue';
  issuedAt: string;
  paidAt?: string;
}

export interface B2BCartItem {
  productId: string;
  productName: string;
  unitPrice: number;
  quantity: number;
}

export interface B2BCartGroup {
  bakeryId: string;
  bakeryName: string;
  items: B2BCartItem[];
  subtotalHt: number;
}

export interface B2BCart {
  groups: B2BCartGroup[];
  totalHt: number;
}

export interface B2BOrderPricing {
  subtotalHt: number;
  discountRate: number;
  discountAmount: number;
  tvaRate: number;
  tvaAmount: number;
  totalTtc: number;
}
```

### API Client

New file `frontend/src/api/b2b-client.ts`:

```typescript
import { apiFetch } from './client';
import type {
  BusinessProfile, DeliverySite, B2BConfig,
  SavedList, B2BInvoice, B2BOrderPricing,
} from '../types/b2b';
import type { Bakery, Product, Order } from '../types/bakery';

// Registration
export function registerBusiness(data: {
  username: string;
  password: string;
  companyName: string;
  vatSiret: string;
  iban: string;
  billingEmail: string;
  billingContactName: string;
}) {
  return apiFetch<{ token: string; profile: BusinessProfile }>('/comptoir/register', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// Profile
export const getProfile = () => apiFetch<BusinessProfile>('/comptoir/profile');
export const updateProfile = (data: Partial<BusinessProfile>) =>
  apiFetch<BusinessProfile>('/comptoir/profile', { method: 'PUT', body: JSON.stringify(data) });

// Delivery Sites
export const listSites = () => apiFetch<DeliverySite[]>('/comptoir/sites');
export const createSite = (data: Omit<DeliverySite, 'id' | 'userId' | 'createdAt' | 'updatedAt'>) =>
  apiFetch<DeliverySite>('/comptoir/sites', { method: 'POST', body: JSON.stringify(data) });
export const updateSite = (id: string, data: Partial<DeliverySite>) =>
  apiFetch<DeliverySite>(`/comptoir/sites/${id}`, { method: 'PUT', body: JSON.stringify(data) });
export const deleteSite = (id: string) =>
  apiFetch<void>(`/comptoir/sites/${id}`, { method: 'DELETE' });

// Access
export const requestAccess = (bakeryId: string) =>
  apiFetch<void>(`/comptoir/access/request/${bakeryId}`, { method: 'POST' });
export const listApprovedBakeries = () => apiFetch<Bakery[]>('/comptoir/bakeries');

// Products & Config
export const getProducts = (bakeryId: string) =>
  apiFetch<Record<string, Product[]>>(`/comptoir/bakeries/${bakeryId}/products`);
export const getConfig = (bakeryId: string) =>
  apiFetch<B2BConfig>(`/comptoir/bakeries/${bakeryId}/config`);

// Checkout
export const checkout = (data: { bakeryId: string; deliverySiteId: string; items: { productId: string; quantity: number }[] }) =>
  apiFetch<Order>('/comptoir/checkout', { method: 'POST', body: JSON.stringify(data) });
export const editOrder = (orderId: string, items: { productId: string; quantity: number }[]) =>
  apiFetch<Order>(`/comptoir/orders/${orderId}`, { method: 'PUT', body: JSON.stringify({ items }) });
export const computePricing = (data: { bakeryId: string; items: { productId: string; quantity: number }[] }) =>
  apiFetch<B2BOrderPricing>('/comptoir/pricing', { method: 'POST', body: JSON.stringify(data) });

// Saved Lists
export const listSavedLists = (bakeryId: string) =>
  apiFetch<SavedList[]>(`/comptoir/lists?bakeryId=${bakeryId}`);
export const createSavedList = (data: { bakeryId: string; name: string; items: { productId: string; quantity: number }[] }) =>
  apiFetch<SavedList>('/comptoir/lists', { method: 'POST', body: JSON.stringify(data) });
export const deleteSavedList = (id: string) =>
  apiFetch<void>(`/comptoir/lists/${id}`, { method: 'DELETE' });

// Deliveries
export const listDeliveries = (params?: { bakeryId?: string; status?: string; dateFrom?: string; dateTo?: string; page?: number }) =>
  apiFetch<{ items: Order[]; page: number; pageSize: number; total: number }>(`/comptoir/deliveries?${new URLSearchParams(params as Record<string, string>)}`);
export const getLastOrder = (bakeryId: string) => apiFetch<Order>(`/comptoir/orders/${bakeryId}/last`);

// Invoices
export const listInvoices = (page?: number) =>
  apiFetch<{ items: B2BInvoice[]; page: number; pageSize: number; total: number }>(`/comptoir/invoices?page=${page ?? 1}`);
export const downloadInvoicePDF = (invoiceId: string) =>
  fetch(`/api/comptoir/invoices/${invoiceId}/pdf`, { headers: { Authorization: `Bearer ${localStorage.getItem('auth_token')}` } });
```

---

## Error Handling

| Scenario | HTTP Status | Error Code | Message |
|----------|-------------|------------|---------|
| Missing/invalid registration fields | 400 | `VALIDATION_ERROR` | Field-specific error messages |
| VAT/SIRET already registered | 409 | `COMPANY_ALREADY_EXISTS` | "Company is already registered" |
| Access not approved for bakery | 403 | `ACCESS_DENIED` | "B2B access not approved for this bakery" |
| Non-business role accessing B2B endpoints | 403 | `FORBIDDEN` | "Business role required" |
| Order total below minimum | 422 | `BELOW_MINIMUM` | "Order total X is below minimum Y" |
| Order submission after cutoff | 422 | `CUTOFF_PASSED` | "Cutoff time has passed for this bakery" |
| Edit attempt after cutoff | 422 | `CUTOFF_PASSED` | "Order can no longer be modified" |
| Delete only remaining delivery site | 422 | `LAST_SITE` | "At least one delivery site is required" |
| No delivery site before checkout | 422 | `NO_DELIVERY_SITE` | "A delivery site is required to place an order" |
| Access request already exists | 409 | `ACCESS_EXISTS` | "Access request already exists for this bakery" |
| Business profile not found | 404 | `PROFILE_NOT_FOUND` | "Business profile not found" |
| Delivery site not found | 404 | `SITE_NOT_FOUND` | "Delivery site not found" |
| Order not found | 404 | `ORDER_NOT_FOUND` | "Order not found" |
| Invoice not found | 404 | `INVOICE_NOT_FOUND` | "Invoice not found" |
| Unavailable product in order | 422 | `PRODUCT_UNAVAILABLE` | "Product X is no longer available" |

---

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: B2B Registration Validation

*For any* registration payload where one or more required fields (company name, VAT/SIRET, IBAN, billing email, billing contact name) are missing or violate their format constraints (max length, valid email format), the API SHALL return a 400 status code and the system SHALL NOT create any User or BusinessProfile record.

**Validates: Requirements 1.1, 1.3**

### Property 2: VAT/SIRET Immutability

*For any* authenticated Business_User and any profile update request, the VAT/SIRET value stored in the BusinessProfile SHALL remain identical to the value provided at registration time, regardless of the update payload content.

**Validates: Requirements 1.5**

### Property 3: Access Control Enforcement

*For any* (Business_User, Bakery) pair where the access status is NOT "approved", all requests to browse products or place orders for that bakery SHALL return a 403 status code. Conversely, *for any* pair where access IS "approved", product listing and checkout requests SHALL be permitted (subject to other validations).

**Validates: Requirements 3.6, 3.10, 4.5**

### Property 4: Cart Consistency Invariant

*For any* sequence of add, remove, and quantity-change operations on the B2B_Cart: (a) items SHALL be grouped by bakery_id with no empty groups, (b) each bakery group subtotal SHALL equal the sum of (quantity * unitPrice) for all items in that group, and (c) the overall cart total SHALL equal the sum of all bakery group subtotals.

**Validates: Requirements 5.1, 5.2, 5.3, 5.4**

### Property 5: Pricing Computation Correctness

*For any* set of order items and a bakery's B2B config (pro_discount percentage, TVA rate of 6%), the computed pricing SHALL satisfy: discount_amount = subtotal_ht * pro_discount / 100, tva_amount = (subtotal_ht - discount_amount) * 6 / 100, total_ttc = subtotal_ht - discount_amount + tva_amount, where subtotal_ht = sum of (quantity * unit_price) for all items.

**Validates: Requirements 14.2, 14.3, 14.4**

### Property 6: Cutoff Time Enforcement

*For any* B2B order creation or edit request, if the current time is after the bakery's configured cutoff_time, the API SHALL return a 422 status code and SHALL NOT create or modify the order. If the current time is before the cutoff_time, the request SHALL be processed (subject to other validations).

**Validates: Requirements 7.3, 7.4, 7.5, 9.5**

### Property 7: Order Minimum Validation

*For any* checkout request where the subtotal_ht (sum of quantity * unit_price for all items) is strictly less than the bakery's configured order_minimum, the API SHALL return a 422 status code with the minimum amount and SHALL NOT create the order.

**Validates: Requirements 6.6, 6.7**

### Property 8: B2B Role Authorization Gate

*For any* HTTP request to a `/api/comptoir/*` endpoint (excluding `/api/comptoir/register`) that does not carry a valid JWT with role value 3, the API SHALL return a 403 status code and SHALL NOT execute the requested operation.

**Validates: Requirements 15.2, 2.8, 8.7, 10.6**

---

## Testing Strategy

### Unit Tests (Go)

- **Service layer tests** (`internal/service/b2b_service_test.go`): Mock the `B2BRepository` and test all business logic — validation, access checks, cutoff enforcement, pricing computation, minimum validation.
- **Handler tests** (`internal/api/b2b_handler_test.go`): Test HTTP routing, request parsing, auth middleware, and response formatting. Use `httptest.NewRecorder`.
- **Pricing pure function tests**: The `ComputePricing` logic is a pure computation — test with many input combinations.

### Property-Based Tests (Go)

Using `pgregory.net/rapid` (Go property testing library):

- **Property 5 (Pricing)**: Generate random item lists (quantity 1-999, price 1-100000) and random discount rates (0-100). Assert the pricing formula invariants.
- **Property 1 (Validation)**: Generate random registration payloads with systematically invalid fields. Assert rejection.
- **Property 4 (Cart)**: Generate random sequences of cart operations. Assert group invariants hold after each operation.

### Integration Tests

- **Database tests**: Verify migrations apply cleanly, constraints work (unique VAT, FK cascades, check constraints).
- **Access control flow**: Register → request access → approve → browse products → checkout. Full lifecycle.
- **Cutoff enforcement**: Use injectable clock to test before/after cutoff behavior.

### Frontend Tests (Vitest + React Testing Library)

- **Cart logic** (`useB2BCart.test.ts`): Test add/remove/update operations, localStorage persistence round-trip, group subtotal calculations.
- **ComptoirNav**: Verify tabs render, site switcher appears with multiple sites.
- **CommandeRapide**: Verify grid renders products, quantity input updates cart.
- **Pricing display**: Verify summary shows correct breakdown, "Remise pro" hidden when zero.

### E2E Tests (Playwright)

- B2B registration flow
- Full ordering cycle: login → select bakery → add items → checkout → verify delivery list
- Cutoff behavior: verify disabled state after cutoff
