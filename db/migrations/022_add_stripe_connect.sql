-- +goose Up
-- Add Stripe Connect account ID to bakeries
ALTER TABLE bakeries ADD COLUMN stripe_connect_id VARCHAR(255);

-- Platform commission rate (default 10%)
ALTER TABLE bakeries ADD COLUMN commission_rate INT NOT NULL DEFAULT 10 CHECK (commission_rate >= 0 AND commission_rate <= 100);

-- Payout tracking
CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,          -- bakery's share in cents
    commission BIGINT NOT NULL,      -- platform fee in cents
    stripe_transfer_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'transferred', 'failed', 'refunded')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    transferred_at TIMESTAMP
);

CREATE INDEX idx_payouts_order ON payouts(order_id);
CREATE INDEX idx_payouts_bakery ON payouts(bakery_id);
CREATE INDEX idx_payouts_status ON payouts(bakery_id, status);

-- +goose Down
DROP TABLE IF EXISTS payouts;
ALTER TABLE bakeries DROP COLUMN IF EXISTS commission_rate;
ALTER TABLE bakeries DROP COLUMN IF EXISTS stripe_connect_id;
