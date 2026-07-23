-- +goose Up
-- Convert all money columns from DECIMAL(10,2) to BIGINT (storing cents).
-- The USING clause multiplies existing decimal values by 100 to convert to cents.
-- CHECK constraints (price > 0, unit_price > 0, subtotal > 0, total_amount >= 0)
-- remain valid for BIGINT and are preserved automatically by Postgres.

ALTER TABLE products
    ALTER COLUMN price TYPE BIGINT USING (price * 100)::BIGINT;

ALTER TABLE orders
    ALTER COLUMN total_amount TYPE BIGINT USING (total_amount * 100)::BIGINT;

ALTER TABLE order_items
    ALTER COLUMN unit_price TYPE BIGINT USING (unit_price * 100)::BIGINT,
    ALTER COLUMN subtotal TYPE BIGINT USING (subtotal * 100)::BIGINT;

ALTER TABLE reservations
    ALTER COLUMN total_amount TYPE BIGINT USING (total_amount * 100)::BIGINT;

-- +goose Down
-- Reverse: convert BIGINT cents back to DECIMAL(10,2) by dividing by 100.

ALTER TABLE products
    ALTER COLUMN price TYPE DECIMAL(10,2) USING (price / 100.0)::DECIMAL(10,2);

ALTER TABLE orders
    ALTER COLUMN total_amount TYPE DECIMAL(10,2) USING (total_amount / 100.0)::DECIMAL(10,2);

ALTER TABLE order_items
    ALTER COLUMN unit_price TYPE DECIMAL(10,2) USING (unit_price / 100.0)::DECIMAL(10,2),
    ALTER COLUMN subtotal TYPE DECIMAL(10,2) USING (subtotal / 100.0)::DECIMAL(10,2);

ALTER TABLE reservations
    ALTER COLUMN total_amount TYPE DECIMAL(10,2) USING (total_amount / 100.0)::DECIMAL(10,2);
