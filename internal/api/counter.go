package api

import (
	"context"
	"github.com/orochi-keydream/counter-service/internal/model"
	"github.com/orochi-keydream/counter-service/internal/proto/counter"
	"github.com/orochi-keydream/counter-service/internal/service"
)

type CounterService struct {
	counter.UnimplementedCounterServiceServer

	counterService *service.CounterService
}

func NewCounterGrpcService(counterService *service.CounterService) *CounterService {
	return &CounterService{
		counterService: counterService,
	}
}

func (s *CounterService) GetUnreadCountTotalV1(
	ctx context.Context,
	req *counter.GetUnreadCountTotalV1Request,
) (*counter.GetUnreadCountTotalV1Response, error) {
	cmd := model.GetUnreadCountTotalCommand{
		UserId: model.UserId(req.UserId),
	}

	count, err := s.counterService.GetUnreadCountTotal(ctx, cmd)

	if err != nil {
		return nil, err
	}

	resp := &counter.GetUnreadCountTotalV1Response{
		Count: int32(count),
	}

	return resp, nil
}

func (s *CounterService) GetUnreadCountV1(
	ctx context.Context,
	req *counter.GetUnreadCountV1Request,
) (*counter.GetUnreadCountV1Response, error) {
	cmd := model.GetUnreadCountCommand{
		CurrentUserId: model.UserId(req.CurrentUserId),
		ChatUserId:    model.UserId(req.ChatUserId),
	}

	count, err := s.counterService.GetUnreadCount(ctx, cmd)

	if err != nil {
		return nil, err
	}

	resp := &counter.GetUnreadCountV1Response{
		Count: int32(count),
	}

	return resp, nil
}

func (s *CounterService) MarkMessagesAsReadV1(
	ctx context.Context,
	req *counter.MarkMessagesAsReadV1Request,
) (*counter.MarkMessagesAsReadV1Response, error) {
	messageIds := make([]model.MessageId, len(req.MessageIds))

	for i, id := range req.MessageIds {
		messageIds[i] = model.MessageId(id)
	}

	cmd := model.MarkMessagesAsReadCommand{
		UserId:     model.UserId(req.UserId),
		MessageIds: messageIds,
	}

	err := s.counterService.MarkMessagesAsRead(ctx, cmd)

	if err != nil {
		return nil, err
	}

	return &counter.MarkMessagesAsReadV1Response{}, nil
}
