-- +goose Up
-- +goose StatementBegin
ALTER TABLE user
    ADD COLUMN email TEXT NOT NULL DEFAULT '';
ALTER TABLE user
    ADD COLUMN password TEXT NOT NULL DEFAULT '';
ALTER TABLE user
    ADD COLUMN isadmin BOOL NOT NULL DEFAULT 0;
ALTER TABLE user
    DROP COLUMN external_id;
ALTER TABLE user
    DROP COLUMN status;
-- +goose StatementEnd
