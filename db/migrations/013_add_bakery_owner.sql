-- +goose Up
ALTER TABLE bakeries ADD COLUMN owner_id UUID;
CREATE INDEX idx_bakeries_owner_id ON bakeries(owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_bakeries_owner_id;
ALTER TABLE bakeries DROP COLUMN IF EXISTS owner_id;
