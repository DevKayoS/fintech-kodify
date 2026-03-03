package utils

import (
	"fmt"
	"time"
)

// ParseMonthRange parseia uma string no formato "YYYY-MM" e retorna
// o início e o fim do mês em UTC.
func ParseMonthRange(month string) (time.Time, time.Time, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("formato de mês inválido, use YYYY-MM")
	}

	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	return start, end, nil
}

// CurrentMonthRange retorna o início e o fim do mês atual em UTC.
func CurrentMonthRange() (time.Time, time.Time) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end
}

// ToReais converte centavos (int64) para reais (float64).
func ToReais(centavos int64) float64 {
	return float64(centavos) / 100.0
}

// ToCentavos converte reais (float64) para centavos (int64).
func ToCentavos(reais float64) int64 {
	return int64(reais * 100)
}
