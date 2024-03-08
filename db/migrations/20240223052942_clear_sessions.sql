-- +goose Up
-- +goose StatementBegin
DELETE FROM sessions;
-- +goose StatementEnd
