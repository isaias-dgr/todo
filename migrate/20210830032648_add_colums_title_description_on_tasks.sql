-- +goose Up
-- +goose StatementBegin
ALTER TABLE task
ADD COLUMN title varchar(255) AFTER id,
ADD COLUMN description varchar(255) AFTER title;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task
DROP COLUMN title,
DROP COLUMN description;
-- +goose StatementEnd
