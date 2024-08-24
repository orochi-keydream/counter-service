package service

import (
	"context"
	"fmt"
	"github.com/orochi-keydream/counter-service/internal/model"
	"log/slog"
)

type OutboxService struct {
	outboxRepository        IOutboxRepository
	dialogueCommandProducer DialogueCommandProducer
}

type DialogueCommandProducer interface {
	SendMessage(message *model.OutboxMessage) error
}

func NewOutboxService(
	outboxRepository IOutboxRepository,
	outboxProducer DialogueCommandProducer,
) *OutboxService {
	return &OutboxService{
		outboxRepository:        outboxRepository,
		dialogueCommandProducer: outboxProducer,
	}
}

func (s *OutboxService) Send(ctx context.Context) error {
	messages, err := s.outboxRepository.GetUnsent(ctx, nil)

	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	for _, message := range messages {
		err = s.dialogueCommandProducer.SendMessage(message)

		if err != nil {
			return err
		}

		message.IsSent = true
	}

	err = s.outboxRepository.Update(ctx, messages, nil)

	if err != nil {
		return err
	}

	messageIds := make([]int64, len(messages))

	for i, message := range messages {
		messageIds[i] = message.Id
	}

	slog.Info(fmt.Sprintf("Sent %d messages from outbox: %v", len(messages), messageIds))

	return nil
}
