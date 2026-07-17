-- +goose Up
ALTER TABLE users ADD COLUMN favorite_products TEXT[] DEFAULT '{}';

-- +goose Down
ALTER TABLE users DROP COLUMN favorite_products;
