-- +goose Up
ALTER TABLE users ADD COLUMN stripe_customer_id VARCHAR(255);

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
