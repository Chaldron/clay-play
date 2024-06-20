-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS event (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    capacity INTEGER NOT NULL,
    start DATETIME NOT NULL,
    location TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    creator_id TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    group_id TEXT
);

CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL
    password TEXT NOT NULL
    created_at DATETIME NOT NULL,
    isadmin BOOL NOT NULL DEFAULT 0,
    picture TEXT
);


CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expiry REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);

CREATE TABLE IF NOT EXISTS event_response (
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    attendee_count INT NOT NULL DEFAULT 0,
    on_waitlist BOOL NOT NULL DEFAULT 0,
    PRIMARY KEY (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS user_group (
    id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    creator_id TEXT NOT NULL,
    is_deleted BOOL NOT NULL DEFAULT 0,
    name TEXT NOT NULL,
    invite_id TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_group_member (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    PRIMARY KEY (group_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event;
DROP TABLE IF EXISTS user;
DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS sessions_expiry_idx;
DROP TABLE IF EXISTS event_response;
DROP TABLE IF EXISTS user_group;
DROP TABLE IF EXISTS user_group_member;
-- +goose StatementEnd
