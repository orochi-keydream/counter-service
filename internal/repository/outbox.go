package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/orochi-keydream/counter-service/internal/model"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{
		db: db,
	}
}

func (r *OutboxRepository) Add(ctx context.Context, message *model.OutboxMessage, tx *sql.Tx) error {
	const query = "insert into outbox (type, message_key, message_value, is_sent) values ($1, $2, $3, $4)"

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.db
	}

	messageKeyBytes, err := toMessageKeyBytes(message.MessageKey, message.Type)

	if err != nil {
		return err
	}

	messageValueBytes, err := toMessageValueBytes(message.MessageValue, message.Type)

	if err != nil {
		return err
	}

	dto := outboxMessageDto{
		MessageType:  int(message.Type),
		MessageKey:   string(messageKeyBytes),
		MessageValue: string(messageValueBytes),
		IsSent:       false,
	}

	_, err = ec.ExecContext(ctx, query, dto.MessageType, dto.MessageKey, dto.MessageValue, dto.IsSent)

	return err
}

func (r *OutboxRepository) GetUnsent(ctx context.Context, tx *sql.Tx) ([]*model.OutboxMessage, error) {
	const query = "select id, type, message_key, message_value, is_sent from outbox where is_sent = false"

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.db
	}

	rows, err := ec.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	messages := make([]*model.OutboxMessage, 0)

	for rows.Next() {
		dto := outboxMessageDto{}

		err = rows.Scan(&dto.Id, &dto.MessageType, &dto.MessageKey, &dto.MessageValue, &dto.IsSent)

		if err != nil {
			return nil, err
		}

		messageType := model.OutboxMessageType(dto.MessageType)

		messageKey, err := fromMessageKeyBytes([]byte(dto.MessageKey), messageType)

		if err != nil {
			return nil, err
		}

		messageValue, err := fromMessageValueBytes([]byte(dto.MessageValue), messageType)

		if err != nil {
			return nil, err
		}

		message := &model.OutboxMessage{
			Id:           dto.Id,
			Type:         messageType,
			MessageKey:   messageKey,
			MessageValue: messageValue,
			IsSent:       dto.IsSent,
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (r *OutboxRepository) Update(ctx context.Context, messages []*model.OutboxMessage, tx *sql.Tx) error {
	const query = "update outbox set is_sent = $1 where id = any ($2)"

	var ec IExecutionContext

	if tx != nil {
		ec = tx
	} else {
		ec = r.db
	}

	messageIds := make([]int64, len(messages))

	for i, message := range messages {
		messageIds[i] = message.Id
	}

	_, err := ec.ExecContext(ctx, query, true, messageIds)

	return err
}

func toMessageKeyBytes(key any, messageType model.OutboxMessageType) ([]byte, error) {
	switch messageType {
	case model.OutboxMessageTypeRollbackMessage:
		s, ok := key.(model.ChatId)

		if !ok {
			return nil, fmt.Errorf("failed to parse message key")
		}

		return []byte(s), nil
	case model.OutboxMessageTypeCommitMessage:
		s, ok := key.(model.ChatId)

		if !ok {
			return nil, fmt.Errorf("failed to parse message key")
		}

		return []byte(s), nil
	default:
		return nil, fmt.Errorf("unsupported message type")
	}
}

func fromMessageKeyBytes(key []byte, messageType model.OutboxMessageType) (any, error) {
	switch messageType {
	case model.OutboxMessageTypeRollbackMessage:
		return string(key), nil
	case model.OutboxMessageTypeCommitMessage:
		return string(key), nil
	default:
		return nil, fmt.Errorf("unsupported message type")
	}
}

func toMessageValueBytes(payload any, messageType model.OutboxMessageType) ([]byte, error) {
	switch messageType {
	case model.OutboxMessageTypeRollbackMessage:
		return mapRollbackMessage(payload.(model.RollbackMessage))
	case model.OutboxMessageTypeCommitMessage:
		return mapCommitMessage(payload.(model.CommitMessage))
	default:
		err := fmt.Errorf("unsupported message type")
		return nil, err
	}
}

func mapRollbackMessage(payload model.RollbackMessage) ([]byte, error) {
	jsonDto := rollbackMessagePayloadDto{
		CorrelationId: string(payload.CorrelationId),
		MessageId:     int64(payload.MessageId),
	}

	return json.Marshal(jsonDto)
}

func mapCommitMessage(payload model.CommitMessage) ([]byte, error) {
	jsonDto := commitMessagePayloadDto{
		CorrelationId: string(payload.CorrelationId),
		MessageId:     int64(payload.MessageId),
	}

	return json.Marshal(jsonDto)
}

func fromMessageValueBytes(bytes []byte, messageType model.OutboxMessageType) (any, error) {
	switch messageType {
	case model.OutboxMessageTypeRollbackMessage:
		payloadDto := rollbackMessagePayloadDto{}
		err := json.Unmarshal(bytes, &payloadDto)

		if err != nil {
			return nil, err
		}

		message := model.RollbackMessage{
			CorrelationId: model.CorrelationId(payloadDto.CorrelationId),
			MessageId:     model.MessageId(payloadDto.MessageId),
		}

		return message, nil
	case model.OutboxMessageTypeCommitMessage:
		payloadDto := commitMessagePayloadDto{}
		err := json.Unmarshal(bytes, &payloadDto)

		if err != nil {
			return nil, err
		}

		message := model.CommitMessage{
			CorrelationId: model.CorrelationId(payloadDto.CorrelationId),
			MessageId:     model.MessageId(payloadDto.MessageId),
		}

		return message, nil
	default:
		err := fmt.Errorf("unsupported payloadDto type")
		return nil, err
	}
}

type outboxMessageDto struct {
	Id           int64  `db:"id"`
	MessageType  int    `db:"type"`
	MessageKey   string `db:"message_key"`
	MessageValue string `db:"message_value"`
	IsSent       bool   `db:"is_sent"`
}

type rollbackMessagePayloadDto struct {
	CorrelationId string `json:"correlationId"`
	MessageId     int64  `json:"messageId"`
}

type commitMessagePayloadDto struct {
	CorrelationId string `json:"correlationId"`
	MessageId     int64  `json:"messageId"`
}
