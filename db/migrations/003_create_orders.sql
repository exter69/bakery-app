-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL,
    scheduled_day INT NOT NULL CHECK (scheduled_day >= 0 AND scheduled_day <= 6),
    scheduled_start_time TIME NOT NULL,
    scheduled_end_time TIME NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_payment'
        CHECK (status IN ('pending_payment', 'confirmed', 'preparing', 'ready', 'delivered', 'cancelled')),
    total_amount DECIMAL(10, 2) NOT NULL CHECK (total_amount >= 0),
    payment_method VARCHAR(20) NOT NULL
        CHECK (payment_method IN ('online', 'on_spot')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (scheduled_start_time < scheduled_end_time)
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_bakery_id ON orders(bakery_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);

-- +goose Down
DROP TABLE IF EXISTS orders;
