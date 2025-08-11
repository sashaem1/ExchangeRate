package internal

import (
	"context"
	"fmt"
	"log"
	"time"
)

type ExchangeID string

type Exchange struct {
	ID             ExchangeID
	BaseCurrency   Currency
	TargetCurrency Currency
	Rate           float64
	Timestamp      time.Time
}

func NewExchange(baseCurrency, targetCurrency Currency, rate float64, timestamp time.Time) (Exchange, error) {
	op := "internal.NewExchange"

	if rate <= 0.0 {
		return Exchange{}, fmt.Errorf("%s: Значение курса должно быть положительным", op)
	}

	newExchange := Exchange{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
		Rate:           rate,
		Timestamp:      timestamp,
	}
	return newExchange, nil
}

type ExchangeStorage interface {
	Get(ctx context.Context, baseCurrencyCode, targetCurrencyCode string, date time.Time) (Exchange, error)
	Set(ctx context.Context, exchange Exchange) error
	VerificationAPIKey(ctx context.Context, APIKey string) (bool, error)
}

type ExchangeExternalAPI interface {
	GetByBase(baseCurrencyCode, targetCurrencyCode string) (Exchange, error)
	GetByDate(baseCurrencyCode string, targetCurrencyCode []string, date time.Time) ([]Exchange, error)
}

type ExchangeRepository struct {
	storage     ExchangeStorage
	externalAPI ExchangeExternalAPI
}

func NewExchangeRepository(storage ExchangeStorage, externalAPI ExchangeExternalAPI) *ExchangeRepository {
	return &ExchangeRepository{
		storage:     storage,
		externalAPI: externalAPI,
	}
}

func (rr *ExchangeRepository) GetByBase(baseCurrencyCode, targetCurrencyCode string) (Exchange, error) {
	op := "Exchange.GetByBase"
	ctx := context.Background()

	exchange, err := rr.storage.Get(ctx, baseCurrencyCode, targetCurrencyCode, time.Now())
	if err != nil {
		return Exchange{}, fmt.Errorf("%s:%s", op, err)
	}

	if exchange.Timestamp.IsZero() {
		exchange, err = rr.externalAPI.GetByBase(baseCurrencyCode, targetCurrencyCode)
		if err != nil {

			return Exchange{}, fmt.Errorf("%s:%s", op, err)
		}

		err = rr.storage.Set(ctx, exchange)
		if err != nil {
			return Exchange{}, fmt.Errorf("%s:%s", op, err)
		}
	}

	return exchange, nil
}

func (rr *ExchangeRepository) GetByDate(date string) ([]Exchange, error) {
	op := "Exchange.GetByDate"
	ctx := context.Background()

	exchanges := []Exchange{}
	missingExchange := make(map[string][]string)
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return exchanges, fmt.Errorf("%s: %s", op, err)
	}

	for baseCurrencyCode, targetCurrencyCodes := range defaultBase {
		for _, tcc := range targetCurrencyCodes {
			currentExchange, err := rr.storage.Get(ctx, baseCurrencyCode, tcc, parsedDate)
			if err != nil {
				return exchanges, fmt.Errorf("%s: %s", op, err)
			}

			if currentExchange.Timestamp.IsZero() {
				missingExchange[baseCurrencyCode] = append(missingExchange[baseCurrencyCode], tcc)
			} else {
				exchanges = append(exchanges, currentExchange)
			}

		}
	}

	if len(missingExchange) == 0 {
		return exchanges, nil
	} else {
		exchanges = exchanges[:0]
		for baseCurrencyCode, targetCurrencyCodes := range defaultBase {

			currentExchange, err := rr.externalAPI.GetByDate(
				baseCurrencyCode, targetCurrencyCodes, parsedDate)

			if err != nil {
				return exchanges, fmt.Errorf("%s: %s", op, err)
			}

			log.Printf("currentExchange from external API: ", currentExchange)
			exchanges = append(exchanges, currentExchange...)
		}

		for _, exchange := range exchanges {
			misTarCurrencyCodes, ok := missingExchange[exchange.BaseCurrency.Code]
			if ok {
				for _, mtcc := range misTarCurrencyCodes {
					if exchange.TargetCurrency.Code == mtcc {
						rr.storage.Set(ctx, exchange)
					}
				}
			}
		}
	}

	return exchanges, nil
}

func (rr *ExchangeRepository) VerificationAPIKey(APIKey string) (bool, error) {
	op := "internal.VerificationAPIKey"
	ctx := context.Background()

	verificationStatus, err := rr.storage.VerificationAPIKey(ctx, APIKey)
	if err != nil {
		return false, fmt.Errorf("%s: %s", op, err)
	}

	return verificationStatus, nil
}
