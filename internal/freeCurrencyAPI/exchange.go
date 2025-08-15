package freecurrencyapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashaem1/ExchangeRate/internal"
)

var baseURL string = "https://api.freecurrencyapi.com/v1/latest"

type ExchangeExternalAPI struct {
	APIKey string
}

const baseTimeFormate string = "2006-01-02"

type RateResponse struct {
	Base  string
	Rates map[string]float64 `json:"data"`
}

func NewExchangeExternalAPI(APIKey string) *ExchangeExternalAPI {
	return &ExchangeExternalAPI{
		APIKey: APIKey,
	}
}

func (fc *ExchangeExternalAPI) GetByBase(baseCurrencyCode, targetCurrencyCode string) (internal.Exchange, error) {
	op := "FreeCurrencyAPI.exchange.GetByBase"

	requestUrl := fmt.Sprintf("%s?&apikey=%s&base_currency=%s&currencies=%s", baseURL, fc.APIKey, baseCurrencyCode, targetCurrencyCode)

	resp, err := http.Get(requestUrl)
	if err != nil {
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp RateResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}

	if apiResp.Rates[targetCurrencyCode] == 0 {
		err := fmt.Sprintf("Не удалось получить данные со стороннего апи")
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}

	exchange, err := internal.NewExchange(baseCurrencyCode, targetCurrencyCode, apiResp.Rates[targetCurrencyCode], time.Now())
	if err != nil {
		return internal.Exchange{}, fmt.Errorf("%s: %s", op, err)
	}

	return exchange, nil
}

func (fc *ExchangeExternalAPI) GetByDate(baseCurrencyCode string, targetCurrencyCode []string, date time.Time) ([]internal.Exchange, error) {
	op := "FreeCurrencyAPI.exchange.GetByDate"
	parsedDate := date.Format(baseTimeFormate)
	result := make([]internal.Exchange, 0, 4)

	requestUrl := fmt.Sprintf("%s?&apikey=%s&date=%s&base_currency=%s&currencies=%s", baseURL, fc.APIKey, parsedDate, baseCurrencyCode, strings.Join(targetCurrencyCode, ","))

	resp, err := http.Get(requestUrl)
	if err != nil {
		return result, fmt.Errorf("%s: %s", op, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp RateResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return result, fmt.Errorf("%s: %s", op, err)
	}

	for tcc, rate := range apiResp.Rates {
		if rate == 0 {
			err := fmt.Sprintf("Не удалось получить данные со стороннего апи")
			return result, fmt.Errorf("%s: %s", op, err)
		}

		curExchange, err := internal.NewExchange(baseCurrencyCode, tcc, rate, date)
		if err != nil {
			return result, fmt.Errorf("%s: %s", op, err)
		}

		result = append(result, curExchange)
	}

	return result, nil
}
