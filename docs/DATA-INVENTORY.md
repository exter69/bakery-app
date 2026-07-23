# Data Inventory — Mie & Beurre

This document inventories all personal data collected, processed, and stored by the Mie & Beurre platform, in accordance with GDPR Article 30 (Records of Processing Activities).

## 1. Personal Data Collected

### User Accounts
| Field | Purpose | Basis |
|-------|---------|-------|
| Username | Account identification | Contract |
| Email (contactEmail) | Transactional notifications | Contract |
| Password hash (bcrypt) | Authentication | Contract |
| Role | Access control | Contract |
| Locale | UI language preference | Consent |
| Favorite products | Personalization | Consent |
| Holiday mode dates | Service scheduling | Contract |
| Stripe Customer ID | Payment processing | Contract |

### Orders & Reservations
| Field | Purpose | Basis |
|-------|---------|-------|
| Bakery reference | Service fulfillment | Contract |
| Items (products, quantities) | Service fulfillment | Contract |
| Amounts (in cents) | Payment & accounting | Contract |
| Timestamps (created, updated) | Audit trail | Legitimate interest |
| Status | Service tracking | Contract |
| Payment method | Payment processing | Contract |

### Reviews
| Field | Purpose | Basis |
|-------|---------|-------|
| Bakery reference | Content association | Contract |
| Rating (1-5) | Public feedback | Consent |
| Text content | Public feedback | Consent |
| Hidden flag | Content moderation | Legitimate interest |

### Geolocation
| Field | Purpose | Basis |
|-------|---------|-------|
| Browser location (lat/lng) | Sorting bakeries by proximity | Consent |

**Note:** Geolocation is used client-side only and is NOT stored on our servers.

### Social Logins
| Field | Purpose | Basis |
|-------|---------|-------|
| Provider (google/apple) | Authentication method | Contract |
| Provider user ID | Identity linking | Contract |
| Email | Account recovery | Contract |

### B2B Profiles (Business Accounts)
| Field | Purpose | Basis |
|-------|---------|-------|
| Company name | Invoicing | Contract |
| VAT/SIRET number | Fiscal compliance | Legal obligation |
| IBAN | Payment processing | Contract |
| Billing email | Invoice delivery | Contract |
| Billing contact name | Communication | Contract |
| Delivery sites (address, instructions) | Service fulfillment | Contract |

## 2. Data Processors (Sub-processors)

| Processor | Purpose | Data Shared | Location |
|-----------|---------|-------------|----------|
| Stripe | Payment processing | Customer ID, payment amounts | EU/US (SCCs) |
| Railway | Application hosting | All application data (encrypted at rest) | EU |
| SMTP Provider | Transactional email | Email addresses, order confirmations | EU |
| Sentry | Error tracking | Stack traces (PII scrubbed before send) | EU |

## 3. Data Retention

| Data Category | Retention Period | Deletion Method |
|---------------|-----------------|-----------------|
| Active accounts | Until account deletion | Anonymization |
| Deleted accounts | Immediate anonymization | Username → "deleted-{id}", email/password cleared; JWT invalidated via middleware check |
| Order history | Indefinite (anonymized on account deletion) | User reference anonymized |
| Reviews | Indefinite (anonymized on account deletion) | User reference anonymized |
| Recurring orders | Until account deletion | Hard delete |
| B2B profiles | Until account deletion | Hard delete (profile + saved lists) |
| Delivery sites | Until account deletion | Hard delete |
| Social logins | Until account deletion | Hard delete |
| Push subscriptions | Until account deletion | Cleared from memory |
| Stripe Customer | Until account deletion | Deleted via Stripe API (best-effort) |

## 4. Data Subject Rights Implementation

| Right | Implementation | Endpoint |
|-------|---------------|----------|
| Right of access (Art. 15) | Data export as JSON | `GET /api/user/data-export` |
| Right to rectification (Art. 16) | Profile update | `PUT /api/user/profile` |
| Right to erasure (Art. 17) | Account anonymization | `DELETE /api/user/account` |
| Right to data portability (Art. 20) | JSON export download | `GET /api/user/data-export` |
| Right to object (Art. 21) | Contact via email | privacy@mieetbeurre.com |

## 5. Security Measures

- Passwords stored as bcrypt hashes (never plaintext)
- JWT-based authentication with short-lived tokens
- HTTPS enforced in production (Railway + Vercel)
- CORS restricted to frontend origin
- Rate limiting on authentication endpoints
- Security headers (CSP, HSTS, X-Frame-Options)
- PII scrubbed from error tracking (Sentry)
- No sensitive data in server logs

## 6. Cookies

| Cookie/Storage | Type | Purpose |
|----------------|------|---------|
| auth_token (localStorage) | Essential | JWT session token |
| cookie_consent (localStorage) | Essential | Consent preference |
| guest_mode (localStorage) | Essential | Guest browsing flag |
| theme (localStorage) | Functional | UI theme preference |
| locale (localStorage) | Functional | Language preference |

No third-party tracking cookies are used. Analytics (if enabled) only loads with explicit "Accept all" consent.
