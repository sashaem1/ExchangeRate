package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaem1/ExchangeRate/internal"
)

type ActionLogStorage struct {
	pgPool *pgxpool.Pool
}

func NewActionLogStorage(pgPool *pgxpool.Pool) *ActionLogStorage {
	return &ActionLogStorage{pgPool: pgPool}
}

func (ls *ActionLogStorage) Set(ctx context.Context, ActionLog internal.ActionLog) error {
	op := "postgresql.logDb.Set"

	query := `INSERT INTO exchange_rates_log (action_name, updated_at) VALUES ($1, $2)`
	_, err := ls.pgPool.Exec(ctx, query, ActionLog.ActionLogType.Action, ActionLog.Timestamp)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
