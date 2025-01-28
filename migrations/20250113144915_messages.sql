-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGSERIAL NOT NULL,
    tg_user_id BIGSERIAL NOT NULL,
    tg_chat_id BIGSERIAL NOT NULL,
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP NULL,
    text TEXT NOT NULL,
    amount DOUBLE PRECISION NULL,
    comment TEXT NULL,
    CONSTRAINT messages_chat_unique UNIQUE (message_id, chat_id),
    CONSTRAINT messages_users_tg_id_foreign FOREIGN KEY (tg_user_id) REFERENCES public.users(tg_id) ON DELETE CASCADE,
    CONSTRAINT messages_chats_tg_id_foreign FOREIGN KEY (tg_chat_id) REFERENCES public.chats(tg_id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages CASCADE;
-- +goose StatementEnd
