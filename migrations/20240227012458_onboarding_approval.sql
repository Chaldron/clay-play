-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_review (
    user_id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    reviewed_at DATETIME,
    comment TEXT,
    is_approved BOOL NOT NULL DEFAULT 0
);

ALTER TABLE user
ADD COLUMN status INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_review;

ALTER TABLE user DROP COLUMN status;
-- +goose StatementEnd
