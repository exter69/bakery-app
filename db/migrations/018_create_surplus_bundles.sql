-- +goose Up

-- Surplus bundles table
CREATE TABLE surplus_bundles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('compose', 'surprise')),
    photo_url VARCHAR(500),
    description VARCHAR(200),
    estimated_value BIGINT,
    original_price BIGINT NOT NULL CHECK (original_price > 0),
    discounted_price BIGINT NOT NULL CHECK (discounted_price > 0),
    quantity_total INT NOT NULL CHECK (quantity_total >= 1),
    quantity_remaining INT NOT NULL CHECK (quantity_remaining >= 0),
    pickup_start_time TIME NOT NULL,
    pickup_end_time TIME NOT NULL,
    published_date DATE,
    expires_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'expired', 'sold_out')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_bundle_discount CHECK (discounted_price < original_price),
    CONSTRAINT chk_bundle_pickup_window CHECK (pickup_start_time < pickup_end_time),
    CONSTRAINT chk_bundle_stock CHECK (quantity_remaining <= quantity_total)
);

CREATE INDEX idx_surplus_bundles_bakery_id ON surplus_bundles(bakery_id);
CREATE INDEX idx_surplus_bundles_status ON surplus_bundles(status);
CREATE INDEX idx_surplus_bundles_status_expires ON surplus_bundles(status, expires_at)
    WHERE status = 'published';
CREATE INDEX idx_surplus_bundles_published_date ON surplus_bundles(published_date DESC)
    WHERE status = 'published';

-- Bundle items for composé bundles
CREATE TABLE surplus_bundle_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bundle_id UUID NOT NULL REFERENCES surplus_bundles(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    description VARCHAR(200) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 1)
);

CREATE INDEX idx_surplus_bundle_items_bundle_id ON surplus_bundle_items(bundle_id);

-- Bundle reservations
CREATE TABLE bundle_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bundle_id UUID NOT NULL REFERENCES surplus_bundles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'picked_up', 'released', 'cancelled')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bundle_reservations_bundle_id ON bundle_reservations(bundle_id);
CREATE INDEX idx_bundle_reservations_user_id ON bundle_reservations(user_id);
CREATE INDEX idx_bundle_reservations_status ON bundle_reservations(status);

-- Enforce max one active reservation per customer per bundle
CREATE UNIQUE INDEX idx_bundle_reservations_active_unique
    ON bundle_reservations(user_id, bundle_id)
    WHERE status IN ('pending', 'confirmed');

-- +goose Down
DROP TABLE IF EXISTS bundle_reservations;
DROP TABLE IF EXISTS surplus_bundle_items;
DROP TABLE IF EXISTS surplus_bundles;
