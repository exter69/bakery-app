-- +goose Up
-- Add 'capturing' to the orders.status CHECK constraint for the capture-flow atomicity fix.
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_status_check,
    ADD CONSTRAINT orders_status_check
        CHECK (status IN ('pending_payment', 'confirmed', 'preparing', 'ready', 'capturing', 'delivered', 'cancelled'));

-- +goose Down
-- Remove 'capturing' from the orders.status CHECK constraint.
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_status_check,
    ADD CONSTRAINT orders_status_check
        CHECK (status IN ('pending_payment', 'confirmed', 'preparing', 'ready', 'delivered', 'cancelled'));
