-- +goose Up
-- +goose StatementBegin

CREATE TYPE vp_backend_users_status_enum AS ENUM (
  'active', 
  'disabled'
);

CREATE TABLE users
(
  id                BIGSERIAL                       PRIMARY KEY,
  keycloak_id       uuid                            NOT NULL,
  created_at        TIMESTAMP                       DEFAULT NOW() NOT NULL,
  updated_at        TIMESTAMP                       DEFAULT NOW() NOT NULL,
  deleted_at        TIMESTAMP                       NULL,
  email             VARCHAR(50)                     NULL,
  status            vp_backend_users_status_enum    DEFAULT 'active'::vp_backend_users_status_enum NOT NULL
);

CREATE INDEX  ix_users_keycloak_id ON users (keycloak_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;

DROP TYPE vp_backend_users_status_enum;
-- +goose StatementEnd
