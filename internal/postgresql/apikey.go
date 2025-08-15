package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaem1/ExchangeRate/internal"
)

type APIKeyStorage struct {
	pgPool *pgxpool.Pool
}

func NewAPIKeyStorage(pgPool *pgxpool.Pool) *APIKeyStorage {
	return &APIKeyStorage{pgPool: pgPool}
}

func (es *APIKeyStorage) Get(ctx context.Context, APIKey string) (internal.APIKey, error) {
	op := "postgresql.apikey.GetExchange"

	query := `SELECT key
              FROM api_keys 
              WHERE key = $1`

	result := internal.NewAPIKey(APIKey)

	err := es.pgPool.QueryRow(ctx, query, APIKey).Scan(
		&result.Key,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return result, fmt.Errorf("%s: %s", op, err)
	}

	result.Valid = true
	return result, nil
}

func (es *APIKeyStorage) Set(ctx context.Context, APIKey internal.APIKey) error {
	op := "postgresql.apikey.SetAPIKey"

	query := `INSERT INTO api_keys (key) VALUES ($1) ON CONFLICT (key) DO NOTHING`
	_, err := es.pgPool.Exec(ctx, query, APIKey.Key)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
