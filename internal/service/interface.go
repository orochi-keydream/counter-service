package service

import (
	"context"
	"database/sql"
	"github.com/orochi-keydream/counter-service/internal/model"
)

type IOutboxRepository interface {
	Add(ctx context.Context, message *model.OutboxMessage, tx *sql.Tx) error
	GetUnsent(ctx context.Context, tx *sql.Tx) ([]*model.OutboxMessage, error)
	Update(ctx context.Context, messages []*model.OutboxMessage, tx *sql.Tx) error
}

type ITransactionManager interface {
	Begin(ctx context.Context) (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error
}
