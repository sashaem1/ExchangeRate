package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type ExchangeID string

type Exchange struct {
	ID             ExchangeID
	BaseCurrency   Currency
	TargetCurrency Currency
	Rate           float64
	Timestamp      time.Time
}

const dataFormat string = "2006-01-02"
const cronUpdateTime string = "00 12 * * *"

var initDates []time.Time = []time.Time{
	time.Date(2025, time.July, 21, 0, 0, 0, 0, time.UTC),
	time.Date(2025, time.July, 22, 0, 0, 0, 0, time.UTC),
	time.Date(2025, time.July, 23, 0, 0, 0, 0, time.UTC),
	time.Date(2025, time.July, 24, 0, 0, 0, 0, time.UTC),
	time.Date(2025, time.July, 25, 0, 0, 0, 0, time.UTC),
	time.Now(),
}

func NewExchange(baseCurrency, targetCurrency Currency, rate float64, timestamp time.Time) (Exchange, error) {
	op := "internal.Exchange.NewExchange"

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
	op := "internal.Exchange.GetByBase"
	ctx := context.Background()

	exchange, err := rr.storage.Get(ctx, baseCurrencyCode, targetCurrencyCode, time.Now())
	if err != nil {
		return Exchange{}, fmt.Errorf("%s: %s", op, err)
	}

	if exchange.Timestamp.IsZero() {
		exchange, err = rr.externalAPI.GetByBase(baseCurrencyCode, targetCurrencyCode)
		if err != nil {

			return Exchange{}, fmt.Errorf("%s: %s", op, err)
		}

		err = rr.storage.Set(ctx, exchange)
		if err != nil {
			return Exchange{}, fmt.Errorf("%s: %s", op, err)
		}
	}

	return exchange, nil
}

func (rr *ExchangeRepository) GetByDate(date string) ([]Exchange, error) {
	op := "internal.Exchange.GetByDate"
	ctx := context.Background()

	exchanges := []Exchange{}
	parsedDate, err := time.Parse(dataFormat, date)
	if err != nil {
		return exchanges, fmt.Errorf("%s: %s", op, err)
	}

	exchanges, missingExchange, err := rr.getByDateFromDb(ctx, parsedDate)
	if err != nil {
		return exchanges, fmt.Errorf("%s: %s", op, err)
	}

	if len(missingExchange) == 0 {
		return exchanges, nil
	} else {
		exchanges, err = rr.getByDateFromExAPI(ctx, parsedDate)
		if err != nil {
			return exchanges, fmt.Errorf("%s: %s", op, err)
		}

		err = rr.setByMisToDb(ctx, parsedDate, missingExchange, exchanges)
		if err != nil {
			return exchanges, fmt.Errorf("%s: %s", op, err)
		}

	}

	return exchanges, nil
}

func (rr *ExchangeRepository) getByDateFromDb(ctx context.Context, date time.Time) (exchanges []Exchange, missingExchange map[string][]string, err error) {
	op := "internal.Exchange.GetByDateFromDb"
	exchanges = []Exchange{}
	missingExchange = make(map[string][]string)

	for baseCurrencyCode, targetCurrencyCodes := range defaultBase {
		for _, tcc := range targetCurrencyCodes {
			currentExchange, err := rr.storage.Get(ctx, baseCurrencyCode, tcc, date)
			if err != nil {
				return exchanges, missingExchange, fmt.Errorf("%s: %s", op, err)
			}

			if currentExchange.Timestamp.IsZero() {
				missingExchange[baseCurrencyCode] = append(missingExchange[baseCurrencyCode], tcc)
			} else {
				exchanges = append(exchanges, currentExchange)
			}

		}
	}

	return exchanges, missingExchange, nil
}

func (rr *ExchangeRepository) getByDateFromExAPI(ctx context.Context, date time.Time) (exchanges []Exchange, err error) {
	op := "internal.Exchange.GetByDateFromExAPI"
	exchanges = []Exchange{}

	for baseCurrencyCode, targetCurrencyCodes := range defaultBase {

		currentExchange, err := rr.externalAPI.GetByDate(
			baseCurrencyCode, targetCurrencyCodes, date)

		if err != nil {
			return exchanges, fmt.Errorf("%s: %s", op, err)
		}

		exchanges = append(exchanges, currentExchange...)
	}

	return
}

func (rr *ExchangeRepository) setByMisToDb(ctx context.Context, date time.Time, missingExchange map[string][]string, exchanges []Exchange) error {
	op := "internal.Exchange.setByMisToDb"

	for _, exchange := range exchanges {
		misTarCurrencyCodes, ok := missingExchange[exchange.BaseCurrency.Code]
		if ok {
			for _, mtcc := range misTarCurrencyCodes {
				if exchange.TargetCurrency.Code == mtcc {
					err := rr.storage.Set(ctx, exchange)
					if err != nil {
						return fmt.Errorf("%s: %s", op, err)
					}
				}
			}
		}
	}

	return nil
}

func (rr *ExchangeRepository) InitExchangeRepository(ctx context.Context) error {
	op := "internal.Exchange.InitExchangeRepository"

	err := rr.initData(ctx, initDates)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	err = rr.сronUpdateData(ctx)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}

func (rr *ExchangeRepository) initData(ctx context.Context, initDates []time.Time) error {
	op := "internal.Exchange.initData"

	for _, date := range initDates {
		exchanges, err := rr.getByDateFromExAPI(ctx, date)
		if err != nil {
			return fmt.Errorf("%s: %s", op, err)
		}

		err = rr.setByMisToDb(ctx, date, defaultBase, exchanges)
		if err != nil {
			return fmt.Errorf("%s: %s", op, err)
		}
	}

	return nil
}

func (rr *ExchangeRepository) сronUpdateData(ctx context.Context) error {
	op := "internal.Exchange.InitExchangeRepository"
	cron := cron.New()

	_, err := cron.AddFunc(cronUpdateTime, func() {
		log.Printf("!Проверка крона! Время:", time.Now().Format(time.RFC3339))
		initDates := []time.Time{
			time.Now(),
		}

		err := rr.initData(ctx, initDates)
		if err != nil {
			log.Printf("%s: %s", op, err)
		}

	})

	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	cron.Start()

	return nil
}
