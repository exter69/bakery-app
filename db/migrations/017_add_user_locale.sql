-- +goose Up
ALTER TABLE users ADD COLUMN locale VARCHAR(5) DEFAULT 'en';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS locale;
