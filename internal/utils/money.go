package utils

import "time"

func CurrentMonthRange() (time.Time, time.Time) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end
}

func ToReais(centavos int64) float64 {
	return float64(centavos) / 100.0
}

func ToCentavos(reais float64) int64 {
	return int64(reais * 100)
}
