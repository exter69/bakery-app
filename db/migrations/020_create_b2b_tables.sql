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
    order_minimum BIGINT NOT NULL DEFAULT 2000,
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

CREATE UNIQUE INDEX idx_b2b_invoices_bakery_number ON b2b_invoices(bakery_id, invoice_number);
CREATE INDEX idx_b2b_invoices_business_profile ON b2b_invoices(business_profile_id);
CREATE INDEX idx_b2b_invoices_order ON b2b_invoices(order_id);

-- Add delivery_site_id and B2B pricing columns to orders
ALTER TABLE orders ADD COLUMN delivery_site_id UUID REFERENCES delivery_sites(id);
ALTER TABLE orders ADD COLUMN subtotal_ht BIGINT;
ALTER TABLE orders ADD COLUMN discount_amount BIGINT DEFAULT 0;
ALTER TABLE orders ADD COLUMN tva_amount BIGINT;

-- Add "on_invoice" to the payment_method CHECK constraint on orders
-- Drop existing constraint if any and re-add with the new value
DO $$
BEGIN
    -- Try to drop a named constraint (common pattern)
    BEGIN
        ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
    EXCEPTION WHEN undefined_object THEN
        NULL;
    END;
    -- Add the new constraint allowing all payment methods
    ALTER TABLE orders ADD CONSTRAINT orders_payment_method_check
        CHECK (payment_method IN ('online', 'on_spot', 'on_invoice'));
END$$;

-- +goose Down
-- Remove payment_method constraint change
DO $$
BEGIN
    BEGIN
        ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
    EXCEPTION WHEN undefined_object THEN
        NULL;
    END;
    ALTER TABLE orders ADD CONSTRAINT orders_payment_method_check
        CHECK (payment_method IN ('online', 'on_spot'));
END$$;

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
