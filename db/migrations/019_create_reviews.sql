-- +goose Up

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bakery_id UUID NOT NULL REFERENCES bakeries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text TEXT,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_reviews_user_bakery UNIQUE (user_id, bakery_id)
);

CREATE INDEX idx_reviews_bakery_visible ON reviews(bakery_id) WHERE hidden = FALSE;
CREATE INDEX idx_reviews_user ON reviews(user_id);

CREATE TABLE review_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_reports_user_review UNIQUE (reporter_id, review_id)
);

ALTER TABLE bakeries
    ADD COLUMN rating_avg NUMERIC(2,1),
    ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE bakeries DROP COLUMN IF EXISTS rating_count;
ALTER TABLE bakeries DROP COLUMN IF EXISTS rating_avg;
DROP TABLE IF EXISTS review_reports;
DROP TABLE IF EXISTS reviews;
