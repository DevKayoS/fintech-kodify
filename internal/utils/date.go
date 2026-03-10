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
