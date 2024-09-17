-- +goose Up
-- +goose StatementBegin
CREATE TABLE sat_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    address TEXT NOT NULL,
    city TEXT NOT NULL,
    country TEXT NOT NULL,
    pincode TEXT NOT NULL,
    sat_score INTEGER NOT NULL,
    passed BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sat_scores;
-- +goose StatementEnd
