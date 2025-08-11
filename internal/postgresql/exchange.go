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

func (es *ExchangeStorage) Get(ctx context.Context, baseCurrencyCode, targetCurrencyCode string, date time.Time) (internal.Exchange, error) {
	op := "postgresql.Get"

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

func (es *ExchangeStorage) Set(ctx context.Context, exchange internal.Exchange) error {
	op := "postgresql.Set"

	query := `INSERT INTO exchange_rates (BaseCurrency, TargetCurrency, rate, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := es.pgPool.Exec(ctx, query, exchange.BaseCurrency.Code, exchange.TargetCurrency.Code, exchange.Rate, exchange.Timestamp)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}

func (es *ExchangeStorage) VerificationAPIKey(ctx context.Context, APIKey string) (bool, error) {
	// panic("implement me")
	return true, nil
}
