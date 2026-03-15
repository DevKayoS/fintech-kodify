package utils

import (
	"fmt"
	"time"
)

// ParseDateTime parseia uma string de data em RFC3339 ou YYYY-MM-DD.
// Retorna time.Now().UTC() se a string for vazia.
func ParseDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("formato de data inválido, use RFC3339 ou YYYY-MM-DD")
}

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
