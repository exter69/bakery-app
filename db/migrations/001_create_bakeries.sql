-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bakeries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    photo_url VARCHAR(500),
    description TEXT,
    address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE day_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    open_time TIME,
    close_time TIME,
    is_open BOOLEAN NOT NULL DEFAULT true,
    UNIQUE (bakery_id, day_of_week),
    CHECK (is_open = false OR (open_time IS NOT NULL AND close_time IS NOT NULL AND open_time < close_time))
);

CREATE INDEX idx_day_schedules_bakery_id ON day_schedules(bakery_id);

-- +goose Down
DROP TABLE IF EXISTS day_schedules;
DROP TABLE IF EXISTS bakeries;
