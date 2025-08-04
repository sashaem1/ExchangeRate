package internal

import (
	"fmt"
	"strings"
)

var defaultBase = map[string][]string{
	"USD": {"RUB", "EUR", "JPY"},
	"RUB": {"USD", "EUR", "JPY"},
	"EUR": {"RUB", "USD", "JPY"},
	"JPY": {"RUB", "EUR", "USD"},
}

type Currency struct {
	Code string
}

func NewCurrency(code string) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	if len(code) != 3 {
		return Currency{}, fmt.Errorf("код валюты должен состоять из 3 символов")
	}

	if _, ok := defaultBase[code]; !ok {
		return Currency{}, fmt.Errorf("Отсутствует такая валюта в системе: %s", code)
	}

	return Currency{Code: code}, nil
}
