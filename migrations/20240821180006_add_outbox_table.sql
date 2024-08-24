-- +goose Up
-- +goose StatementBegin
create table outbox
(
    id bigserial not null,
    type integer not null,
    message_key text not null,
    message_value text not null,
    is_sent boolean not null,
    primary key (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table outbox;
-- +goose StatementEnd
