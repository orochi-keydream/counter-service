package repository

import (
	"context"
	"database/sql"
	"github.com/orochi-keydream/counter-service/internal/model"
)

type CommandRepository struct {
	db *sql.DB
}

func NewCommandRepository(db *sql.DB) *CommandRepository {
	return &CommandRepository{
		db: db,
	}
}

func (cr *CommandRepository) Add(ctx context.Context, correlationId model.CorrelationId, tx *sql.Tx) error {
	const query = "insert into handled_commands (correlation_id) values ($1)"

	var ec IExecutionContext

	if tx == nil {
		ec = cr.db
	} else {
		ec = tx
	}

	_, err := ec.ExecContext(ctx, query, correlationId)

	return err
}

func (cr *CommandRepository) Exists(ctx context.Context, correlationId model.CorrelationId, tx *sql.Tx) (bool, error) {
	const query = "select true from handled_commands where correlation_id = $1"

	var ec IExecutionContext

	if tx == nil {
		ec = cr.db
	} else {
		ec = tx
	}

	row := ec.QueryRowContext(ctx, query, correlationId)

	err := row.Err()

	if err != nil {
		return false, err
	}

	var exists bool

	err = row.Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}
