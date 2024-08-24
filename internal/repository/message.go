package repository

import (
	"context"
	"database/sql"
	"github.com/orochi-keydream/counter-service/internal/model"
)

type MessageRepository struct {
	conn *sql.DB
}

func NewMessageRepository(conn *sql.DB) *MessageRepository {
	return &MessageRepository{conn: conn}
}

func (r *MessageRepository) CountUnreadByUser(
	ctx context.Context,
	userId model.UserId,
	tx *sql.Tx,
) (int, error) {
	const query = `
		select count(*)
		from messages
		where user_id = $1 and is_read = false;`

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.conn
	}

	row := ec.QueryRowContext(ctx, query, userId)

	if row.Err() != nil {
		return 0, row.Err()
	}

	var count int

	err := row.Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *MessageRepository) CountUnreadByUserAndChat(
	ctx context.Context,
	userId model.UserId,
	chatId model.ChatId,
	tx *sql.Tx,
) (int, error) {
	const query = `
		select count(*)
		from messages
		where user_id = $1 and chat_id = $2 and is_read = false;`

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.conn
	}

	row := ec.QueryRowContext(ctx, query, userId, chatId)

	if row.Err() != nil {
		return 0, row.Err()
	}

	var count int

	err := row.Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *MessageRepository) GetForUser(
	ctx context.Context,
	userId model.UserId,
	messageIds []model.MessageId,
	tx *sql.Tx,
) ([]*model.Message, error) {
	const query = `
		select
			message_id,
			chat_id,
			user_id,
			is_read
		from messages
		where user_id = $1 and message_id = any ($2);`

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.conn
	}

	rows, err := ec.QueryContext(ctx, query, userId, messageIds)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []*model.Message

	for rows.Next() {
		dto := struct {
			MessageId int64
			ChatId    string
			UserId    string
			IsRead    bool
		}{}

		err = rows.Scan(&dto.MessageId, &dto.ChatId, &dto.UserId, &dto.IsRead)

		if err != nil {
			return nil, err
		}

		message := model.Message{
			MessageId: model.MessageId(dto.MessageId),
			ChatId:    model.ChatId(dto.ChatId),
			UserId:    model.UserId(dto.UserId),
			IsRead:    false,
		}

		messages = append(messages, &message)
	}

	return messages, nil
}

func (r *MessageRepository) Update(
	ctx context.Context,
	messages []*model.Message,
	tx *sql.Tx,
) error {
	const query = `
		update messages m
		set is_read = t.is_read
		from unnest (
			$1::bigint[],
			$2::text[],
			$3::boolean[]
		) as t(message_id, user_id, is_read)
		where m.message_id = t.message_id and m.user_id = t.user_id;`

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.conn
	}

	messageIdArray := make([]int64, len(messages))
	userIdArray := make([]string, len(messages))
	isReadArray := make([]bool, len(messages))

	for i, message := range messages {
		messageIdArray[i] = int64(message.MessageId)
		userIdArray[i] = string(message.UserId)
		isReadArray[i] = message.IsRead
	}

	_, err := ec.ExecContext(ctx, query, messageIdArray, userIdArray, isReadArray)

	return err
}

func (r *MessageRepository) Add(ctx context.Context, message *model.Message, tx *sql.Tx) error {
	const query = `
		insert into messages
		(
			message_id,
			user_id,
			chat_id,
			is_read
		)
		values
		(
			$1,
			$2,
			$3,
			$4
		);`

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.conn
	}

	_, err := ec.ExecContext(ctx, query, message.MessageId, message.UserId, message.ChatId, message.IsRead)

	// TODO: Check if uniqueness is violated.

	return err
}
