-- Create recurring_orders table
CREATE TABLE IF NOT EXISTS recurring_orders (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id),
    bakery_id      TEXT NOT NULL REFERENCES bakeries(id),
    items          JSONB NOT NULL DEFAULT '[]',
    scheduled_day  TEXT NOT NULL,
    scheduled_time JSONB NOT NULL,
    frequency      TEXT NOT NULL DEFAULT 'weekly',
    selection_mode TEXT NOT NULL DEFAULT 'fixed',
    active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_orders_user_id ON recurring_orders(user_id);
CREATE INDEX idx_recurring_orders_active ON recurring_orders(active) WHERE active = TRUE;
