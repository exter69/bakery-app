-- +goose Up
CREATE TABLE registration_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    bakery_name VARCHAR(200),
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_registration_tokens_token ON registration_tokens(token);

ALTER TABLE users ADD COLUMN contact_email VARCHAR(255) DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN contact_email;
DROP TABLE IF EXISTS registration_tokens;
