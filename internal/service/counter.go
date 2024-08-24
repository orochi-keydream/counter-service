package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/orochi-keydream/counter-service/internal/model"
	"log/slog"
)

type IMessageRepository interface {
	CountUnreadByUser(ctx context.Context, userId model.UserId, tx *sql.Tx) (int, error)
	CountUnreadByUserAndChat(ctx context.Context, userId model.UserId, chatId model.ChatId, tx *sql.Tx) (int, error)
	GetForUser(ctx context.Context, userId model.UserId, messageIds []model.MessageId, tx *sql.Tx) ([]*model.Message, error)
	Update(ctx context.Context, messages []*model.Message, tx *sql.Tx) error
	Add(ctx context.Context, message *model.Message, tx *sql.Tx) error
}

type ICommandRepository interface {
	Add(ctx context.Context, correlationId model.CorrelationId, tx *sql.Tx) error
	Exists(ctx context.Context, correlationId model.CorrelationId, tx *sql.Tx) (bool, error)
}

type IRealtimeConfigService interface {
	IsAlwaysRollbackMessagesEnabled() bool
}

type CounterService struct {
	counterRepository     IMessageRepository
	commandRepository     ICommandRepository
	outboxRepository      IOutboxRepository
	transactionManager    ITransactionManager
	realtimeConfigService IRealtimeConfigService
}

func NewCounterService(
	messageRepository IMessageRepository,
	commandRepository ICommandRepository,
	outboxRepository IOutboxRepository,
	transactionManager ITransactionManager,
	realtimeConfigService IRealtimeConfigService,
) *CounterService {
	return &CounterService{
		counterRepository:     messageRepository,
		commandRepository:     commandRepository,
		outboxRepository:      outboxRepository,
		transactionManager:    transactionManager,
		realtimeConfigService: realtimeConfigService,
	}
}

func (s *CounterService) GetUnreadCountTotal(ctx context.Context, cmd model.GetUnreadCountTotalCommand) (int, error) {
	return s.counterRepository.CountUnreadByUser(ctx, cmd.UserId, nil)
}

func (s *CounterService) GetUnreadCount(ctx context.Context, cmd model.GetUnreadCountCommand) (int, error) {
	chatId := s.buildChatId(cmd.CurrentUserId, cmd.ChatUserId)
	return s.counterRepository.CountUnreadByUserAndChat(ctx, cmd.CurrentUserId, chatId, nil)
}

func (s *CounterService) MarkMessagesAsRead(ctx context.Context, cmd model.MarkMessagesAsReadCommand) error {
	messages, err := s.counterRepository.GetForUser(ctx, cmd.UserId, cmd.MessageIds, nil)

	if err != nil {
		return err
	}

	for _, message := range messages {
		message.IsRead = true
	}

	// TODO: Consider using transaction.

	err = s.counterRepository.Update(ctx, messages, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *CounterService) AddNewUnreadMessage(ctx context.Context, cmd model.AddNewUnreadMessageCommand) error {
	exists, err := s.commandRepository.Exists(ctx, cmd.CorrelationId, nil)

	if err != nil {
		return err
	}

	if exists {
		slog.Info(fmt.Sprintf("Command with correlation ID %v was handled before", cmd.CorrelationId))
		return nil
	}

	message := &model.Message{
		MessageId: cmd.MessageId,
		UserId:    cmd.UserId,
		ChatId:    cmd.ChatId,
		IsRead:    false,
	}

	tx, err := s.transactionManager.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if s.realtimeConfigService.IsAlwaysRollbackMessagesEnabled() {
		err = fmt.Errorf("custom error occured")
	} else {
		err = s.counterRepository.Add(ctx, message, tx)
	}

	if err != nil {
		payload := model.RollbackMessage{
			CorrelationId: cmd.CorrelationId,
			MessageId:     cmd.MessageId,
		}

		outboxMessage := &model.OutboxMessage{
			Type:         model.OutboxMessageTypeRollbackMessage,
			MessageKey:   cmd.ChatId,
			MessageValue: payload,
			IsSent:       false,
		}

		err = s.commandRepository.Add(ctx, cmd.CorrelationId, tx)

		if err != nil {
			return err
		}

		err = s.outboxRepository.Add(ctx, outboxMessage, tx)

		if err != nil {
			return err
		}
	} else {
		payload := model.CommitMessage{
			CorrelationId: cmd.CorrelationId,
			MessageId:     cmd.MessageId,
		}

		outboxMessage := &model.OutboxMessage{
			Type:         model.OutboxMessageTypeCommitMessage,
			MessageKey:   cmd.ChatId,
			MessageValue: payload,
			IsSent:       false,
		}

		err = s.outboxRepository.Add(ctx, outboxMessage, tx)

		if err != nil {
			return err
		}

		err = s.commandRepository.Add(ctx, cmd.CorrelationId, tx)

		if err != nil {
			return err
		}
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	slog.Info("Message is successfully handled")

	return nil
}

func (s *CounterService) buildChatId(firstUser, secondUser model.UserId) model.ChatId {
	if firstUser > secondUser {
		return model.ChatId(fmt.Sprintf("%s_%s", secondUser, firstUser))
	} else {
		return model.ChatId(fmt.Sprintf("%s_%s", firstUser, secondUser))
	}
}
