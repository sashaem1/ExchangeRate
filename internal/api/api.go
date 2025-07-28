package api

type ExchangeRates interface {
	GetLatestExchangeRatesByBase(base, rates string) (RateResponse, error)
	GetLatestExchangeRatesByDate(date string) ([]RateResponse, error)
	ValidateDate(dateStr string) error
	ValidateRate(rateStr string) error
	GetDefaultBase() map[string][]string
}

type RateResponse struct {
	Base  string
	Rates map[string]float64 `json:"data"`
}
