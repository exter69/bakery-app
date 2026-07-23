-- +goose Up

-- Volume tier definitions (platform-wide, configurable)
CREATE TABLE volume_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    min_monthly_spend BIGINT NOT NULL, -- cents HT, minimum monthly spend to qualify
    discount_percent INT NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed default tiers
INSERT INTO volume_tiers (min_monthly_spend, discount_percent) VALUES
    (150000, 8),   -- 1500 EUR/month -> 8% volume discount
    (200000, 10);  -- 2000 EUR/month -> 10% volume discount

-- Add per-account pro discount (default 1%, overrides the bakery-level b2b_config.pro_discount)
ALTER TABLE business_profiles ADD COLUMN pro_discount INT NOT NULL DEFAULT 1 CHECK (pro_discount >= 0 AND pro_discount <= 100);

-- Add VAT rate to b2b_config (default 6% for Belgium)
ALTER TABLE b2b_config ADD COLUMN vat_rate INT NOT NULL DEFAULT 6 CHECK (vat_rate >= 0 AND vat_rate <= 100);

-- Track monthly spend per business account (rolling, recalculated)
ALTER TABLE business_profiles ADD COLUMN current_month_spend BIGINT NOT NULL DEFAULT 0;
ALTER TABLE business_profiles ADD COLUMN spend_month VARCHAR(7) NOT NULL DEFAULT to_char(NOW(), 'YYYY-MM');

-- +goose Down
ALTER TABLE business_profiles DROP COLUMN IF EXISTS spend_month;
ALTER TABLE business_profiles DROP COLUMN IF EXISTS current_month_spend;
ALTER TABLE b2b_config DROP COLUMN IF EXISTS vat_rate;
ALTER TABLE business_profiles DROP COLUMN IF EXISTS pro_discount;
DROP TABLE IF EXISTS volume_tiers;
