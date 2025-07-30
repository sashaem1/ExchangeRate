package database

import "github.com/sashaem1/ExchangeRate/internal/api"

type Rate struct {
	Base     string
	Currency string
	Rate     float64
}

type DataBase interface {
	InitDB(api api.ExchangeRates)
	CloseConnect()
	InsertRate(base, currency string, rate float64, updatedAt string) error
	InsertLog(request_type string) error
	GetRatesByPair(base, currency string) (Rate, error)
	GetRatesByDate(bases map[string][]string, DateStr string) (rates []api.RateResponse, missingRates []Rate, err error)
	VerifyApiKey(ApiKey string) (bool, error)
	CronUpdateData(api api.ExchangeRates)
	FillMissingData(date string, missingRates []Rate, api api.ExchangeRates) ([]api.RateResponse, error)
}
