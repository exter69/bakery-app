-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    photo_url VARCHAR(500),
    category VARCHAR(100) NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_bakery_id ON products(bakery_id);
CREATE INDEX idx_products_bakery_category ON products(bakery_id, category);

-- +goose Down
DROP TABLE IF EXISTS products;
