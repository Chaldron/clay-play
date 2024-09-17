-- +goose Up
-- +goose StatementBegin
ALTER TABLE event ADD COLUMN description TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event DROP COLUMN description;
-- +goose StatementEnd
