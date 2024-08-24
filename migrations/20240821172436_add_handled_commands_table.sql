-- +goose Up
-- +goose StatementBegin
create table handled_commands
(
    correlation_id text primary key
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table handled_commands;
-- +goose StatementEnd
