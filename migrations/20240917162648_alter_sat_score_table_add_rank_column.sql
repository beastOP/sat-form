-- +goose Up
-- +goose StatementBegin
ALTER TABLE sat_scores ADD COLUMN rank INTEGER DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sat_scores DROP COLUMN rank;
-- +goose StatementEnd
