-- +goose Up
-- +goose StatementBegin
ALTER TABLE task
ADD INDEX titleIndex (title);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task
DROP INDEX titleIndex;
-- +goose StatementEnd
