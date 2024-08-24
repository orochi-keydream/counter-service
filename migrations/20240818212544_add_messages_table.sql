-- +goose Up
-- +goose StatementBegin
create table messages
(
    message_id bigint not null,
    user_id text not null,
    chat_id text not null,
    is_read boolean not null,
    primary key (message_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table messages
-- +goose StatementEnd
