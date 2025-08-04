package internal

import (
	"fmt"
	"time"
)

type Rate struct {
	BaseCurrency   Currency
	TargetCurrency Currency
	Rate           float64
	Timestamp      time.Time
}

func NewRate(baseCurrency, targetCurrency Currency, rate float64, timestamp time.Time) (Rate, error) {
	if rate <= 0.0 {
		return Rate{}, fmt.Errorf("Значение курса должно быть положительным")
	}

	newRate := Rate{
		BaseCurrency:   baseCurrency,
		TargetCurrency: targetCurrency,
		Rate:           rate,
		Timestamp:      timestamp,
	}
	return newRate, nil
}
