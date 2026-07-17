-- +goose Up
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    reservation_id UUID REFERENCES reservations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 1 AND quantity <= 999),
    unit_price DECIMAL(10, 2) NOT NULL CHECK (unit_price > 0),
    subtotal DECIMAL(10, 2) NOT NULL CHECK (subtotal > 0),
    CHECK (
        (order_id IS NOT NULL AND reservation_id IS NULL)
        OR (order_id IS NULL AND reservation_id IS NOT NULL)
    )
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_reservation_id ON order_items(reservation_id);

-- +goose Down
DROP TABLE IF EXISTS order_items;
