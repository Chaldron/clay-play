-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS audit_log (
    user_id TEXT NOT NULL,
    recorded_at DATETIME NOT NULL,
    description TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log;
-- +goose StatementEnd
