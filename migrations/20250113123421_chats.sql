-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats (
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGSERIAL NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP NULL,
    active BOOLEAN DEFAULT FALSE NOT NULL,
    public BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT chats_tg_unique UNIQUE (tg_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chats CASCADE;
-- +goose StatementEnd
