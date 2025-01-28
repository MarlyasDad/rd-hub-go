-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGSERIAL NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP NULL,
    first_name VARCHAR(50) NULL,
    last_name VARCHAR(50) NULL,
    username VARCHAR(50) NULL,
    language_code VARCHAR(5) NULL,
    admin BOOLEAN DEFAULT FALSE NOT NULL,
    execution BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT users_tg_unique UNIQUE (tg_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
