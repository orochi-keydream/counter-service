package api

import (
	"context"
	"github.com/orochi-keydream/counter-service/internal/proto/admin"
	"github.com/orochi-keydream/counter-service/internal/service"
)

type AdminService struct {
	admin.UnimplementedCounterServiceServer

	realtimeConfigService *service.RealtimeConfigService
}

func NewAdminGrpcService(realtimeConfigService *service.RealtimeConfigService) *AdminService {
	return &AdminService{
		realtimeConfigService: realtimeConfigService,
	}
}

func (s *AdminService) SetRollbackAllMessagesModeV1(
	ctx context.Context,
	req *admin.SetRollbackAllMessagesModeV1Request,
) (*admin.SetRollbackAllMessagesModeV1Response, error) {
	s.realtimeConfigService.SetIsAlwaysRollbackMessagesEnabled(req.IsEnabled)

	return &admin.SetRollbackAllMessagesModeV1Response{}, nil
}
