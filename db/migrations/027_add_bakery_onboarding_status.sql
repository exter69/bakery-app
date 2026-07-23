-- +goose Up
ALTER TABLE bakeries ADD COLUMN charges_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE bakeries ADD COLUMN payouts_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE bakeries DROP COLUMN IF EXISTS payouts_enabled;
ALTER TABLE bakeries DROP COLUMN IF EXISTS charges_enabled;
