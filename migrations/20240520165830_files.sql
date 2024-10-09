-- +goose Up
-- +goose StatementBegin
CREATE TYPE vp_backend_files_type_enum AS ENUM (
  'doc',
  'docx',
  'pdf',
  'png',
  'jpg',
  'jpeg'
);

CREATE TABLE files
(
  id                BIGSERIAL                       PRIMARY KEY,
  user_id           BIGINT                          NOT NULL,
  parent_id         BIGINT                          NULL,
  attachment_id     VARCHAR(50)                     NOT NULL,
  created_at        TIMESTAMP                       DEFAULT NOW() NOT NULL,
  updated_at        TIMESTAMP                       DEFAULT NOW() NOT NULL,
  deleted_at        TIMESTAMP                       NULL,
  name              VARCHAR(50)                     NOT NULL,
  original_name     VARCHAR(255)                    NOT NULL,
  type              vp_backend_files_type_enum      NOT NULL,
  ocr               JSON                            NULL,
  CONSTRAINT        fk_files_user                   FOREIGN KEY(user_id) REFERENCES users(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS files DROP CONSTRAINT IF EXISTS fk_files_user;
DROP TABLE IF EXISTS files CASCADE;

DROP TYPE vp_backend_files_type_enum;
-- +goose StatementEnd
