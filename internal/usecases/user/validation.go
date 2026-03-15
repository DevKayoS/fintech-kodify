package user

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateCPF valida um CPF brasileiro (com ou sem pontuação).
func ValidateCPF(cpf string) bool {
	digits := onlyDigits(cpf)

	if len(digits) != 11 {
		return false
	}

	// CPFs com todos os dígitos iguais são inválidos (111.111.111-11, etc.)
	allSame := true
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Valida primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	remainder := (sum * 10) % 11
	if remainder == 10 || remainder == 11 {
		remainder = 0
	}
	if remainder != int(digits[9]-'0') {
		return false
	}

	// Valida segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	remainder = (sum * 10) % 11
	if remainder == 10 || remainder == 11 {
		remainder = 0
	}
	return remainder == int(digits[10]-'0')
}

// ValidatePhone valida um número de telefone brasileiro (10 ou 11 dígitos com DDD).
func ValidatePhone(phone string) bool {
	digits := onlyDigits(phone)
	return len(digits) == 10 || len(digits) == 11
}

// ValidateEmail valida um endereço de e-mail.
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
