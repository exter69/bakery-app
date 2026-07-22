-- +goose Up
ALTER TABLE products
  ADD COLUMN allergens TEXT[] NOT NULL DEFAULT '{}',
  ADD COLUMN health_score INTEGER NULL;

ALTER TABLE products
  ADD CONSTRAINT chk_health_score
  CHECK (health_score IS NULL OR (health_score >= 1 AND health_score <= 5));

-- +goose Down
ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_health_score;
ALTER TABLE products DROP COLUMN IF EXISTS health_score;
ALTER TABLE products DROP COLUMN IF EXISTS allergens;
