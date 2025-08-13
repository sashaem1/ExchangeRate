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
	op := "internal.currency.NewCurrency"
	code = strings.ToUpper(strings.TrimSpace(code))

	if len(code) != 3 {
		err := fmt.Sprintf("Название действия должно состоять из 4 символов")
		return Currency{}, fmt.Errorf("%s: %s", op, err)
	}

	if _, ok := defaultBase[code]; !ok {
		err := fmt.Sprintf("Отсутствует такая валюта в системе: %s", code)
		return Currency{}, fmt.Errorf("%s: %s", op, err)
	}

	return Currency{Code: code}, nil
}
