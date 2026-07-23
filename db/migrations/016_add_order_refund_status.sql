-- +goose Up
ALTER TABLE orders ADD COLUMN refund_status VARCHAR(20) DEFAULT '';

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS refund_status;
