package jobs

import (
	"context"
	"github.com/orochi-keydream/counter-service/internal/service"
	"log/slog"
	"sync"
	"time"
)

type OutboxJob struct {
	outboxService *service.OutboxService
}

func NewOutboxJob(outboxService *service.OutboxService) *OutboxJob {
	return &OutboxJob{outboxService}
}

func (oj *OutboxJob) Start(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				oj.process(ctx)
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

func (oj *OutboxJob) process(ctx context.Context) {
	err := oj.outboxService.Send(ctx)

	if err != nil {
		slog.Error(err.Error())
	}
}
