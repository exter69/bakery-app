-- +goose Up
ALTER TABLE orders ADD COLUMN payment_intent_id VARCHAR(255);

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS payment_intent_id;
