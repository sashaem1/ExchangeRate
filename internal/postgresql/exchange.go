package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaem1/ExchangeRate/internal"
)

type ExchangeStorage struct {
	pgPool *pgxpool.Pool
}

func NewExchangeStorage(pgPool *pgxpool.Pool) *ExchangeStorage {
	return &ExchangeStorage{pgPool: pgPool}
}

func (es *ExchangeStorage) GetExchange(ctx context.Context, baseCurrencyCode, targetCurrencyCode string, date time.Time) (internal.Exchange, error) {
	op := "postgresql.GetExchange"

	query := `SELECT rate, updated_at 
              FROM exchange_rates 
              WHERE baseCurrency = $1 AND targetCurrency = $2 AND DATE(updated_at) = DATE($3)`

	var exchange internal.Exchange

	err := es.pgPool.QueryRow(ctx, query, baseCurrencyCode, targetCurrencyCode, date).Scan(
		&exchange.Rate,
		&exchange.Timestamp,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return internal.Exchange{}, nil
		}
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}

	baseCurrency, err := internal.NewCurrency(baseCurrencyCode)
	if err != nil {
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}
	exchange.BaseCurrency = baseCurrency

	targetCurrency, err := internal.NewCurrency(targetCurrencyCode)
	if err != nil {
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}
	exchange.TargetCurrency = targetCurrency

	return exchange, nil
}

func (es *ExchangeStorage) SetExchange(ctx context.Context, exchange internal.Exchange) error {
	op := "postgresql.SetExchange"

	query := `INSERT INTO exchange_rates (BaseCurrency, TargetCurrency, rate, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := es.pgPool.Exec(ctx, query, exchange.BaseCurrency.Code, exchange.TargetCurrency.Code, exchange.Rate, exchange.Timestamp)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}

func (es *ExchangeStorage) GetAPIKey(ctx context.Context, APIKey string) (string, error) {
	op := "postgresql.GetExchange"

	query := `SELECT key
              FROM api_keys 
              WHERE key = $1`

	var result string

	err := es.pgPool.QueryRow(ctx, query, APIKey).Scan(
		&result,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return result, fmt.Errorf("%s: %s", op, err)
	}

	return result, nil
}

func (es *ExchangeStorage) SetAPIKey(ctx context.Context, APIKey string) error {
	op := "postgresql.SetAPIKey"

	query := `INSERT INTO api_keys (key) VALUES ($1) ON CONFLICT (key) DO NOTHING`
	_, err := es.pgPool.Exec(ctx, query, APIKey)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
