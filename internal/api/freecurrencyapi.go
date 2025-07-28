// https://app.freecurrencyapi.com/dashboard
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type FreeCurrencyApi struct {
	apiKey string
}

func NewFreeCurrencyApi() *FreeCurrencyApi {
	newFRA := &FreeCurrencyApi{}
	apiKey := os.Getenv("FREECURRENCY_API_KEY")

	if apiKey == "" {
		log.Fatal("FreeCurrency_API_KEY не найден")
	}

	newFRA.apiKey = apiKey
	return newFRA
}

var defaultBase = map[string][]string{
	"USD": {"RUB", "EUR", "JPY"},
	"RUB": {"USD", "EUR", "JPY"},
	"EUR": {"RUB", "USD", "JPY"},
	"JPY": {"RUB", "EUR", "USD"},
}

func (f *FreeCurrencyApi) GetDefaultBase() map[string][]string {
	return defaultBase
}

func (f *FreeCurrencyApi) GetLatestExchangeRatesByBase(base, currency string) (RateResponse, error) {
	url := "https://api.freecurrencyapi.com/v1/latest"

	requestUrl := fmt.Sprintf("%s?&apikey=%s&base_currency=%s&currencies=%s", url, f.apiKey, base, currency)

	resp, err := http.Get(requestUrl)
	if err != nil {
		return RateResponse{}, fmt.Errorf("Не удалось получить данные: %s", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp RateResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return RateResponse{}, fmt.Errorf("Не удалось преобразовать данные: %s", err)
	}

	apiResp.Base = base

	return apiResp, nil
}

func (f *FreeCurrencyApi) GetLatestExchangeRatesByDate(date string) ([]RateResponse, error) {
	url := "https://api.freecurrencyapi.com/v1/latest"
	result := make([]RateResponse, 0, 4)

	for base, currencies := range defaultBase {
		requestUrl := fmt.Sprintf("%s?&apikey=%s&date=%s&base_currency=%s&currencies=%s", url, f.apiKey, date, base, strings.Join(currencies, ","))

		resp, err := http.Get(requestUrl)
		if err != nil {
			panic(err)
		}

		body, _ := io.ReadAll(resp.Body)

		var apiResp RateResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("Не удалось преобразовать данные: %w", err)
		}

		apiResp.Base = base

		result = append(result, apiResp)
		resp.Body.Close()
	}

	return result, nil
}

func (f *FreeCurrencyApi) ValidateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	return err
}

func (f *FreeCurrencyApi) ValidateRate(rateStr string) error {
	if _, ok := defaultBase[rateStr]; !ok {
		return fmt.Errorf("Отсутствует такая валюта в системе: %s", rateStr)
	}
	return nil
}
