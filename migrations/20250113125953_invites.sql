-- +goose Up
-- +goose StatementBegin
CREATE TABLE invites (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGSERIAL NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP NULL,
    expiration_at TIMESTAMP NULL,
    admin BOOLEAN DEFAULT FALSE NOT NULL,
    execution BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT users_tg_unique UNIQUE (tg_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS invites CASCADE;
-- +goose StatementEnd
