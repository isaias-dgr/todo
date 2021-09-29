-- +goose Up
-- +goose StatementBegin
CREATE TABLE task (
  id BINARY(16) NOT NULL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL ,
  updated_at TIMESTAMP NOT NULL 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE task;
-- +goose StatementEnd
