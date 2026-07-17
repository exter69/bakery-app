-- +goose Up
ALTER TABLE bakeries ADD COLUMN latitude DOUBLE PRECISION DEFAULT 0;
ALTER TABLE bakeries ADD COLUMN longitude DOUBLE PRECISION DEFAULT 0;

-- +goose Down
ALTER TABLE bakeries DROP COLUMN latitude;
ALTER TABLE bakeries DROP COLUMN longitude;
